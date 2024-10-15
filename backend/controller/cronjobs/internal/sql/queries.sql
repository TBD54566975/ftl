-- name: GetUnscheduledCronJobs :many
SELECT sqlc.embed(j), sqlc.embed(d)
FROM cron_jobs j
  INNER JOIN deployments d on j.deployment_id = d.id
WHERE d.min_replicas > 0
  AND j.start_time < sqlc.arg('start_time')::TIMESTAMPTZ
  AND (
    j.last_async_call_id IS NULL
    OR NOT EXISTS (
      SELECT 1
      FROM async_calls ac
      WHERE ac.id = j.last_async_call_id
        AND ac.state IN ('pending', 'executing')
    )
  )
FOR UPDATE SKIP LOCKED;

-- name: GetCronJobByKey :one
SELECT sqlc.embed(j), sqlc.embed(d)
FROM cron_jobs j
  INNER JOIN deployments d on j.deployment_id = d.id
WHERE j.key = sqlc.arg('key')::cron_job_key
FOR UPDATE SKIP LOCKED;

-- name: CreateCronJob :exec
INSERT INTO cron_jobs (key, deployment_id, module_name, verb, schedule, start_time, next_execution)
  VALUES (
    sqlc.arg('key')::cron_job_key,
    (SELECT id FROM deployments WHERE key = sqlc.arg('deployment_key')::deployment_key LIMIT 1),
    sqlc.arg('module_name')::TEXT,
    sqlc.arg('verb')::TEXT,
    sqlc.arg('schedule')::TEXT,
    sqlc.arg('start_time')::TIMESTAMPTZ,
    sqlc.arg('next_execution')::TIMESTAMPTZ);

-- name: UpdateCronJobExecution :exec
UPDATE cron_jobs
  SET last_async_call_id = sqlc.arg('last_async_call_id')::BIGINT,
    last_execution = sqlc.arg('last_execution')::TIMESTAMPTZ,
    next_execution = sqlc.arg('next_execution')::TIMESTAMPTZ
  WHERE key = sqlc.arg('key')::cron_job_key;

-- name: IsCronJobPending :one
SELECT EXISTS (
    SELECT 1
    FROM cron_jobs j
      INNER JOIN async_calls ac on j.last_async_call_id = ac.id
    WHERE j.key = sqlc.arg('key')::cron_job_key
      AND ac.scheduled_at > sqlc.arg('start_time')::TIMESTAMPTZ
      AND ac.state = 'pending'
) AS pending;

-- name: DeleteCronJobsForDeployment :exec
DELETE FROM cron_jobs
WHERE deployment_id = (SELECT id FROM deployments WHERE key = sqlc.arg('deployment_key')::deployment_key LIMIT 1);