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

