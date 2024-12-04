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

-- name: GetArtefactDigests :many
-- Return the digests that exist in the database.
SELECT DISTINCT digest
FROM deployment_artefacts
WHERE digest = ANY (@digests::bytea[]);

-- name: GetDeploymentArtefacts :many
-- Get all artefacts matching the given digests.
SELECT da.created_at, executable, path, digest, executable
FROM deployment_artefacts da
WHERE deployment_id = $1;

-- name: AssociateArtefactWithDeployment :exec
INSERT INTO deployment_artefacts (deployment_id, digest, executable, path)
VALUES ((SELECT id FROM deployments WHERE key = @key::deployment_key), $2, $3, $4);

-- name: GetDeployment :one
SELECT sqlc.embed(d), m.language, m.name AS module_name, d.min_replicas
FROM deployments d
         INNER JOIN modules m ON m.id = d.module_id
WHERE d.key = sqlc.arg('key')::deployment_key;

-- name: GetDeploymentsWithArtefacts :many
-- Get all deployments that have artefacts matching the given digests.
SELECT d.id, d.created_at, d.key as deployment_key, d.schema, m.name AS module_name
FROM deployments d
         INNER JOIN modules m ON d.module_id = m.id
WHERE EXISTS (SELECT 1
              FROM deployment_artefacts da
              WHERE da.digest = ANY (@digests::bytea[])
                AND da.deployment_id = d.id
                AND d.schema = @schema::BYTEA
              HAVING COUNT(*) = @count::BIGINT -- Number of unique digests provided
);

-- name: UpsertRunner :one
-- Upsert a runner and return the deployment ID that it is assigned to, if any.
WITH deployment_rel AS (
    SELECT id FROM deployments d
             WHERE d.key = sqlc.arg('deployment_key')::deployment_key
             LIMIT 1)
INSERT
INTO runners (key, endpoint, labels, deployment_id, last_seen)
VALUES ($1,
        $2,
        $3,
        (SELECT id FROM deployment_rel),
        NOW() AT TIME ZONE 'utc')
ON CONFLICT (key) DO UPDATE SET endpoint      = $2,
                                labels        = $3,
                                last_seen     = NOW() AT TIME ZONE 'utc'
RETURNING deployment_id;

-- name: KillStaleRunners :one
WITH matches AS (
    DELETE FROM runners
        WHERE last_seen < (NOW() AT TIME ZONE 'utc') - sqlc.arg('timeout')::INTERVAL
        RETURNING 1)
SELECT COUNT(*)
FROM matches;

-- name: DeregisterRunner :one
WITH matches AS (
    DELETE FROM runners
        WHERE key = sqlc.arg('key')::runner_key
        RETURNING 1)
SELECT COUNT(*)
FROM matches;

-- name: GetActiveRunners :many
SELECT DISTINCT ON (r.key) r.key AS runner_key,
                           r.endpoint,
                           r.labels,
                           r.last_seen,
                           r.module_name,
                           d.key AS deployment_key
FROM runners r
         INNER JOIN deployments d on d.id = r.deployment_id
ORDER BY r.key;

-- name: GetActiveDeployments :many
SELECT sqlc.embed(d), m.name AS module_name, m.language, COUNT(r.id) AS replicas
FROM deployments d
  JOIN modules m ON d.module_id = m.id
  LEFT JOIN runners r ON d.id = r.deployment_id
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

-- name: GetProcessList :many
SELECT d.min_replicas,
       d.key   AS deployment_key,
       d.labels    deployment_labels,
       r.key    AS runner_key,
       r.endpoint,
       r.labels AS runner_labels
FROM deployments d
         LEFT JOIN runners r on d.id = r.deployment_id
WHERE d.min_replicas > 0
ORDER BY d.key;

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

-- name: GetRunner :one
SELECT DISTINCT ON (r.key) r.key                                   AS runner_key,
                           r.endpoint,
                           r.labels,
                           r.last_seen,
                           r.module_name,
                           d.key AS deployment_key
FROM runners r
         INNER JOIN deployments d on d.id = r.deployment_id
WHERE r.key = sqlc.arg('key')::runner_key;

-- name: GetRunnersForDeployment :many
SELECT *
FROM runners r
         INNER JOIN deployments d on r.deployment_id = d.id
WHERE d.key = sqlc.arg('key')::deployment_key;

-- name: CreateRequest :exec
INSERT INTO requests (origin, "key", source_addr)
VALUES ($1, $2, $3);

-- name: UpsertController :one
INSERT INTO controllers (key, endpoint)
VALUES ($1, $2)
ON CONFLICT (key) DO UPDATE SET state     = 'live',
                                endpoint  = $2,
                                last_seen = NOW() AT TIME ZONE 'utc'
RETURNING id;

-- name: KillStaleControllers :one
-- Mark any controller entries that haven't been updated recently as dead.
WITH matches AS (
    UPDATE controllers
        SET state = 'dead'
        WHERE state <> 'dead' AND last_seen < (NOW() AT TIME ZONE 'utc') - sqlc.arg('timeout')::INTERVAL
        RETURNING 1)
SELECT COUNT(*)
FROM matches;

-- name: GetActiveControllers :many
SELECT *
FROM controllers c
WHERE c.state <> 'dead'
ORDER BY c.key;

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
  AND lease_id IS NULL
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
), lease AS (
  INSERT INTO leases (idempotency_key, key, expires_at)
  SELECT gen_random_uuid(), '/system/async_call/' || (SELECT id FROM async_call), (NOW() AT TIME ZONE 'utc') + @ttl::interval
  WHERE (SELECT id FROM async_call) IS NOT NULL
  RETURNING *
)
UPDATE async_calls
SET state = 'executing', lease_id = (SELECT id FROM lease)
WHERE id = (SELECT id FROM async_call)
RETURNING
  id AS async_call_id,
  (SELECT idempotency_key FROM lease) AS lease_idempotency_key,
  (SELECT key FROM lease) AS lease_key,
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
