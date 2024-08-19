// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: queries.sql

package sql

import (
	"context"
	"time"

	"github.com/TBD54566975/ftl/internal/model"
)

const createCronJob = `-- name: CreateCronJob :exec
INSERT INTO cron_jobs (key, deployment_id, module_name, verb, schedule, start_time, next_execution)
  VALUES (
    $1::cron_job_key,
    (SELECT id FROM deployments WHERE key = $2::deployment_key LIMIT 1),
    $3::TEXT,
    $4::TEXT,
    $5::TEXT,
    $6::TIMESTAMPTZ,
    $7::TIMESTAMPTZ)
`

type CreateCronJobParams struct {
	Key           model.CronJobKey
	DeploymentKey model.DeploymentKey
	ModuleName    string
	Verb          string
	Schedule      string
	StartTime     time.Time
	NextExecution time.Time
}

func (q *Queries) CreateCronJob(ctx context.Context, arg CreateCronJobParams) error {
	_, err := q.db.ExecContext(ctx, createCronJob,
		arg.Key,
		arg.DeploymentKey,
		arg.ModuleName,
		arg.Verb,
		arg.Schedule,
		arg.StartTime,
		arg.NextExecution,
	)
	return err
}

const getCronJobByKey = `-- name: GetCronJobByKey :one
SELECT j.id, j.key, j.deployment_id, j.verb, j.schedule, j.start_time, j.next_execution, j.module_name, j.last_execution, d.id, d.created_at, d.module_id, d.key, d.schema, d.labels, d.min_replicas
FROM cron_jobs j
  INNER JOIN deployments d on j.deployment_id = d.id
WHERE j.key = $1::cron_job_key
FOR UPDATE SKIP LOCKED
`

type GetCronJobByKeyRow struct {
	CronJob    CronJob
	Deployment Deployment
}

func (q *Queries) GetCronJobByKey(ctx context.Context, key model.CronJobKey) (GetCronJobByKeyRow, error) {
	row := q.db.QueryRowContext(ctx, getCronJobByKey, key)
	var i GetCronJobByKeyRow
	err := row.Scan(
		&i.CronJob.ID,
		&i.CronJob.Key,
		&i.CronJob.DeploymentID,
		&i.CronJob.Verb,
		&i.CronJob.Schedule,
		&i.CronJob.StartTime,
		&i.CronJob.NextExecution,
		&i.CronJob.ModuleName,
		&i.CronJob.LastExecution,
		&i.Deployment.ID,
		&i.Deployment.CreatedAt,
		&i.Deployment.ModuleID,
		&i.Deployment.Key,
		&i.Deployment.Schema,
		&i.Deployment.Labels,
		&i.Deployment.MinReplicas,
	)
	return i, err
}

const getUnscheduledCronJobs = `-- name: GetUnscheduledCronJobs :many
SELECT j.id, j.key, j.deployment_id, j.verb, j.schedule, j.start_time, j.next_execution, j.module_name, j.last_execution, d.id, d.created_at, d.module_id, d.key, d.schema, d.labels, d.min_replicas
FROM cron_jobs j
  INNER JOIN deployments d on j.deployment_id = d.id
WHERE d.min_replicas > 0
  AND j.start_time < $1::TIMESTAMPTZ
  AND (
    j.last_execution IS NULL
    OR NOT EXISTS (
      SELECT 1
      FROM async_calls ac
      WHERE
        ac.cron_job_key = j.key
        AND ac.scheduled_at > j.last_execution::TIMESTAMPTZ
    )
  )
FOR UPDATE SKIP LOCKED
`

type GetUnscheduledCronJobsRow struct {
	CronJob    CronJob
	Deployment Deployment
}

func (q *Queries) GetUnscheduledCronJobs(ctx context.Context, startTime time.Time) ([]GetUnscheduledCronJobsRow, error) {
	rows, err := q.db.QueryContext(ctx, getUnscheduledCronJobs, startTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetUnscheduledCronJobsRow
	for rows.Next() {
		var i GetUnscheduledCronJobsRow
		if err := rows.Scan(
			&i.CronJob.ID,
			&i.CronJob.Key,
			&i.CronJob.DeploymentID,
			&i.CronJob.Verb,
			&i.CronJob.Schedule,
			&i.CronJob.StartTime,
			&i.CronJob.NextExecution,
			&i.CronJob.ModuleName,
			&i.CronJob.LastExecution,
			&i.Deployment.ID,
			&i.Deployment.CreatedAt,
			&i.Deployment.ModuleID,
			&i.Deployment.Key,
			&i.Deployment.Schema,
			&i.Deployment.Labels,
			&i.Deployment.MinReplicas,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateCronJobExecution = `-- name: UpdateCronJobExecution :exec
UPDATE cron_jobs
  SET last_execution = $1::TIMESTAMPTZ,
    next_execution = $2::TIMESTAMPTZ
  WHERE key = $3::cron_job_key
`

func (q *Queries) UpdateCronJobExecution(ctx context.Context, lastExecution time.Time, nextExecution time.Time, key model.CronJobKey) error {
	_, err := q.db.ExecContext(ctx, updateCronJobExecution, lastExecution, nextExecution, key)
	return err
}
