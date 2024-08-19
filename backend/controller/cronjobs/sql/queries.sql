-- name: GetUnscheduledCronJobs :many
SELECT sqlc.embed(j), sqlc.embed(d)
FROM cron_jobs j
  INNER JOIN deployments d on j.deployment_id = d.id
WHERE d.min_replicas > 0
  AND j.start_time < sqlc.arg('start_time')::TIMESTAMPTZ
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
  SET last_execution = sqlc.arg('last_execution')::TIMESTAMPTZ,
    next_execution = sqlc.arg('next_execution')::TIMESTAMPTZ
  WHERE key = sqlc.arg('key')::cron_job_key;