package dal

import (
	"context"
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"

	cronsql "github.com/TBD54566975/ftl/backend/controller/cronjobs/internal/sql"
	"github.com/TBD54566975/ftl/backend/controller/observability"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/slices"
)

type DAL struct {
	*libdal.Handle[DAL]
	db cronsql.Querier
}

func New(conn libdal.Connection) *DAL {
	return &DAL{
		db: cronsql.New(conn),
		Handle: libdal.New(conn, func(h *libdal.Handle[DAL]) *DAL {
			return &DAL{Handle: h, db: cronsql.New(h.Connection)}
		}),
	}
}

func cronJobFromRow(c cronsql.CronJob, d cronsql.Deployment) model.CronJob {
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

type CreateAsyncCallParams cronsql.CreateAsyncCallParams

// CreateAsyncCall creates an async_call row and returns its id
func (d *DAL) CreateAsyncCall(ctx context.Context, params CreateAsyncCallParams) (int64, error) {
	id, err := d.db.CreateAsyncCall(ctx, cronsql.CreateAsyncCallParams(params))
	if err != nil {
		return 0, fmt.Errorf("failed to create async call: %w", libdal.TranslatePGError(err))
	}
	observability.AsyncCalls.Created(ctx, params.Verb, optional.None[schema.RefKey](), params.Origin, 0, err)
	queueDepth, err := d.db.AsyncCallQueueDepth(ctx)
	if err == nil {
		// Don't error out of a transition just over a queue depth retrieval
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
		return nil, fmt.Errorf("failed to get cron jobs: %w", libdal.TranslatePGError(err))
	}
	return slices.Map(rows, func(r cronsql.GetUnscheduledCronJobsRow) model.CronJob {
		return cronJobFromRow(r.CronJob, r.Deployment)
	}), nil
}

// GetCronJobByKey returns a cron_job row by its key
func (d *DAL) GetCronJobByKey(ctx context.Context, key model.CronJobKey) (model.CronJob, error) {
	row, err := d.db.GetCronJobByKey(ctx, key)
	if err != nil {
		return model.CronJob{}, fmt.Errorf("failed to get cron job %q: %w", key, libdal.TranslatePGError(err))
	}
	return cronJobFromRow(row.CronJob, row.Deployment), nil
}

// IsCronJobPending returns whether this cron job is executing or scheduled in async_calls
func (d *DAL) IsCronJobPending(ctx context.Context, key model.CronJobKey, startTime time.Time) (bool, error) {
	pending, err := d.db.IsCronJobPending(ctx, key, startTime)
	if err != nil {
		return false, fmt.Errorf("failed to check if cron job %q is pending: %w", key, libdal.TranslatePGError(err))
	}
	return pending, nil
}

type UpdateCronJobExecutionParams cronsql.UpdateCronJobExecutionParams

// UpdateCronJobExecution updates the last_async_call_id, last_execution, and next_execution of
// the cron job given by the provided key
func (d *DAL) UpdateCronJobExecution(ctx context.Context, params UpdateCronJobExecutionParams) error {
	err := d.db.UpdateCronJobExecution(ctx, cronsql.UpdateCronJobExecutionParams(params))
	if err != nil {
		return fmt.Errorf("failed to update cron job %q: %w", params.Key, libdal.TranslatePGError(err))
	}
	return nil
}

func (d *DAL) DeleteCronJobsForDeployment(ctx context.Context, key model.DeploymentKey) error {
	err := d.db.DeleteCronAsyncCallsForDeployment(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete cron job async calls for deployment %v: %w", key, libdal.TranslatePGError(err))
	}
	err = d.db.DeleteCronJobsForDeployment(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete cron jobs for deployment %v: %w", key, libdal.TranslatePGError(err))
	}
	return nil
}
