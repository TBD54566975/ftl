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
VALUES ((SELECT id FROM modules WHERE name = @module_name::TEXT LIMIT 1), @schema::BYTEA, @key::deployment_key);

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
                       INNER JOIN artefacts a ON da.artefact_id = a.id
              WHERE a.digest = ANY (@digests::bytea[])
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
ORDER BY d.created_at;

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
SET min_replicas = $2
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

-- name: CreateIngressRoute :exec
INSERT INTO ingress_routes (deployment_id, module, verb, method, path)
VALUES ((SELECT id FROM deployments WHERE key = sqlc.arg('key')::deployment_key LIMIT 1), $2, $3, $4, $5);

-- name: GetIngressRoutes :many
-- Get the runner endpoints corresponding to the given ingress route.
SELECT r.key AS runner_key, d.key AS deployment_key, endpoint, ir.path, ir.module, ir.verb, ir.method
FROM ingress_routes ir
         INNER JOIN runners r ON ir.deployment_id = r.deployment_id
         INNER JOIN deployments d ON ir.deployment_id = d.id;

-- name: GetActiveIngressRoutes :many
SELECT d.key AS deployment_key, ir.module, ir.verb, ir.method, ir.path
FROM ingress_routes ir
         INNER JOIN deployments d ON ir.deployment_id = d.id
WHERE d.min_replicas > 0;

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

-- name: GetFSMInstance :one
SELECT *
FROM fsm_instances
WHERE fsm = @fsm::schema_ref AND key = @key;

-- name: StartFSMTransition :one
-- Start a new FSM transition, populating the destination state and async call ID.
--
-- "key" is the unique identifier for the FSM execution.
INSERT INTO fsm_instances (
  fsm,
  key,
  destination_state,
  async_call_id
) VALUES (
  @fsm,
  @key,
  @destination_state::schema_ref,
  @async_call_id::BIGINT
)
ON CONFLICT(fsm, key) DO
UPDATE SET
  destination_state = @destination_state::schema_ref,
  async_call_id = @async_call_id::BIGINT,
  updated_at = NOW() AT TIME ZONE 'utc'
WHERE
  fsm_instances.async_call_id IS NULL
  AND fsm_instances.destination_state IS NULL
RETURNING *;

-- name: FinishFSMTransition :one
-- Mark an FSM transition as completed, updating the current state and clearing the async call ID.
UPDATE fsm_instances
SET
  current_state = destination_state,
  destination_state = NULL,
  async_call_id = NULL,
  updated_at = NOW() AT TIME ZONE 'utc'
WHERE
  fsm = @fsm::schema_ref AND key = @key::TEXT
RETURNING true;

-- name: SucceedFSMInstance :one
UPDATE fsm_instances
SET
  current_state = destination_state,
  destination_state = NULL,
  async_call_id = NULL,
  status = 'completed'::fsm_status,
  updated_at = NOW() AT TIME ZONE 'utc'
WHERE
  fsm = @fsm::schema_ref AND key = @key::TEXT
RETURNING true;

-- name: FailFSMInstance :one
UPDATE fsm_instances
SET
  current_state = NULL,
  async_call_id = NULL,
  status = 'failed'::fsm_status,
  updated_at = NOW() AT TIME ZONE 'utc'
WHERE
  fsm = @fsm::schema_ref AND key = @key::TEXT
RETURNING true;

-- name: SetNextFSMEvent :one
INSERT INTO fsm_next_event (fsm_instance_id, next_state, request, request_type)
VALUES (
  (SELECT id FROM fsm_instances WHERE fsm = @fsm::schema_ref AND key = @instance_key),
  @event,
  @request,
  sqlc.arg('request_type')::schema_type
)
RETURNING id;

-- name: PopNextFSMEvent :one
DELETE FROM fsm_next_event
WHERE fsm_instance_id = (
  SELECT id
  FROM fsm_instances
  WHERE fsm = @fsm::schema_ref AND key = @instance_key
)
RETURNING *;

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
