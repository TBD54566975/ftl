-- name: UpsertModule :one
INSERT INTO modules (language, name)
VALUES ($1, $2)
ON CONFLICT (name) DO UPDATE SET language = $1
RETURNING id;

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

