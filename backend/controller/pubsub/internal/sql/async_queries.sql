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


-- name: SucceedAsyncCall :one
UPDATE async_calls
SET
  state = 'success'::async_call_state,
  response = @response,
  error = null
WHERE id = @id
RETURNING true;

-- name: FailAsyncCall :one
UPDATE async_calls
SET
  state = 'error'::async_call_state,
  error = @error::TEXT
WHERE id = @id
RETURNING true;

-- name: FailAsyncCallWithRetry :one
WITH updated AS (
  UPDATE async_calls
  SET state = 'error'::async_call_state,
      error = @error::TEXT
  WHERE id = @id::BIGINT
  RETURNING *
)
INSERT INTO async_calls (
  verb,
  origin,
  request,
  catch_verb,
  remaining_attempts,
  backoff,
  max_backoff,
  scheduled_at,
  catching,
  error
)
SELECT
  updated.verb,
  updated.origin,
  updated.request,
  updated.catch_verb,
  @remaining_attempts,
  @backoff::interval,
  @max_backoff::interval,
  @scheduled_at::TIMESTAMPTZ,
  @catching::bool,
  @original_error
FROM updated
RETURNING true;

-- name: LoadAsyncCall :one
SELECT *
FROM async_calls
WHERE id = @id;


-- name: AcquireAsyncCall :one
-- Reserve a pending async call for execution, returning the associated lease
-- reservation key and accompanying metadata.
WITH pending_calls AS (
  SELECT id
  FROM async_calls
  WHERE state = 'pending' AND scheduled_at <= (NOW() AT TIME ZONE 'utc')
  ORDER BY created_at
), async_call AS (
  SELECT id
  FROM pending_calls
  LIMIT 1
  FOR UPDATE SKIP LOCKED
)
UPDATE async_calls
SET state = 'executing'
WHERE id = (SELECT id FROM async_call)
RETURNING
  id AS async_call_id,
  (SELECT count(*) FROM pending_calls) AS queue_depth,
  origin,
  verb,
  catch_verb,
  request,
  scheduled_at,
  remaining_attempts,
  error,
  backoff,
  max_backoff,
  parent_request_key,
  trace_context,
  catching;


-- name: GetZombieAsyncCalls :many
SELECT *
FROM async_calls
WHERE state = 'executing'
ORDER BY created_at ASC
LIMIT sqlc.arg('limit')::INT;
