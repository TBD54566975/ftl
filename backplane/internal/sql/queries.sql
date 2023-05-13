-- name: ListModules :many
SELECT * FROM modules;

-- name: CreateModule :one
INSERT INTO modules (language, name) VALUES ($1, $2)
ON CONFLICT (name) DO UPDATE SET language = $1
RETURNING id;

-- name: ListDeployments :many
SELECT * FROM deployments WHERE module_id = $1;

-- name: CreateDeployment :one
INSERT INTO deployments (module_id, "schema")
VALUES ((SELECT id FROM modules WHERE name = @module_name::TEXT LIMIT 1), @schema::BYTEA)
RETURNING key;

-- name: GetArtefactDigests :many
-- Return the digests that exist in the database.
SELECT id, digest FROM artefacts WHERE digest = ANY(@digests::bytea[]);

-- name: GetDeploymentArtefacts :many
-- Get all artefacts matching the given digests.
SELECT deployment_artefacts.created_at, executable, path, digest, executable, content
FROM deployment_artefacts
INNER JOIN artefacts ON artefacts.id = deployment_artefacts.artefact_id
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
SELECT deployments.*, modules.language, modules.name AS module_name FROM deployments
INNER JOIN modules ON modules.id = deployments.module_id
WHERE deployments.key = $1;

-- name: GetLatestDeployment :one
SELECT deployments.*, modules.language, modules.name AS module_name FROM deployments
INNER JOIN modules ON modules.id = deployments.module_id
WHERE modules.name = @module_name
ORDER BY created_at DESC LIMIT 1;

-- name: GetDeploymentsWithArtefacts :many
SELECT d.id, d.created_at, d.key, m.name
FROM deployments d
JOIN modules m ON d.module_id = m.id
WHERE EXISTS (
  SELECT 1
  FROM deployment_artefacts da
  JOIN artefacts a ON da.artefact_id = a.id
  WHERE a.digest = ANY(@digests::bytea[])
    AND da.deployment_id = d.id
  HAVING COUNT(*) = @count -- Number of unique digests provided
);