package cronjobs

import (
	"context"
	"fmt"
	"time"

	"github.com/TBD54566975/ftl/backend/controller/cronjobs/sql"
	dalerrs "github.com/TBD54566975/ftl/backend/dal"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/slices"
)

type DAL struct {
	db sql.DBI
}

func newDAL(conn sql.ConnI) *DAL {
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

func cronJobFromGetByKeyRow(row sql.GetCronJobByKeyRow) model.CronJob {
	return model.CronJob{
		Key:           row.Key,
		DeploymentKey: row.DeploymentKey,
		Verb:          schema.Ref{Module: row.Module, Name: row.Verb},
		Schedule:      row.Schedule,
		StartTime:     row.StartTime,
		NextExecution: row.NextExecution,
		LastExecution: row.LastExecution,
	}
}

func cronJobFromGetUnscheduledRow(row sql.GetUnscheduledCronJobsRow) model.CronJob {
	return model.CronJob{
		Key:           row.Key,
		DeploymentKey: row.DeploymentKey,
		Verb:          schema.Ref{Module: row.Module, Name: row.Verb},
		Schedule:      row.Schedule,
		StartTime:     row.StartTime,
		NextExecution: row.NextExecution,
		LastExecution: row.LastExecution,
	}
}

// GetUnscheduledCronJobs returns all cron jobs with start_time before provided startTime for
// deployments with min replicas > 0 with no async calls after last_execution
func (d *DAL) GetUnscheduledCronJobs(ctx context.Context, startTime time.Time) ([]model.CronJob, error) {
	rows, err := d.db.GetUnscheduledCronJobs(ctx, startTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get cron jobs: %w", dalerrs.TranslatePGError(err))
	}
	return slices.Map(rows, cronJobFromGetUnscheduledRow), nil
}

// GetCronJobByKey returns a cron job by its key
func (d *DAL) GetCronJobByKey(ctx context.Context, key model.CronJobKey) (model.CronJob, error) {
	row, err := d.db.GetCronJobByKey(ctx, key)
	if err != nil {
		return model.CronJob{}, fmt.Errorf("failed to get cron job %q: %w", key, dalerrs.TranslatePGError(err))
	}
	return cronJobFromGetByKeyRow(row), nil
}
