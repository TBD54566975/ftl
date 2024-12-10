-- name: UpsertModule :one
INSERT INTO modules (language, name)
VALUES ($1, $2)
ON CONFLICT (name) DO UPDATE SET language = $1
RETURNING id;

-- name: GetDeploymentsByID :many
SELECT *
FROM deployments
WHERE id = ANY (@ids::BIGINT[]);

-- name: GetModulesByID :many
SELECT *
FROM modules
WHERE id = ANY (@ids::BIGINT[]);

-- name: CreateDeployment :exec
INSERT INTO deployments (module_id, "schema", "key")
VALUES ((SELECT id FROM modules WHERE name = @module_name::TEXT LIMIT 1), @schema::module_schema_pb, @key::deployment_key);

-- Note that this can result in a race condition if the deployment is being updated by another process. This will go
-- away once we ditch the DB.
--
-- name: UpdateDeploymentSchema :exec
UPDATE deployments
SET schema = @schema::module_schema_pb
WHERE key = @key::deployment_key
RETURNING 1;

-- name: GetDeployment :one
SELECT sqlc.embed(d), m.language, m.name AS module_name, d.min_replicas
FROM deployments d
         INNER JOIN modules m ON m.id = d.module_id
WHERE d.key = sqlc.arg('key')::deployment_key;

-- name: GetActiveDeployments :many
SELECT sqlc.embed(d), m.name AS module_name, m.language
FROM deployments d
  JOIN modules m ON d.module_id = m.id
WHERE min_replicas > 0
GROUP BY d.id, m.name, m.language
ORDER BY d.last_activated_at;

-- name: GetDeploymentsWithMinReplicas :many
SELECT sqlc.embed(d), m.name AS module_name, m.language
FROM deployments d
  INNER JOIN modules m on d.module_id = m.id
WHERE min_replicas > 0
ORDER BY d.created_at,d.key;

-- name: GetActiveDeploymentSchemas :many
SELECT key, schema FROM deployments WHERE min_replicas > 0;

-- name: GetSchemaForDeployment :one
SELECT schema FROM deployments WHERE key = sqlc.arg('key')::deployment_key;

-- name: SetDeploymentDesiredReplicas :exec
UPDATE deployments
SET min_replicas = $2, last_activated_at = CASE WHEN min_replicas = 0 THEN (NOW() AT TIME ZONE 'utc') ELSE  last_activated_at END
WHERE key = sqlc.arg('key')::deployment_key
RETURNING 1;

-- name: GetExistingDeploymentForModule :one
SELECT *
FROM deployments d
         INNER JOIN modules m on d.module_id = m.id
WHERE m.name = $1
  AND min_replicas > 0
LIMIT 1;

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

-- name: GetZombieAsyncCalls :many
SELECT *
FROM async_calls
WHERE state = 'executing'
ORDER BY created_at ASC
LIMIT sqlc.arg('limit')::INT;

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
