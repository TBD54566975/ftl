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
RETURNING id;

-- name: GetArtefactDigests :many
-- Return the digests that exist in the database.
SELECT digest FROM artefact_contents WHERE digest = ANY(@digests::text[]);

-- name: GetDeploymentArtefacts :many
-- Get all artefacts matching the given digests.
SELECT created_at, executable, path, content
FROM artefacts
INNER JOIN deployment_artefacts ON artefacts.id = deployment_artefacts.artefact_id AND deployment_artefacts.deployment_id = $1
INNER JOIN artefact_contents ON artefacts.id = artefact_contents.artefact_id
WHERE deployment_id = $1;

-- name: CreateArtefact :one
-- Create a new artefact and return the artefact ID.
WITH new_artefact AS (
  INSERT INTO artefacts (executable, path)
  VALUES ($1, $2)
  RETURNING id AS artefact_id
)
INSERT INTO artefact_contents (artefact_id, digest, content)
VALUES ((SELECT artefact_id FROM new_artefact), $3, $4)
RETURNING artefact_id;

-- name: AssociateArtefactWithDeployment :exec
INSERT INTO deployment_artefacts (deployment_id, artefact_id)
VALUES ($1, $2);

-- name: GetLatestDeployment :one
SELECT deployments.*, modules.language, modules.name AS module_name FROM deployments
INNER JOIN modules ON modules.id = deployments.module_id
WHERE modules.name = @module_name
ORDER BY created_at DESC LIMIT 1;

-- name: GetArtefactsForDeployment :many
SELECT * FROM artefacts
INNER JOIN deployment_artefacts ON artefacts.id = deployment_artefacts.artefact_id
INNER JOIN artefact_contents ON artefacts.id = artefact_contents.artefact_id
WHERE deployment_id = $1;