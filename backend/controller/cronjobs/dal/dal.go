// Package dal provides a data abstraction layer for cron jobs
package dal

import (
	"context"
	"fmt"
	"time"

	"github.com/TBD54566975/ftl/backend/controller/cronjobs/sql"
	"github.com/TBD54566975/ftl/backend/controller/sql/sqltypes"
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

func cronJobFromRow(row sql.GetCronJobsRow) model.CronJob {
	return model.CronJob{
		Key:           row.Key,
		DeploymentKey: row.DeploymentKey,
		Verb:          schema.Ref{Module: row.Module, Name: row.Verb},
		Schedule:      row.Schedule,
		StartTime:     row.StartTime,
		NextExecution: row.NextExecution,
		State:         row.State,
	}
}

// GetCronJobs returns all cron jobs for deployments with min replicas > 0
func (d *DAL) GetCronJobs(ctx context.Context) ([]model.CronJob, error) {
	rows, err := d.db.GetCronJobs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cron jobs: %w", dalerrs.TranslatePGError(err))
	}
	return slices.Map(rows, cronJobFromRow), nil
}

type AttemptedCronJob struct {
	DidStartExecution bool
	HasMinReplicas    bool
	model.CronJob
}

// StartCronJobs returns a full list of results so that the caller can update their list of jobs whether or not they successfully updated the row
func (d *DAL) StartCronJobs(ctx context.Context, jobs []model.CronJob) (attemptedJobs []AttemptedCronJob, err error) {
	if len(jobs) == 0 {
		return nil, nil
	}
	rows, err := d.db.StartCronJobs(ctx, slices.Map(jobs, func(job model.CronJob) string { return job.Key.String() }))
	if err != nil {
		return nil, fmt.Errorf("failed to start cron jobs: %w", dalerrs.TranslatePGError(err))
	}

	attemptedJobs = []AttemptedCronJob{}
	for _, row := range rows {
		job := AttemptedCronJob{
			CronJob: model.CronJob{
				Key:           row.Key,
				DeploymentKey: row.DeploymentKey,
				Verb:          schema.Ref{Module: row.Module, Name: row.Verb},
				Schedule:      row.Schedule,
				StartTime:     row.StartTime,
				NextExecution: row.NextExecution,
				State:         row.State,
			},
			DidStartExecution: row.Updated,
			HasMinReplicas:    row.HasMinReplicas,
		}
		attemptedJobs = append(attemptedJobs, job)
	}
	return attemptedJobs, nil
}

// EndCronJob sets the status from executing to idle and updates the next execution time
// Can be called on the successful completion of a job, or if the job failed to execute (error or timeout)
func (d *DAL) EndCronJob(ctx context.Context, job model.CronJob, next time.Time) (model.CronJob, error) {
	row, err := d.db.EndCronJob(ctx, next, job.Key, job.StartTime)
	if err != nil {
		return model.CronJob{}, fmt.Errorf("failed to end cron job: %w", dalerrs.TranslatePGError(err))
	}
	return cronJobFromRow(sql.GetCronJobsRow(row)), nil
}

// GetStaleCronJobs returns a list of cron jobs that have been executing longer than the duration
func (d *DAL) GetStaleCronJobs(ctx context.Context, duration time.Duration) ([]model.CronJob, error) {
	rows, err := d.db.GetStaleCronJobs(ctx, sqltypes.Duration(duration))
	if err != nil {
		return nil, fmt.Errorf("failed to get stale cron jobs: %w", dalerrs.TranslatePGError(err))
	}
	return slices.Map(rows, func(row sql.GetStaleCronJobsRow) model.CronJob {
		return cronJobFromRow(sql.GetCronJobsRow(row))
	}), nil
}
