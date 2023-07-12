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
INSERT INTO deployments (module_id, "schema", key)
VALUES ((SELECT id FROM modules WHERE name = @module_name::TEXT LIMIT 1), @schema::BYTEA, $1);

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
RETURNING id;

-- name: AssociateArtefactWithDeployment :exec
INSERT INTO deployment_artefacts (deployment_id, artefact_id, executable, path)
VALUES ((SELECT id FROM deployments WHERE key = $1), $2, $3, $4);

-- name: GetDeployment :one
SELECT d.*, m.language, m.name AS module_name
FROM deployments d
         INNER JOIN modules m ON m.id = d.module_id
WHERE d.key = $1;

-- name: GetDeploymentsWithArtefacts :many
-- Get all deployments that have artefacts matching the given digests.
SELECT d.id, d.created_at, d.key, m.name
FROM deployments d
         INNER JOIN modules m ON d.module_id = m.id
WHERE EXISTS (SELECT 1
              FROM deployment_artefacts da
                       INNER JOIN artefacts a ON da.artefact_id = a.id
              WHERE a.digest = ANY (@digests::bytea[])
                AND da.deployment_id = d.id
              HAVING COUNT(*) = @count -- Number of unique digests provided
);

-- name: GetArtefactContentRange :one
SELECT SUBSTRING(a.content FROM @start FOR @count)::BYTEA AS content
FROM artefacts a
WHERE a.id = @id;

-- name: UpsertRunner :one
-- Upsert a runner and return the deployment ID that it is assigned to, if any.
WITH deployment_rel AS (
-- If the deployment key is null, then deployment_rel.id will be null,
-- otherwise we try to retrieve the deployments.id using the key. If
-- there is no corresponding deployment, then the deployment ID is -1
-- and the parent statement will fail due to a foreign key constraint.
    SELECT CASE
               WHEN sqlc.narg('deployment_key')::UUID IS NULL THEN NULL
               ELSE COALESCE((SELECT id
                              FROM deployments d
                              WHERE d.key = sqlc.narg('deployment_key')
                              LIMIT 1), -1) END AS id)
INSERT
INTO runners (key, language, endpoint, state, deployment_id, last_seen)
VALUES ($1, $2, $3, $4, (SELECT id FROM deployment_rel), NOW() AT TIME ZONE 'utc')
ON CONFLICT (key) DO UPDATE SET language      = $2,
                                endpoint      = $3,
                                state         = $4,
                                deployment_id = (SELECT id FROM deployment_rel),
                                last_seen     = NOW() AT TIME ZONE 'utc'
RETURNING deployment_id;

-- name: KillStaleRunners :one
WITH matches AS (
    UPDATE runners
        SET state = 'dead'
        WHERE state <> 'dead' AND last_seen < (NOW() AT TIME ZONE 'utc') - $1::INTERVAL
        RETURNING 1)
SELECT COUNT(*)
FROM matches;


-- name: DeregisterRunner :one
WITH matches AS (
    UPDATE runners
        SET state = 'dead'
        WHERE key = $1
        RETURNING 1)
SELECT COUNT(*)
FROM matches;

-- name: GetActiveRunners :many
SELECT DISTINCT ON (r.key) r.key                                                      AS runner_key,
                           r.language,
                           r.endpoint,
                           r.state,
                           r.last_seen,
                           COALESCE(CASE WHEN r.deployment_id IS NOT NULL THEN d.key END, NULL) AS deployment_key
FROM runners r
         LEFT JOIN deployments d on d.id = r.deployment_id OR r.deployment_id IS NULL
WHERE sqlc.arg('all')::bool = true OR r.state <> 'dead'
ORDER BY r.key;

-- name: GetDeployments :many
SELECT d.id, d.key, d.min_replicas, d.created_at, d.schema, m.name AS module_name, m.language
FROM deployments d
         INNER JOIN modules m on d.module_id = m.id
WHERE sqlc.arg('all')::bool = true OR min_replicas > 0
ORDER BY d.key;

-- name: GetIdleRunnersForLanguage :many
SELECT *
FROM runners
WHERE language = $1
  AND state = 'idle'
LIMIT $2;

-- name: SetDeploymentDesiredReplicas :exec
UPDATE deployments
SET min_replicas = $2
WHERE key = $1
RETURNING 1;


-- name: GetDeploymentsNeedingReconciliation :many
-- Get deployments that have a mismatch between the number of assigned and required replicas.
SELECT d.key                  AS key,
       m.name                 AS module_name,
       m.language             AS language,
       COUNT(r.id)            AS assigned_runners_count,
       d.min_replicas::BIGINT AS required_runners_count
FROM deployments d
         LEFT JOIN runners r ON d.id = r.deployment_id AND r.state <> 'dead'
         JOIN modules m ON d.module_id = m.id
GROUP BY d.key, d.min_replicas, m.name, m.language
HAVING COUNT(r.id) <> d.min_replicas;

-- name: ReserveRunner :one
-- Find an idle runner and reserve it for the given deployment.
UPDATE runners
SET state               = 'reserved',
    reservation_timeout = $2,
    -- If a deployment is not found, then the deployment ID is -1
    -- and the update will fail due to a FK constraint.
    deployment_id       = COALESCE((SELECT id
                                    FROM deployments d
                                    WHERE d.key = @deployment_key
                                    LIMIT 1), -1)
WHERE id = (SELECT id
            FROM runners r
            WHERE r.language = $1
              AND r.state = 'idle'
            LIMIT 1 FOR UPDATE SKIP LOCKED)
RETURNING runners.*;

-- name: GetRunnerState :one
SELECT state
FROM runners
WHERE key = $1;

-- name: GetRoutingTable :many
SELECT endpoint
FROM runners r
         INNER JOIN deployments d on r.deployment_id = d.id
         INNER JOIN modules m on d.module_id = m.id
WHERE state = 'assigned'
  AND m.name = $1;

-- name: GetRunnersForDeployment :many
SELECT r.*
FROM runners r
         INNER JOIN deployments d on r.deployment_id = d.id
WHERE state = 'assigned'
  AND d.key = $1;

-- name: ExpireRunnerReservations :one
WITH rows AS (
    UPDATE runners
        SET state = 'idle',
            deployment_id = NULL,
            reservation_timeout = NULL
        WHERE state = 'reserved'
            AND reservation_timeout < (NOW() AT TIME ZONE 'utc')
        RETURNING 1)
SELECT COUNT(*)
FROM rows;


-- name: InsertDeploymentLogEntry :exec
INSERT INTO deployment_logs (deployment_id, time_stamp, level, scope, message, error)
VALUES ((SELECT id FROM deployments WHERE key = $1 LIMIT 1)::UUID, $2, $3, $4, $5, $6);

-- name: InsertMetricEntry :exec
INSERT INTO metrics (runner_id, start_time, end_time, source_module, source_verb, dest_module, dest_verb, name, type,
                     value)
VALUES ((SELECT id FROM runners WHERE key = $1), $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: GetLatestModuleMetrics :many
SELECT DISTINCT ON (dest_module, dest_verb, source_module, source_verb, name)
       r.key AS runner_key, m.*
FROM runners r, metrics m
WHERE dest_module = ANY(@modules::text[])
ORDER BY dest_module, dest_verb, source_module, source_verb, name, end_time DESC;
