package dal

import (
	"context"
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/cronjobs/sql"
	"github.com/TBD54566975/ftl/backend/controller/observability"
	dalerrs "github.com/TBD54566975/ftl/backend/dal"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/slices"
)

type DAL struct {
	db sql.DBI
}

func New(conn sql.ConnI) *DAL {
	return &DAL{db: sql.NewDB(conn)}
}

type Tx struct {
	*DAL
}

func (d *DAL) Begin(ctx context.Context) (*Tx, error) {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", dalerrs.TranslatePGError(err))
	}
	return &Tx{DAL: &DAL{db: tx}}, nil
}

func (t *Tx) CommitOrRollback(ctx context.Context, err *error) {
	tx, ok := t.db.(*sql.Tx)
	if !ok {
		panic("inconceivable")
	}
	tx.CommitOrRollback(ctx, err)
}

func (t *Tx) Commit(ctx context.Context) error {
	tx, ok := t.db.(*sql.Tx)
	if !ok {
		panic("inconcievable")
	}
	err := tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", dalerrs.TranslatePGError(err))
	}
	return nil
}

func (t *Tx) Rollback(ctx context.Context) error {
	tx, ok := t.db.(*sql.Tx)
	if !ok {
		panic("inconcievable")
	}
	err := tx.Rollback(ctx)
	if err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", dalerrs.TranslatePGError(err))
	}
	return nil
}

func cronJobFromRow(c sql.CronJob, d sql.Deployment) model.CronJob {
	return model.CronJob{
		Key:           c.Key,
		DeploymentKey: d.Key,
		Verb:          schema.Ref{Module: c.ModuleName, Name: c.Verb},
		Schedule:      c.Schedule,
		StartTime:     c.StartTime,
		NextExecution: c.NextExecution,
		LastExecution: c.LastExecution,
	}
}

// CreateAsyncCall creates an async_call row and returns its id
func (d *DAL) CreateAsyncCall(ctx context.Context, params sql.CreateAsyncCallParams) (int64, error) {
	id, err := d.db.CreateAsyncCall(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("failed to create async call: %w", dalerrs.TranslatePGError(err))
	}
	observability.AsyncCalls.Created(ctx, params.Verb, optional.None[schema.RefKey](), params.Origin, 0, err)
	queueDepth, err := d.db.AsyncCallQueueDepth(ctx)
	if err == nil {
		// Don't error out of an FSM transition just over a queue depth retrieval
		// error because this is only used for an observability gauge.
		observability.AsyncCalls.RecordQueueDepth(ctx, queueDepth)
	}
	return id, nil
}

// GetUnscheduledCronJobs returns all cron_jobs rows with start_time before provided startTime for
// deployments with min replicas > 0 with no pending corresponding async_calls after last_execution
func (d *DAL) GetUnscheduledCronJobs(ctx context.Context, startTime time.Time) ([]model.CronJob, error) {
	rows, err := d.db.GetUnscheduledCronJobs(ctx, startTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get cron jobs: %w", dalerrs.TranslatePGError(err))
	}
	return slices.Map(rows, func(r sql.GetUnscheduledCronJobsRow) model.CronJob {
		return cronJobFromRow(r.CronJob, r.Deployment)
	}), nil
}

// GetCronJobByKey returns a cron_job row by its key
func (d *DAL) GetCronJobByKey(ctx context.Context, key model.CronJobKey) (model.CronJob, error) {
	row, err := d.db.GetCronJobByKey(ctx, key)
	if err != nil {
		return model.CronJob{}, fmt.Errorf("failed to get cron job %q: %w", key, dalerrs.TranslatePGError(err))
	}
	return cronJobFromRow(row.CronJob, row.Deployment), nil
}

// IsCronJobPending returns whether this cron job is executing or scheduled in async_calls
func (d *DAL) IsCronJobPending(ctx context.Context, key model.CronJobKey, startTime time.Time) (bool, error) {
	pending, err := d.db.IsCronJobPending(ctx, key, startTime)
	if err != nil {
		return false, fmt.Errorf("failed to check if cron job %q is pending: %w", key, dalerrs.TranslatePGError(err))
	}
	return pending, nil
}

// UpdateCronJobExecution updates the last_async_call_id, last_execution, and next_execution of
// the cron job given by the provided key
func (d *DAL) UpdateCronJobExecution(ctx context.Context, params sql.UpdateCronJobExecutionParams) error {
	err := d.db.UpdateCronJobExecution(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update cron job %q: %w", params.Key, dalerrs.TranslatePGError(err))
	}
	return nil
}
