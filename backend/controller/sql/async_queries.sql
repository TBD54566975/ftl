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
  trace_context
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
  @trace_context::jsonb
)
RETURNING id;

-- name: IsCronJobPending :one
SELECT EXISTS (
    SELECT 1
    FROM cron_jobs j
      INNER JOIN async_calls ac on j.last_async_call_id = ac.id
    WHERE j.key = sqlc.arg('key')::cron_job_key
      AND ac.scheduled_at > sqlc.arg('start_time')::TIMESTAMPTZ
      AND ac.state = 'pending'
) AS pending;
