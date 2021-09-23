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

package testlib

import (
	"context"
	"testing"
	"time"

	"vitess.io/vitess/go/vt/discovery"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/sets"

	"vitess.io/vitess/go/mysql"
	"vitess.io/vitess/go/vt/logutil"
	"vitess.io/vitess/go/vt/topo/memorytopo"
	"vitess.io/vitess/go/vt/topo/topoproto"
	"vitess.io/vitess/go/vt/vttablet/tmclient"
	"vitess.io/vitess/go/vt/wrangler"

	topodatapb "vitess.io/vitess/go/vt/proto/topodata"
)

func TestEmergencyReparentShard(t *testing.T) {
	delay := discovery.GetTabletPickerRetryDelay()
	defer func() {
		discovery.SetTabletPickerRetryDelay(delay)
	}()
	discovery.SetTabletPickerRetryDelay(5 * time.Millisecond)

	ts := memorytopo.NewServer("cell1", "cell2")
	wr := wrangler.New(logutil.NewConsoleLogger(), ts, tmclient.NewTabletManagerClient())
	vp := NewVtctlPipe(t, ts)
	defer vp.Close()

	// Create a primary, a couple good replicas
	oldPrimary := NewFakeTablet(t, wr, "cell1", 0, topodatapb.TabletType_PRIMARY, nil)
	newPrimary := NewFakeTablet(t, wr, "cell1", 1, topodatapb.TabletType_REPLICA, nil)
	goodReplica1 := NewFakeTablet(t, wr, "cell1", 2, topodatapb.TabletType_REPLICA, nil)
	goodReplica2 := NewFakeTablet(t, wr, "cell2", 3, topodatapb.TabletType_REPLICA, nil)

	oldPrimary.FakeMysqlDaemon.Replicating = false
	oldPrimary.FakeMysqlDaemon.CurrentPrimaryPosition = mysql.Position{
		GTIDSet: mysql.MariadbGTIDSet{
			2: mysql.MariadbGTID{
				Domain:   2,
				Server:   123,
				Sequence: 456,
			},
		},
	}
	currentPrimaryFilePosition, _ := mysql.ParseFilePosGTIDSet("mariadb-bin.000010:456")
	oldPrimary.FakeMysqlDaemon.CurrentSourceFilePosition = mysql.Position{
		GTIDSet: currentPrimaryFilePosition,
	}

	// new primary
	newPrimary.FakeMysqlDaemon.ReadOnly = true
	newPrimary.FakeMysqlDaemon.Replicating = true
	newPrimary.FakeMysqlDaemon.CurrentPrimaryPosition = mysql.Position{
		GTIDSet: mysql.MariadbGTIDSet{
			2: mysql.MariadbGTID{
				Domain:   2,
				Server:   123,
				Sequence: 456,
			},
		},
	}
	newPrimaryRelayLogPos, _ := mysql.ParseFilePosGTIDSet("relay-bin.000004:456")
	newPrimary.FakeMysqlDaemon.CurrentSourceFilePosition = mysql.Position{
		GTIDSet: newPrimaryRelayLogPos,
	}
	newPrimary.FakeMysqlDaemon.WaitPrimaryPosition = newPrimary.FakeMysqlDaemon.CurrentSourceFilePosition
	newPrimary.FakeMysqlDaemon.ExpectedExecuteSuperQueryList = []string{
		"STOP SLAVE IO_THREAD",
		"CREATE DATABASE IF NOT EXISTS _vt",
		"SUBCREATE TABLE IF NOT EXISTS _vt.reparent_journal",
		"SUBINSERT INTO _vt.reparent_journal (time_created_ns, action_name, master_alias, replication_position) VALUES",
	}
	newPrimary.FakeMysqlDaemon.PromoteResult = mysql.Position{
		GTIDSet: mysql.MariadbGTIDSet{
			2: mysql.MariadbGTID{
				Domain:   2,
				Server:   123,
				Sequence: 456,
			},
		},
	}
	newPrimary.StartActionLoop(t, wr)
	defer newPrimary.StopActionLoop(t)

	// old primary, will be scrapped
	oldPrimary.FakeMysqlDaemon.ReadOnly = false
	oldPrimary.FakeMysqlDaemon.ReplicationStatusError = mysql.ErrNotReplica
	oldPrimary.FakeMysqlDaemon.SetReplicationSourceInput = topoproto.MysqlAddr(newPrimary.Tablet)
	oldPrimary.FakeMysqlDaemon.ExpectedExecuteSuperQueryList = []string{
		"STOP SLAVE",
	}
	oldPrimary.StartActionLoop(t, wr)
	defer oldPrimary.StopActionLoop(t)

	// good replica 1 is replicating
	goodReplica1.FakeMysqlDaemon.ReadOnly = true
	goodReplica1.FakeMysqlDaemon.Replicating = true
	goodReplica1.FakeMysqlDaemon.CurrentPrimaryPosition = mysql.Position{
		GTIDSet: mysql.MariadbGTIDSet{
			2: mysql.MariadbGTID{
				Domain:   2,
				Server:   123,
				Sequence: 455,
			},
		},
	}
	goodReplica1RelayLogPos, _ := mysql.ParseFilePosGTIDSet("relay-bin.000004:455")
	goodReplica1.FakeMysqlDaemon.CurrentSourceFilePosition = mysql.Position{
		GTIDSet: goodReplica1RelayLogPos,
	}
	goodReplica1.FakeMysqlDaemon.WaitPrimaryPosition = goodReplica1.FakeMysqlDaemon.CurrentSourceFilePosition
	goodReplica1.FakeMysqlDaemon.SetReplicationSourceInput = topoproto.MysqlAddr(newPrimary.Tablet)
	goodReplica1.FakeMysqlDaemon.ExpectedExecuteSuperQueryList = []string{
		"STOP SLAVE IO_THREAD",
		"STOP SLAVE",
		"FAKE SET MASTER",
		"START SLAVE",
	}
	goodReplica1.StartActionLoop(t, wr)
	defer goodReplica1.StopActionLoop(t)

	// good replica 2 is not replicating
	goodReplica2.FakeMysqlDaemon.ReadOnly = true
	goodReplica2.FakeMysqlDaemon.Replicating = false
	goodReplica2.FakeMysqlDaemon.CurrentPrimaryPosition = mysql.Position{
		GTIDSet: mysql.MariadbGTIDSet{
			2: mysql.MariadbGTID{
				Domain:   2,
				Server:   123,
				Sequence: 454,
			},
		},
	}
	goodReplica2RelayLogPos, _ := mysql.ParseFilePosGTIDSet("relay-bin.000004:454")
	goodReplica2.FakeMysqlDaemon.CurrentSourceFilePosition = mysql.Position{
		GTIDSet: goodReplica2RelayLogPos,
	}
	goodReplica2.FakeMysqlDaemon.WaitPrimaryPosition = goodReplica2.FakeMysqlDaemon.CurrentSourceFilePosition
	goodReplica2.FakeMysqlDaemon.SetReplicationSourceInput = topoproto.MysqlAddr(newPrimary.Tablet)
	goodReplica2.StartActionLoop(t, wr)
	goodReplica2.FakeMysqlDaemon.ExpectedExecuteSuperQueryList = []string{
		"FAKE SET MASTER",
	}
	defer goodReplica2.StopActionLoop(t)

	// run EmergencyReparentShard
	// using deprecated flag until it is removed completely. at that time this should be replaced with -wait_replicas_timeout
	waitReplicaTimeout := time.Second * 2
	err := vp.Run([]string{"EmergencyReparentShard", "-wait_replicas_timeout", waitReplicaTimeout.String(), newPrimary.Tablet.Keyspace + "/" + newPrimary.Tablet.Shard,
		topoproto.TabletAliasString(newPrimary.Tablet.Alias)})
	require.NoError(t, err)
	// check what was run
	err = newPrimary.FakeMysqlDaemon.CheckSuperQueryList()
	require.NoError(t, err)

	assert.False(t, newPrimary.FakeMysqlDaemon.ReadOnly, "newPrimary.FakeMysqlDaemon.ReadOnly set")
	checkSemiSyncEnabled(t, true, true, newPrimary)
}

// TestEmergencyReparentShardPrimaryElectNotBest tries to emergency reparent
// to a host that is not the latest in replication position.
func TestEmergencyReparentShardPrimaryElectNotBest(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	delay := discovery.GetTabletPickerRetryDelay()
	defer func() {
		discovery.SetTabletPickerRetryDelay(delay)
	}()
	discovery.SetTabletPickerRetryDelay(5 * time.Millisecond)

	ts := memorytopo.NewServer("cell1", "cell2")
	wr := wrangler.New(logutil.NewConsoleLogger(), ts, tmclient.NewTabletManagerClient())

	// Create a primary, a couple good replicas
	oldPrimary := NewFakeTablet(t, wr, "cell1", 0, topodatapb.TabletType_PRIMARY, nil)
	newPrimary := NewFakeTablet(t, wr, "cell1", 1, topodatapb.TabletType_REPLICA, nil)
	moreAdvancedReplica := NewFakeTablet(t, wr, "cell1", 2, topodatapb.TabletType_REPLICA, nil)

	// new primary
	newPrimary.FakeMysqlDaemon.Replicating = true
	// this server has executed upto 455, which is the highest among replicas
	newPrimary.FakeMysqlDaemon.CurrentPrimaryPosition = mysql.Position{
		GTIDSet: mysql.MariadbGTIDSet{
			2: mysql.MariadbGTID{
				Domain:   2,
				Server:   123,
				Sequence: 455,
			},
		},
	}
	// It has more transactions in its relay log, but not as many as
	// moreAdvancedReplica
	newPrimary.FakeMysqlDaemon.CurrentPrimaryPosition = mysql.Position{
		GTIDSet: mysql.MariadbGTIDSet{
			2: mysql.MariadbGTID{
				Domain:   2,
				Server:   123,
				Sequence: 456,
			},
		},
	}
	newPrimaryRelayLogPos, _ := mysql.ParseFilePosGTIDSet("relay-bin.000004:456")
	newPrimary.FakeMysqlDaemon.CurrentSourceFilePosition = mysql.Position{
		GTIDSet: newPrimaryRelayLogPos,
	}
	newPrimary.FakeMysqlDaemon.WaitPrimaryPosition = newPrimary.FakeMysqlDaemon.CurrentSourceFilePosition
	newPrimary.FakeMysqlDaemon.ExpectedExecuteSuperQueryList = []string{
		"STOP SLAVE IO_THREAD",
	}
	newPrimary.StartActionLoop(t, wr)
	defer newPrimary.StopActionLoop(t)

	// old primary, will be scrapped
	oldPrimary.FakeMysqlDaemon.ReplicationStatusError = mysql.ErrNotReplica
	oldPrimary.StartActionLoop(t, wr)
	defer oldPrimary.StopActionLoop(t)

	// more advanced replica
	moreAdvancedReplica.FakeMysqlDaemon.Replicating = true
	// position up to which this replica has executed is behind desired new primary
	moreAdvancedReplica.FakeMysqlDaemon.CurrentPrimaryPosition = mysql.Position{
		GTIDSet: mysql.MariadbGTIDSet{
			2: mysql.MariadbGTID{
				Domain:   2,
				Server:   123,
				Sequence: 454,
			},
		},
	}
	// relay log position is more advanced than desired new primary
	moreAdvancedReplica.FakeMysqlDaemon.CurrentPrimaryPosition = mysql.Position{
		GTIDSet: mysql.MariadbGTIDSet{
			2: mysql.MariadbGTID{
				Domain:   2,
				Server:   123,
				Sequence: 457,
			},
		},
	}
	moreAdvancedReplicaLogPos, _ := mysql.ParseFilePosGTIDSet("relay-bin.000004:457")
	moreAdvancedReplica.FakeMysqlDaemon.CurrentSourceFilePosition = mysql.Position{
		GTIDSet: moreAdvancedReplicaLogPos,
	}
	moreAdvancedReplica.FakeMysqlDaemon.WaitPrimaryPosition = moreAdvancedReplica.FakeMysqlDaemon.CurrentSourceFilePosition
	moreAdvancedReplica.FakeMysqlDaemon.ExpectedExecuteSuperQueryList = []string{
		"STOP SLAVE IO_THREAD",
	}
	moreAdvancedReplica.StartActionLoop(t, wr)
	defer moreAdvancedReplica.StopActionLoop(t)

	// run EmergencyReparentShard
	err := wr.EmergencyReparentShard(ctx, newPrimary.Tablet.Keyspace, newPrimary.Tablet.Shard, newPrimary.Tablet.Alias, 10*time.Second, sets.NewString(), false)
	cancel()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not fully caught up")
	// check what was run
	err = newPrimary.FakeMysqlDaemon.CheckSuperQueryList()
	require.NoError(t, err)
	err = oldPrimary.FakeMysqlDaemon.CheckSuperQueryList()
	require.NoError(t, err)
	err = moreAdvancedReplica.FakeMysqlDaemon.CheckSuperQueryList()
	require.NoError(t, err)
}
