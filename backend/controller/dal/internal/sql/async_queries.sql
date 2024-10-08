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

-- name: AsyncCallQueueDepth :one
SELECT count(*)
FROM async_calls
WHERE state = 'pending' AND scheduled_at <= (NOW() AT TIME ZONE 'utc');
