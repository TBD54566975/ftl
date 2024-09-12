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

-- name: GetArtefactDigests :many
-- Return the digests that exist in the database.
SELECT id, digest
FROM artefacts
WHERE digest = ANY (@digests::bytea[]);

-- name: GetDeploymentArtefacts :many
-- Get all artefacts matching the given digests.
SELECT da.created_at, artefact_id AS id, executable, path, digest, executable
FROM deployment_artefacts da
         INNER JOIN artefacts ON artefacts.id = da.artefact_id
WHERE deployment_id = $1;

-- name: CreateArtefact :one
-- Create a new artefact and return the artefact ID.
INSERT INTO artefacts (digest, content)
VALUES ($1, $2)
ON CONFLICT (digest)
DO UPDATE SET digest = EXCLUDED.digest
RETURNING id;

-- name: AssociateArtefactWithDeployment :exec
INSERT INTO deployment_artefacts (deployment_id, artefact_id, executable, path)
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
                       INNER JOIN artefacts a ON da.artefact_id = a.id
              WHERE a.digest = ANY (@digests::bytea[])
                AND da.deployment_id = d.id
                AND d.schema = @schema::BYTEA
              HAVING COUNT(*) = @count::BIGINT -- Number of unique digests provided
);

-- name: GetArtefactContentRange :one
SELECT SUBSTRING(a.content FROM @start FOR @count)::BYTEA AS content
FROM artefacts a
WHERE a.id = @id;

-- name: UpsertRunner :one
-- Upsert a runner and return the deployment ID that it is assigned to, if any.
WITH deployment_rel AS (
    SELECT id FROM deployments d
             WHERE d.key = sqlc.arg('deployment_key')::deployment_key
             LIMIT 1)
INSERT
INTO runners (key, endpoint, state, labels, deployment_id, last_seen)
VALUES ($1,
        $2,
        $3,
        $4,
        (SELECT id FROM deployment_rel),
        NOW() AT TIME ZONE 'utc')
ON CONFLICT (key) DO UPDATE SET endpoint      = $2,
                                state         = $3,
                                labels        = $4,
                                last_seen     = NOW() AT TIME ZONE 'utc'
RETURNING deployment_id;

-- name: KillStaleRunners :one
WITH matches AS (
    UPDATE runners
        SET state = 'dead'
        WHERE state <> 'dead' AND last_seen < (NOW() AT TIME ZONE 'utc') - sqlc.arg('timeout')::INTERVAL
        RETURNING 1)
SELECT COUNT(*)
FROM matches;

-- name: DeregisterRunner :one
WITH matches AS (
    UPDATE runners
        SET state = 'dead'
        WHERE key = sqlc.arg('key')::runner_key
        RETURNING 1)
SELECT COUNT(*)
FROM matches;

-- name: GetActiveRunners :many
SELECT DISTINCT ON (r.key) r.key                                   AS runner_key,
                           r.endpoint,
                           r.state,
                           r.labels,
                           r.last_seen,
                           r.module_name,
                           d.key AS deployment_key
FROM runners r
         INNER JOIN deployments d on d.id = r.deployment_id
WHERE r.state <> 'dead'
ORDER BY r.key;

-- name: GetActiveDeployments :many
SELECT sqlc.embed(d), m.name AS module_name, m.language, COUNT(r.id) AS replicas
FROM deployments d
  JOIN modules m ON d.module_id = m.id
  LEFT JOIN runners r ON d.id = r.deployment_id AND r.state = 'assigned'
WHERE min_replicas > 0
GROUP BY d.id, m.name, m.language;

-- name: GetDeploymentsWithMinReplicas :many
SELECT sqlc.embed(d), m.name AS module_name, m.language
FROM deployments d
  INNER JOIN modules m on d.module_id = m.id
WHERE min_replicas > 0
ORDER BY d.key;

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
         LEFT JOIN runners r on d.id = r.deployment_id AND r.state != 'dead'
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

-- name: GetRunnerState :one
SELECT state
FROM runners
WHERE key = sqlc.arg('key')::runner_key;

-- name: GetRunner :one
SELECT DISTINCT ON (r.key) r.key                                   AS runner_key,
                           r.endpoint,
                           r.state,
                           r.labels,
                           r.last_seen,
                           r.module_name,
                           d.key AS deployment_key
FROM runners r
         INNER JOIN deployments d on d.id = r.deployment_id
WHERE r.key = sqlc.arg('key')::runner_key;

-- name: GetRoutingTable :many
SELECT endpoint, r.key AS runner_key, r.module_name, d.key deployment_key
FROM runners r
         LEFT JOIN deployments d on r.deployment_id = d.id
WHERE state = 'assigned'
  AND (COALESCE(cardinality(sqlc.arg('modules')::TEXT[]), 0) = 0
    OR module_name = ANY (sqlc.arg('modules')::TEXT[]));

-- name: GetRouteForRunner :one
-- Retrieve routing information for a runner.
SELECT endpoint, r.key AS runner_key, r.module_name, d.key deployment_key, r.state
FROM runners r
         LEFT JOIN deployments d on r.deployment_id = d.id
WHERE r.key = sqlc.arg('key')::runner_key;

-- name: GetRunnersForDeployment :many
SELECT *
FROM runners r
         INNER JOIN deployments d on r.deployment_id = d.id
WHERE state = 'assigned'
  AND d.key = sqlc.arg('key')::deployment_key;

-- name: InsertTimelineLogEvent :exec
INSERT INTO timeline (
  deployment_id,
  request_id,
  time_stamp,
  custom_key_1,
  type,
  payload
)
VALUES (
  (SELECT id FROM deployments d WHERE d.key = sqlc.arg('deployment_key')::deployment_key LIMIT 1),
  (
    CASE
      WHEN sqlc.narg('request_key')::TEXT IS NULL THEN NULL
      ELSE (SELECT id FROM requests ir WHERE ir.key = sqlc.narg('request_key')::TEXT LIMIT 1)
    END
  ),
  sqlc.arg('time_stamp')::TIMESTAMPTZ,
  sqlc.arg('level')::INT,
  'log',
  sqlc.arg('payload')
);

-- name: InsertTimelineDeploymentCreatedEvent :exec
INSERT INTO timeline (
  deployment_id,
  type,
  custom_key_1,
  custom_key_2,
  payload
)
VALUES (
  (
    SELECT id
    FROM deployments
    WHERE deployments.key = sqlc.arg('deployment_key')::deployment_key
  ),
  'deployment_created',
  sqlc.arg('language')::TEXT,
  sqlc.arg('module_name')::TEXT,
  sqlc.arg('payload')
);

-- name: InsertTimelineDeploymentUpdatedEvent :exec
INSERT INTO timeline (
  deployment_id,
  type,
  custom_key_1,
  custom_key_2,
  payload
)
VALUES (
  (
    SELECT id
    FROM deployments
    WHERE deployments.key = sqlc.arg('deployment_key')::deployment_key
  ),
  'deployment_updated',
  sqlc.arg('language')::TEXT,
  sqlc.arg('module_name')::TEXT,
  sqlc.arg('payload')
);

-- name: InsertTimelineCallEvent :exec
INSERT INTO timeline (
  deployment_id,
  request_id,
  parent_request_id,
  time_stamp,
  type,
  custom_key_1,
  custom_key_2,
  custom_key_3,
  custom_key_4,
  payload
)
VALUES (
  (SELECT id FROM deployments WHERE deployments.key = sqlc.arg('deployment_key')::deployment_key),
  (CASE
      WHEN sqlc.narg('request_key')::TEXT IS NULL THEN NULL
      ELSE (SELECT id FROM requests ir WHERE ir.key = sqlc.narg('request_key')::TEXT)
    END),
  (CASE
      WHEN sqlc.narg('parent_request_key')::TEXT IS NULL THEN NULL
      ELSE (SELECT id FROM requests ir WHERE ir.key = sqlc.narg('parent_request_key')::TEXT)
    END),
  sqlc.arg('time_stamp')::TIMESTAMPTZ,
  'call',
  sqlc.narg('source_module')::TEXT,
  sqlc.narg('source_verb')::TEXT,
  sqlc.arg('dest_module')::TEXT,
  sqlc.arg('dest_verb')::TEXT,
  sqlc.arg('payload')
);

-- name: DeleteOldTimelineEvents :one
WITH deleted AS (
    DELETE FROM timeline
    WHERE time_stamp < (NOW() AT TIME ZONE 'utc') - sqlc.arg('timeout')::INTERVAL
      AND type = sqlc.arg('type')
    RETURNING 1
)
SELECT COUNT(*)
FROM deleted;

-- name: CreateRequest :exec
INSERT INTO requests (origin, "key", source_addr)
VALUES ($1, $2, $3);

-- name: UpsertController :one
INSERT INTO controller (key, endpoint)
VALUES ($1, $2)
ON CONFLICT (key) DO UPDATE SET state     = 'live',
                                endpoint  = $2,
                                last_seen = NOW() AT TIME ZONE 'utc'
RETURNING id;

-- name: KillStaleControllers :one
-- Mark any controller entries that haven't been updated recently as dead.
WITH matches AS (
    UPDATE controller
        SET state = 'dead'
        WHERE state <> 'dead' AND last_seen < (NOW() AT TIME ZONE 'utc') - sqlc.arg('timeout')::INTERVAL
        RETURNING 1)
SELECT COUNT(*)
FROM matches;

-- name: GetActiveControllers :many
SELECT *
FROM controller c
WHERE c.state <> 'dead'
ORDER BY c.key;

-- name: CreateIngressRoute :exec
INSERT INTO ingress_routes (deployment_id, module, verb, method, path)
VALUES ((SELECT id FROM deployments WHERE key = sqlc.arg('key')::deployment_key LIMIT 1), $2, $3, $4, $5);

-- name: GetIngressRoutes :many
-- Get the runner endpoints corresponding to the given ingress route.
SELECT r.key AS runner_key, d.key AS deployment_key, endpoint, ir.path, ir.module, ir.verb
FROM ingress_routes ir
         INNER JOIN runners r ON ir.deployment_id = r.deployment_id
         INNER JOIN deployments d ON ir.deployment_id = d.id
WHERE r.state = 'assigned'
  AND ir.method = $1;

-- name: GetActiveIngressRoutes :many
SELECT d.key AS deployment_key, ir.module, ir.verb, ir.method, ir.path
FROM ingress_routes ir
         INNER JOIN deployments d ON ir.deployment_id = d.id
WHERE d.min_replicas > 0;


-- name: InsertTimelineEvent :exec
INSERT INTO timeline (deployment_id, request_id, parent_request_id, type,
                    custom_key_1, custom_key_2, custom_key_3, custom_key_4,
                    payload)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id;

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

-- name: UpsertTopic :exec
INSERT INTO topics (key, module_id, name, type)
VALUES (
  sqlc.arg('topic')::topic_key,
  (SELECT id FROM modules WHERE name = sqlc.arg('module')::TEXT LIMIT 1),
  sqlc.arg('name')::TEXT,
  sqlc.arg('event_type')::TEXT
)
ON CONFLICT (name, module_id) DO
UPDATE SET
  type = sqlc.arg('event_type')::TEXT
RETURNING id;

-- name: UpsertSubscription :one
INSERT INTO topic_subscriptions (
  key,
  topic_id,
  module_id,
  deployment_id,
  name)
VALUES (
  sqlc.arg('key')::subscription_key,
  (
    SELECT topics.id as id
    FROM topics
    INNER JOIN modules ON topics.module_id = modules.id
    WHERE modules.name = sqlc.arg('topic_module')::TEXT
      AND topics.name = sqlc.arg('topic_name')::TEXT
  ),
  (SELECT id FROM modules WHERE name = sqlc.arg('module')::TEXT),
  (SELECT id FROM deployments WHERE key = sqlc.arg('deployment')::deployment_key),
  sqlc.arg('name')::TEXT
)
ON CONFLICT (name, module_id) DO
UPDATE SET
  topic_id = excluded.topic_id,
  deployment_id = (SELECT id FROM deployments WHERE key = sqlc.arg('deployment')::deployment_key)
RETURNING
  id,
  CASE
    WHEN xmax = 0 THEN true
    ELSE false
  END AS inserted;

-- name: DeleteSubscriptions :many
DELETE FROM topic_subscriptions
WHERE deployment_id IN (
  SELECT deployments.id
  FROM deployments
  WHERE deployments.key = sqlc.arg('deployment')::deployment_key
)
RETURNING topic_subscriptions.key;

-- name: DeleteSubscribers :many
DELETE FROM topic_subscribers
WHERE deployment_id IN (
  SELECT deployments.id
  FROM deployments
  WHERE deployments.key = sqlc.arg('deployment')::deployment_key
)
RETURNING topic_subscribers.key;

-- name: InsertSubscriber :exec
INSERT INTO topic_subscribers (
  key,
  topic_subscriptions_id,
  deployment_id,
  sink,
  retry_attempts,
  backoff,
  max_backoff,
  catch_verb
)
VALUES (
  sqlc.arg('key')::subscriber_key,
  (
    SELECT topic_subscriptions.id as id
    FROM topic_subscriptions
    INNER JOIN modules ON topic_subscriptions.module_id = modules.id
    WHERE modules.name = sqlc.arg('module')::TEXT
      AND topic_subscriptions.name = sqlc.arg('subscription_name')::TEXT
  ),
  (SELECT id FROM deployments WHERE key = sqlc.arg('deployment')::deployment_key),
  sqlc.arg('sink'),
  sqlc.arg('retry_attempts'),
  sqlc.arg('backoff')::interval,
  sqlc.arg('max_backoff')::interval,
  sqlc.arg('catch_verb')
);

-- name: PublishEventForTopic :exec
INSERT INTO topic_events (
    "key",
    topic_id,
    caller,
    payload,
    request_key,
    trace_context
  )
VALUES (
  sqlc.arg('key')::topic_event_key,
  (
    SELECT topics.id
    FROM topics
    INNER JOIN modules ON topics.module_id = modules.id
    WHERE modules.name = sqlc.arg('module')::TEXT
      AND topics.name = sqlc.arg('topic')::TEXT
  ),
  sqlc.arg('caller')::TEXT,
  sqlc.arg('payload'),
  sqlc.arg('request_key')::TEXT,
  sqlc.arg('trace_context')::jsonb
);

-- name: GetSubscriptionsNeedingUpdate :many
-- Results may not be ready to be scheduled yet due to event consumption delay
-- Sorting ensures that brand new events (that may not be ready for consumption)
-- don't prevent older events from being consumed
-- We also make sure that the subscription belongs to a deployment that has at least one runner
WITH runner_count AS (
  SELECT count(r.deployment_id) as runner_count,
         r.deployment_id as deployment
        FROM runners r WHERE r.state = 'assigned'
        GROUP BY deployment HAVING count(r.deployment_id) > 0
)
SELECT
  subs.key::subscription_key as key,
  curser.key as cursor,
  topics.key::topic_key as topic,
  subs.name
FROM topic_subscriptions subs
JOIN runner_count on subs.deployment_id = runner_count.deployment
LEFT JOIN topics ON subs.topic_id = topics.id
LEFT JOIN topic_events curser ON subs.cursor = curser.id
WHERE subs.cursor IS DISTINCT FROM topics.head
  AND subs.state = 'idle'
ORDER BY curser.created_at
LIMIT 3
FOR UPDATE OF subs SKIP LOCKED;

-- name: GetNextEventForSubscription :one
WITH cursor AS (
  SELECT
    created_at,
    id
  FROM topic_events
  WHERE "key" = sqlc.narg('cursor')::topic_event_key
)
SELECT events."key" as event,
        events.payload,
        events.created_at,
        events.caller,
        events.request_key,
        events.trace_context,
        NOW() - events.created_at >= sqlc.arg('consumption_delay')::interval AS ready
FROM topics
LEFT JOIN topic_events as events ON events.topic_id = topics.id
WHERE topics.key = sqlc.arg('topic')::topic_key
  AND (events.created_at, events.id) > (SELECT COALESCE(MAX(cursor.created_at), '1900-01-01'), COALESCE(MAX(cursor.id), 0) FROM cursor)
ORDER BY events.created_at, events.id
LIMIT 1;

-- name: GetRandomSubscriber :one
SELECT
  subscribers.sink as sink,
  subscribers.retry_attempts as retry_attempts,
  subscribers.backoff as backoff,
  subscribers.max_backoff as max_backoff,
  subscribers.catch_verb as catch_verb
FROM topic_subscribers as subscribers
JOIN topic_subscriptions ON subscribers.topic_subscriptions_id = topic_subscriptions.id
WHERE topic_subscriptions.key = sqlc.arg('key')::subscription_key
ORDER BY RANDOM()
LIMIT 1;

-- name: BeginConsumingTopicEvent :exec
WITH event AS (
  SELECT *
  FROM topic_events
  WHERE "key" = sqlc.arg('event')::topic_event_key
)
UPDATE topic_subscriptions
SET state = 'executing',
    cursor = (SELECT id FROM event)
WHERE key = sqlc.arg('subscription')::subscription_key;

-- name: CompleteEventForSubscription :exec
WITH module AS (
  SELECT id
  FROM modules
  WHERE name = sqlc.arg('module')::TEXT
)
UPDATE topic_subscriptions
SET state = 'idle'
WHERE name = @name::TEXT
      AND module_id = (SELECT id FROM module);

-- name: GetSubscription :one
WITH module AS (
  SELECT id
  FROM modules
  WHERE name = $2::TEXT
)
SELECT *
FROM topic_subscriptions
WHERE name = $1::TEXT
      AND module_id = (SELECT id FROM module);

-- name: SetSubscriptionCursor :exec
WITH event AS (
  SELECT id, created_at, key, topic_id, payload
  FROM topic_events
  WHERE "key" = $2::topic_event_key
)
UPDATE topic_subscriptions
SET cursor = (SELECT id FROM event)
WHERE key = $1::subscription_key;

-- name: GetTopic :one
SELECT *
FROM topics
WHERE id = $1::BIGINT;

-- name: GetTopicEvent :one
SELECT *
FROM topic_events
WHERE id = $1::BIGINT;

-- name: GetOnlyEncryptionKey :one
SELECT key, verify_timeline, verify_async
FROM encryption_keys
WHERE id = 1;

-- name: CreateOnlyEncryptionKey :exec
INSERT INTO encryption_keys (id, key)
VALUES (1, $1);

-- name: UpdateEncryptionVerification :exec
UPDATE encryption_keys
SET verify_timeline = $1,
    verify_async = $2
WHERE id = 1;

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