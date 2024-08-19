-- name: CreateAsyncCall :one
INSERT INTO async_calls (
  scheduled_at,
  verb,
  origin,
  request,
  remaining_attempts,
  backoff,
  max_backoff,
  catch_verb,
  parent_request_key,
  trace_context,
  cron_job_key
)
VALUES (
  @scheduled_at::TIMESTAMPTZ,
  @verb,
  @origin,
  @request,
  @remaining_attempts,
  @backoff::interval,
  @max_backoff::interval,
  @catch_verb,
  @parent_request_key,
  @trace_context::jsonb,
  @cron_job_key
)
RETURNING id;

-- name: IsCronJobPending :one
SELECT EXISTS (
    SELECT 1
    FROM async_calls ac
    WHERE ac.cron_job_key = sqlc.arg('key')::cron_job_key
      AND ac.scheduled_at > sqlc.arg('start_time')::TIMESTAMPTZ
      AND ac.state = 'pending'
) AS pending;
