-- name: CreateModule :one
INSERT INTO modules (language, name) VALUES ($1, $2)
ON CONFLICT (name) DO UPDATE SET language = $1
RETURNING id;

-- name: GetDeploymentsByID :many
SELECT * FROM deployments
WHERE id = ANY(@ids::BIGINT[]);

-- name: GetModulesByID :many
SELECT * FROM modules
WHERE id = ANY(@ids::BIGINT[]);

-- name: CreateDeployment :one
INSERT INTO deployments (module_id, "schema")
VALUES ((SELECT id FROM modules WHERE name = @module_name::TEXT LIMIT 1), @schema::BYTEA)
RETURNING key;

-- name: GetArtefactDigests :many
-- Return the digests that exist in the database.
SELECT id, digest FROM artefacts WHERE digest = ANY(@digests::bytea[]);

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
ORDER BY created_at DESC LIMIT 1;

-- name: GetDeploymentsWithArtefacts :many
-- Get all deployments that have artefacts matching the given digests.
SELECT d.id, d.created_at, d.key, m.name
FROM deployments d
INNER JOIN modules m ON d.module_id = m.id
WHERE EXISTS (
  SELECT 1
  FROM deployment_artefacts da
  INNER JOIN artefacts a ON da.artefact_id = a.id
  WHERE a.digest = ANY(@digests::bytea[])
    AND da.deployment_id = d.id
  HAVING COUNT(*) = @count -- Number of unique digests provided
);

-- name: GetArtefactContentRange :one
SELECT SUBSTRING(a.content FROM @start FOR @count)::BYTEA AS content
FROM artefacts a
WHERE a.id = @id;

-- name: RegisterRunner :one
INSERT INTO runners (key, language, endpoint) VALUES ($1, $2, $3)
ON CONFLICT (key) DO UPDATE SET language = $2, endpoint = $3
RETURNING id;

-- name: DeleteStaleRunners :one
WITH deleted AS (
  DELETE FROM runners
  WHERE last_seen < (NOW() AT TIME ZONE 'utc') - $1::INTERVAL
  RETURNING *
)
SELECT COUNT(*) FROM deleted;

-- name: HeartbeatRunner :exec
UPDATE runners SET last_seen = (NOW() AT TIME ZONE 'utc') WHERE id = $1;

-- name: DeregisterRunner :exec
DELETE FROM runners WHERE id = $1;

-- name: GetIdleRunnersForLanguage :many
SELECT * FROM runners
WHERE language = $1
  AND deployment_id IS NULL;

-- name: GetRunnersForModule :many
-- Get all runners that are assigned to run the given module.
SELECT r.*, d.key AS deployment_key, m.id AS module_id, m.name AS module_name
FROM runners r
JOIN deployments d ON r.deployment_id = d.id
JOIN modules m ON d.module_id = m.id
WHERE m.name = $1;

-- name: AssignDeployment :exec
UPDATE runners
SET deployment_id = $2
WHERE id = $1;
