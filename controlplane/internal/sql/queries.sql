-- name: CreateModule :one
INSERT INTO modules (language, name)
VALUES ($1, $2)
ON CONFLICT (name) DO UPDATE SET language = $1
RETURNING id;

-- name: GetDeploymentsByID :many
SELECT *
FROM deployments
WHERE id = ANY(@ids::BIGINT[]);

-- name: GetModulesByID :many
SELECT *
FROM modules
WHERE id = ANY(@ids::BIGINT[]);

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

-- name: GetLatestDeployment :one
SELECT d.*, m.language, m.name AS module_name
FROM deployments d
         INNER JOIN modules m ON m.id = d.module_id
WHERE m.name = @module_name
ORDER BY created_at DESC
LIMIT 1;

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

-- name: RegisterRunner :one
INSERT
INTO runners (key, language, endpoint)
VALUES ($1, $2, $3)
ON CONFLICT (key) DO UPDATE SET language = $2,
                                endpoint = $3
RETURNING id;

-- name: DeleteStaleRunners :one
WITH deleted AS (
    DELETE FROM runners
        WHERE last_seen < (NOW() AT TIME ZONE 'utc') - $1::INTERVAL
        RETURNING 1)
SELECT COUNT(*)
FROM deleted;


-- name: DeregisterRunner :one
WITH deleted AS (
    DELETE FROM runners WHERE key = $1
        RETURNING 1)
SELECT COUNT(*)
FROM deleted;

-- name: GetIdleRunnersForLanguage :many
SELECT * FROM runners
WHERE language = $1
  AND state = 'idle'
LIMIT $2;

-- name: GetRunnersForModule :many
-- Get all runners that are assigned to run the given module.
SELECT r.*, d.key AS deployment_key, m.id AS module_id, m.name AS module_name
FROM runners r
         JOIN deployments d ON r.deployment_id = d.id
         JOIN modules m ON d.module_id = m.id
WHERE m.name = $1
  AND r.state = 'assigned';

-- name: ReserveRunners :one
-- Find idle runners and reserve them for the given deployment.
UPDATE runners
SET state         = 'reserved',
    deployment_id = COALESCE((SELECT id
                              FROM deployments d
                              WHERE d.key = @deployment_key
                              LIMIT 1), -1)
WHERE id = (SELECT id
            FROM runners r
            WHERE r.language = $1
              AND r.state = 'idle'
            LIMIT $2 FOR UPDATE SKIP LOCKED)
RETURNING runners.*;

-- name: GetIdleRunnerCountsByLanguage :many
SELECT language, COUNT(*) AS count
FROM runners
WHERE state = 'idle'
GROUP BY language
ORDER BY language;

-- name: GetRunnerState :one
SELECT state
FROM runners
WHERE key = $1;

-- name: UpdateRunner :one
UPDATE runners r
SET state         = $2,
    last_seen     = (NOW() AT TIME ZONE 'utc'),
    deployment_id = COALESCE((SELECT id
                              FROM deployments d
                              WHERE d.key = sqlc.narg('deployment_key')
                              LIMIT 1), -1)
WHERE r.key = $1
RETURNING r.deployment_id;

-- name: GetRoutingTable :many
SELECT endpoint
FROM runners r
WHERE state = 'assigned'
  AND r.deployment_id = COALESCE((SELECT d.id
                                  FROM deployments d
                                           INNER JOIN modules m ON d.module_id = m.id
                                  WHERE m.name = $1), -1);

-- name: ExpireRunnerReservations :one
WITH rows AS (
    UPDATE runners
        SET state = 'idle',
            deployment_id = NULL,
            reservation_timeout = NULL
    WHERE state = 'reserved'
        AND reservation_timeout < (NOW() AT TIME ZONE 'utc')
    RETURNING 1
)
SELECT COUNT(*)
FROM rows;


-- name: InsertDeploymentLogEntry :exec
INSERT INTO deployment_logs (deployment_id, time_stamp, level, scope, message, error)
VALUES ((SELECT id FROM deployments WHERE key=$1 LIMIT 1)::UUID, $2, $3, $4, $5, $6);
