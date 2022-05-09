/*
Copyright 2019 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package repltracker

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/protobuf/proto"

	"vitess.io/vitess/go/vt/withddl"

	"context"

	"vitess.io/vitess/go/sqltypes"
	"vitess.io/vitess/go/timer"
	"vitess.io/vitess/go/vt/dbconnpool"
	"vitess.io/vitess/go/vt/log"
	"vitess.io/vitess/go/vt/logutil"
	"vitess.io/vitess/go/vt/mysqlctl"
	"vitess.io/vitess/go/vt/sqlparser"
	"vitess.io/vitess/go/vt/vttablet/tabletserver/tabletenv"

	querypb "vitess.io/vitess/go/vt/proto/query"
	topodatapb "vitess.io/vitess/go/vt/proto/topodata"
)

const (
	sqlCreateSidecarDB      = "create database if not exists %s"
	sqlCreateHeartbeatTable = `CREATE TABLE IF NOT EXISTS %s.heartbeat (
  keyspaceShard VARBINARY(256) NOT NULL PRIMARY KEY,
  tabletUid INT UNSIGNED NOT NULL,
  ts BIGINT UNSIGNED NOT NULL
        ) engine=InnoDB`
	sqlUpsertHeartbeat = "INSERT INTO %s.heartbeat (ts, tabletUid, keyspaceShard) VALUES (%a, %a, %a) ON DUPLICATE KEY UPDATE ts=VALUES(ts), tabletUid=VALUES(tabletUid)"
)

var withDDL = withddl.New([]string{
	fmt.Sprintf(sqlCreateSidecarDB, "_vt"),
	fmt.Sprintf(sqlCreateHeartbeatTable, "_vt"),
})

// heartbeatWriter runs on primary tablets and writes heartbeats to the _vt.heartbeat
// table at a regular interval, defined by heartbeat_interval.
type heartbeatWriter struct {
	env tabletenv.Env

	enabled       bool
	interval      time.Duration
	tabletAlias   *topodatapb.TabletAlias
	keyspaceShard string
	now           func() time.Time
	errorLog      *logutil.ThrottledLogger

	mu           sync.Mutex
	isOpen       bool
	appPool      *dbconnpool.ConnectionPool
	allPrivsPool *dbconnpool.ConnectionPool
	ticks        *timer.Timer

	onDemandDuration                time.Duration
	concurrentHeartbeatRequests     int64
	requestHeartbeatsReentranceFlag int64
}

// newHeartbeatWriter creates a new heartbeatWriter.
func newHeartbeatWriter(env tabletenv.Env, alias *topodatapb.TabletAlias) *heartbeatWriter {
	config := env.Config()

	// config.EnableLagThrottler is a feature flag for the throttler; if throttler runs, then heartbeat must also run
	if config.ReplicationTracker.Mode != tabletenv.Heartbeat && !config.EnableLagThrottler {
		return &heartbeatWriter{}
	}
	heartbeatInterval := config.ReplicationTracker.HeartbeatIntervalSeconds.Get()
	return &heartbeatWriter{
		env:              env,
		enabled:          true,
		tabletAlias:      proto.Clone(alias).(*topodatapb.TabletAlias),
		now:              time.Now,
		interval:         heartbeatInterval,
		onDemandDuration: config.ReplicationTracker.HeartbeatOnDemandSeconds.Get(),
		ticks:            timer.NewTimer(heartbeatInterval),
		errorLog:         logutil.NewThrottledLogger("HeartbeatWriter", 60*time.Second),
		// We make this pool size 2; to prevent pool exhausted
		// stats from incrementing continually, and causing concern
		appPool:      dbconnpool.NewConnectionPool("HeartbeatWriteAppPool", 2, *mysqlctl.DbaIdleTimeout, *mysqlctl.PoolDynamicHostnameResolution),
		allPrivsPool: dbconnpool.NewConnectionPool("HeartbeatWriteAllPrivsPool", 2, *mysqlctl.DbaIdleTimeout, *mysqlctl.PoolDynamicHostnameResolution),
	}
}

// InitDBConfig initializes the target name for the heartbeatWriter.
func (w *heartbeatWriter) InitDBConfig(target *querypb.Target) {
	w.keyspaceShard = fmt.Sprintf("%s:%s", target.Keyspace, target.Shard)
}

// Open sets up the heartbeatWriter's db connection and launches the ticker
// responsible for periodically writing to the heartbeat table.
func (w *heartbeatWriter) Open() {
	if !w.enabled {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.isOpen {
		return
	}
	log.Info("Hearbeat Writer: opening")

	// We cannot create the database and tables in this Open function
	// since, this is run when a tablet changes to Primary type. The other replicas
	// might not have started replication. So if we run the create commands, it will
	// block this thread, and we could end up in a deadlock.
	// Instead, we try creating the database and table in each tick which runs in a go routine
	// keeping us safe from hanging the main thread.
	w.appPool.Open(w.env.Config().DB.AppWithDB())
	w.allPrivsPool.Open(w.env.Config().DB.AllPrivsWithDB())
	if w.onDemandDuration == 0 {
		w.enableWrites(true)
		// when onDemandDuration > 0 we only enable writes per request
	}
	w.isOpen = true
}

// Close closes the heartbeatWriter's db connection and stops the periodic ticker.
func (w *heartbeatWriter) Close() {
	if !w.enabled {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.isOpen {
		return
	}

	w.enableWrites(false)
	w.appPool.Close()
	w.allPrivsPool.Close()
	w.isOpen = false
	log.Info("Hearbeat Writer: closed")
}

// bindHeartbeatVars takes a heartbeat write (insert or update) and
// adds the necessary fields to the query as bind vars. This is done
// to protect ourselves against a badly formed keyspace or shard name.
func (w *heartbeatWriter) bindHeartbeatVars(query string) (string, error) {
	bindVars := map[string]*querypb.BindVariable{
		"ks":  sqltypes.StringBindVariable(w.keyspaceShard),
		"ts":  sqltypes.Int64BindVariable(w.now().UnixNano()),
		"uid": sqltypes.Int64BindVariable(int64(w.tabletAlias.Uid)),
	}
	parsed := sqlparser.BuildParsedQuery(query, "_vt", ":ts", ":uid", ":ks")
	bound, err := parsed.GenerateQuery(bindVars, nil)
	if err != nil {
		return "", err
	}
	return bound, nil
}

// writeHeartbeat updates the heartbeat row for this tablet with the current time in nanoseconds.
func (w *heartbeatWriter) writeHeartbeat() {
	if err := w.write(); err != nil {
		w.recordError(err)
		return
	}
	writes.Add(1)
}

func (w *heartbeatWriter) write() error {
	defer w.env.LogError()
	ctx, cancel := context.WithDeadline(context.Background(), w.now().Add(w.interval))
	defer cancel()
	allPrivsConn, err := w.allPrivsPool.Get(ctx)
	if err != nil {
		return err
	}
	defer allPrivsConn.Recycle()

	upsert, err := w.bindHeartbeatVars(sqlUpsertHeartbeat)
	if err != nil {
		return err
	}
	appConn, err := w.appPool.Get(ctx)
	if err != nil {
		return err
	}
	defer appConn.Recycle()
	_, err = withDDL.Exec(ctx, upsert, appConn.ExecuteFetch, allPrivsConn.ExecuteFetch)
	if err != nil {
		return err
	}
	return nil
}

func (w *heartbeatWriter) recordError(err error) {
	w.errorLog.Errorf("%v", err)
	writeErrors.Add(1)
}

// enableWrites actives or deactives heartbeat writes
func (w *heartbeatWriter) enableWrites(enable bool) {
	if w.ticks == nil {
		return
	}
	if enable {
		w.ticks.Start(w.writeHeartbeat)
	} else {
		w.ticks.Stop()
	}
}

func (w *heartbeatWriter) RequestHeartbeats() {
	if w.onDemandDuration == 0 {
		// heartbeats are not by demand. Therefore they are just coming in on their own (if enabled)
		return
	}
	// In this function we're going to create a timer to activate heartbeats by-demand. Creating a timer has a cost.
	// Now, this function can be spammed by clients (the lag throttler). We therefore only allow this function to
	// actually operate once per X seconds (1/4 of onDemandDuration as a reasonable oversampling value):
	if atomic.CompareAndSwapInt64(&w.requestHeartbeatsReentranceFlag, 0, 1) {
		defer time.AfterFunc(w.onDemandDuration/4, func() { atomic.StoreInt64(&w.requestHeartbeatsReentranceFlag, 0) })
	} else {
		// An instance of this function is already running
		return
	}

	// Now for the actual logic. A client requests heartbeats. If it were only this client, we would
	// activate heartbeats for the duration of onDemandDuration, and then turn heartbeats off.
	// However, there may be multiple clients interested in heartbeats, or maybe the same single client
	// requesting heartbeats again and again. So we keep track of how many _concurrent_ requests we have.
	// We enable heartbeats as soon as we have a single concurrent request; we turn heartbeats off once
	// we have zero concurrent requests
	if numClients := atomic.AddInt64(&w.concurrentHeartbeatRequests, 1); numClients == 1 {
		// means we previously had 0 clients interested in heartbeats.
		w.enableWrites(true)
	}
	time.AfterFunc(w.onDemandDuration, func() {
		if numClients := atomic.AddInt64(&w.concurrentHeartbeatRequests, -1); numClients == 0 {
			// means there are currently no more clients interested in heartbeats
			w.enableWrites(false)
		}
	})
}
