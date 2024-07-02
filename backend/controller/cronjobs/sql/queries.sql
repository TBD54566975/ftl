-- name: GetCronJobs :many
SELECT j.key as key, d.key as deployment_key, j.module_name as module, j.verb, j.schedule, j.start_time, j.next_execution, j.state
FROM cron_jobs j
  INNER JOIN deployments d on j.deployment_id = d.id
WHERE d.min_replicas > 0;

-- name: StartCronJobs :many
WITH updates AS (
  UPDATE cron_jobs
  SET state = 'executing',
    start_time = (NOW() AT TIME ZONE 'utc')::TIMESTAMPTZ
  WHERE key = ANY (sqlc.arg('keys'))
    AND state = 'idle'
    AND start_time < next_execution
    AND (next_execution AT TIME ZONE 'utc') < (NOW() AT TIME ZONE 'utc')::TIMESTAMPTZ
  RETURNING id, key, state, start_time, next_execution)
SELECT j.key as key, d.key as deployment_key, j.module_name as module, j.verb, j.schedule,
  COALESCE(u.start_time, j.start_time) as start_time,
  COALESCE(u.next_execution, j.next_execution) as next_execution,
  COALESCE(u.state, j.state) as state,
  d.min_replicas > 0 as has_min_replicas,
  CASE WHEN u.key IS NULL THEN FALSE ELSE TRUE END as updated
FROM cron_jobs j
  INNER JOIN deployments d on j.deployment_id = d.id
  LEFT JOIN updates u on j.id = u.id
WHERE j.key = ANY (sqlc.arg('keys'));

-- name: EndCronJob :one
WITH j AS (
UPDATE cron_jobs
  SET state = 'idle',
    next_execution = sqlc.arg('next_execution')::TIMESTAMPTZ
  WHERE key = sqlc.arg('key')::cron_job_key
    AND state = 'executing'
    AND start_time = sqlc.arg('start_time')::TIMESTAMPTZ
  RETURNING *
)
SELECT j.key as key, d.key as deployment_key, j.module_name as module, j.verb, j.schedule, j.start_time, j.next_execution, j.state
  FROM j
  INNER JOIN deployments d on j.deployment_id = d.id
  LIMIT 1;

-- name: GetStaleCronJobs :many
SELECT j.key as key, d.key as deployment_key, j.module_name as module, j.verb, j.schedule, j.start_time, j.next_execution, j.state
FROM cron_jobs j
  INNER JOIN deployments d on j.deployment_id = d.id
WHERE state = 'executing'
  AND start_time < (NOW() AT TIME ZONE 'utc') - $1::INTERVAL;
