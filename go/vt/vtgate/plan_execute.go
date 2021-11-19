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

package vtgate

import (
	"context"
	"time"

	"vitess.io/vitess/go/sqltypes"
	querypb "vitess.io/vitess/go/vt/proto/query"
	vtrpcpb "vitess.io/vitess/go/vt/proto/vtrpc"
	"vitess.io/vitess/go/vt/sqlparser"
	"vitess.io/vitess/go/vt/vterrors"
	"vitess.io/vitess/go/vt/vtgate/engine"
	"vitess.io/vitess/go/vt/vtgate/planbuilder"
)

type planExec func(plan *engine.Plan, vc *vcursorImpl, bindVars map[string]*querypb.BindVariable, startTime time.Time) error
type txResult func(sqlparser.StatementType, *sqltypes.Result) error

func (e *Executor) newExecute(
	ctx context.Context,
	safeSession *SafeSession,
	sql string,
	bindVars map[string]*querypb.BindVariable,
	logStats *LogStats,
	execPlan planExec, // used when there is a plan to execute
	recResult txResult, // used when it's something simple like begin/commit/rollback/savepoint
) error {
	// 1: Prepare before planning and execution

	// Start an implicit transaction if necessary.
	err := e.startTxIfNecessary(ctx, safeSession)
	if err != nil {
		return err
	}

	if bindVars == nil {
		bindVars = make(map[string]*querypb.BindVariable)
	}

	query, comments := sqlparser.SplitMarginComments(sql)
	vcursor, err := newVCursorImpl(ctx, safeSession, comments, e, logStats, e.vm, e.VSchema(), e.resolver.resolver, e.serv, e.warnShardedOnly)
	if err != nil {
		return err
	}

	// 2: Create a plan for the query
	plan, err := e.getPlan(
		vcursor,
		query,
		comments,
		bindVars,
		safeSession,
		logStats,
	)
	if err == planbuilder.ErrPlanNotSupported {
		return err
	}
	execStart := e.logPlanningFinished(logStats, plan)

	if err != nil {
		safeSession.ClearWarnings()
		return err
	}

	if plan.Type != sqlparser.StmtShow {
		safeSession.ClearWarnings()
	}

	// add any warnings that the planner wants to add
	for _, warning := range plan.Warnings {
		safeSession.RecordWarning(warning)
	}

	result, err := e.handleTransactions(ctx, safeSession, plan, logStats, vcursor)
	if err != nil {
		return err
	}
	if result != nil {
		return recResult(plan.Type, result)
	}

	// 3: Prepare for execution
	err = e.addNeededBindVars(plan.BindVarNeeds, bindVars, safeSession)
	if err != nil {
		logStats.Error = err
		return err
	}

	if plan.Instructions.NeedsTransaction() {
		return e.insideTransaction(ctx, safeSession, logStats,
			func() error {
				return execPlan(plan, vcursor, bindVars, execStart)
			})
	}

	return execPlan(plan, vcursor, bindVars, execStart)
}

// handleTransactions deals with transactional queries: begin, commit, rollback and savepoint management
func (e *Executor) handleTransactions(ctx context.Context, safeSession *SafeSession, plan *engine.Plan, logStats *LogStats, vcursor *vcursorImpl) (*sqltypes.Result, error) {
	// We need to explicitly handle errors, and begin/commit/rollback, since these control transactions. Everything else
	// will fall through and be handled through planning
	switch plan.Type {
	case sqlparser.StmtBegin:
		qr, err := e.handleBegin(ctx, safeSession, logStats)
		return qr, err
	case sqlparser.StmtCommit:
		qr, err := e.handleCommit(ctx, safeSession, logStats)
		return qr, err
	case sqlparser.StmtRollback:
		qr, err := e.handleRollback(ctx, safeSession, logStats)
		return qr, err
	case sqlparser.StmtSavepoint:
		qr, err := e.handleSavepoint(ctx, safeSession, plan.Original, "Savepoint", logStats, func(_ string) (*sqltypes.Result, error) {
			// Safely to ignore as there is no transaction.
			return &sqltypes.Result{}, nil
		}, vcursor.ignoreMaxMemoryRows)
		return qr, err
	case sqlparser.StmtSRollback:
		qr, err := e.handleSavepoint(ctx, safeSession, plan.Original, "Rollback Savepoint", logStats, func(query string) (*sqltypes.Result, error) {
			// Error as there is no transaction, so there is no savepoint that exists.
			return nil, vterrors.NewErrorf(vtrpcpb.Code_NOT_FOUND, vterrors.SPDoesNotExist, "SAVEPOINT does not exist: %s", query)
		}, vcursor.ignoreMaxMemoryRows)
		return qr, err
	case sqlparser.StmtRelease:
		qr, err := e.handleSavepoint(ctx, safeSession, plan.Original, "Release Savepoint", logStats, func(query string) (*sqltypes.Result, error) {
			// Error as there is no transaction, so there is no savepoint that exists.
			return nil, vterrors.NewErrorf(vtrpcpb.Code_NOT_FOUND, vterrors.SPDoesNotExist, "SAVEPOINT does not exist: %s", query)
		}, vcursor.ignoreMaxMemoryRows)
		return qr, err
	}
	return nil, nil
}

func (e *Executor) startTxIfNecessary(ctx context.Context, safeSession *SafeSession) error {
	if !safeSession.Autocommit && !safeSession.InTransaction() {
		if err := e.txConn.Begin(ctx, safeSession); err != nil {
			return err
		}
	}
	return nil
}

func (e *Executor) insideTransaction(ctx context.Context, safeSession *SafeSession, logStats *LogStats, execPlan func() error) error {
	mustCommit := false
	if safeSession.Autocommit && !safeSession.InTransaction() {
		mustCommit = true
		if err := e.txConn.Begin(ctx, safeSession); err != nil {
			return err
		}
		// The defer acts as a failsafe. If commit was successful,
		// the rollback will be a no-op.
		defer e.txConn.Rollback(ctx, safeSession) // nolint:errcheck
	}

	// The SetAutocommitable flag should be same as mustCommit.
	// If we started a transaction because of autocommit, then mustCommit
	// will be true, which means that we can autocommit. If we were already
	// in a transaction, it means that the app started it, or we are being
	// called recursively. If so, we cannot autocommit because whatever we
	// do is likely not final.
	// The control flow is such that autocommitable can only be turned on
	// at the beginning, but never after.
	safeSession.SetAutocommittable(mustCommit)

	// Execute!
	err := execPlan()
	if err != nil {
		return err
	}

	if mustCommit {
		commitStart := time.Now()
		if err := e.txConn.Commit(ctx, safeSession); err != nil {
			return err
		}
		logStats.CommitTime = time.Since(commitStart)
	}
	return nil
}

func (e *Executor) executePlan(
	ctx context.Context,
	safeSession *SafeSession,
	plan *engine.Plan,
	vcursor *vcursorImpl,
	bindVars map[string]*querypb.BindVariable,
	logStats *LogStats,
	execStart time.Time,
) (*sqltypes.Result, error) {

	// 4: Execute!
	qr, err := vcursor.ExecutePrimitive(plan.Instructions, bindVars, true)

	// 5: Log and add statistics
	e.setLogStats(logStats, plan, vcursor, execStart, err, qr)

	// Check if there was partial DML execution. If so, rollback the transaction.
	if err != nil && safeSession.InTransaction() && vcursor.rollbackOnPartialExec != "" {
		_, _, sErr := e.execute(ctx, safeSession, vcursor.rollbackOnPartialExec, bindVars, logStats)
		if sErr == nil {
			err = vterrors.Errorf(vtrpcpb.Code_ABORTED, "failed to execute due to partial DML execution: %v", err)
		} else {
			// not able to rollback changes of the failed query, so have to abort the complete transaction.
			_ = e.txConn.Rollback(ctx, safeSession)
			err = vterrors.Errorf(vtrpcpb.Code_ABORTED, "transaction rolled back due to not able to reverse changes of partial DML execution: %v:%v", sErr, err)
		}
	}
	return qr, err
}

func (e *Executor) setLogStats(logStats *LogStats, plan *engine.Plan, vcursor *vcursorImpl, execStart time.Time, err error, qr *sqltypes.Result) {
	logStats.StmtType = plan.Type.String()
	logStats.Keyspace = plan.Instructions.GetKeyspaceName()
	logStats.Table = plan.Instructions.GetTableName()
	logStats.TabletType = vcursor.TabletType().String()
	errCount := e.logExecutionEnd(logStats, execStart, plan, err, qr)
	plan.AddStats(1, time.Since(logStats.StartTime), logStats.ShardQueries, logStats.RowsAffected, logStats.RowsReturned, errCount)
}

func (e *Executor) logExecutionEnd(logStats *LogStats, execStart time.Time, plan *engine.Plan, err error, qr *sqltypes.Result) uint64 {
	logStats.ExecuteTime = time.Since(execStart)

	e.updateQueryCounts(plan.Instructions.RouteType(), plan.Instructions.GetKeyspaceName(), plan.Instructions.GetTableName(), int64(logStats.ShardQueries))

	var errCount uint64
	if err != nil {
		logStats.Error = err
		errCount = 1
	} else {
		logStats.RowsAffected = qr.RowsAffected
		logStats.RowsReturned = uint64(len(qr.Rows))
	}
	return errCount
}

func (e *Executor) logPlanningFinished(logStats *LogStats, plan *engine.Plan) time.Time {
	execStart := time.Now()
	if plan != nil {
		logStats.StmtType = plan.Type.String()
	}
	logStats.PlanTime = execStart.Sub(logStats.StartTime)
	return execStart
}
