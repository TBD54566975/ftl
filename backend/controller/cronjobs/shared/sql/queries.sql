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
