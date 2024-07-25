-- name: GetModuleConfiguration :one
SELECT value
FROM module_configuration
WHERE
  (module IS NULL OR module = @module)
  AND name = @name
ORDER BY module NULLS LAST
LIMIT 1;

-- name: ListModuleConfiguration :many
SELECT *
FROM module_configuration
ORDER BY module, name;

-- name: SetModuleConfiguration :exec
INSERT INTO module_configuration (module, name, value)
VALUES ($1, $2, $3)
ON CONFLICT (module, name) DO UPDATE SET value = $3;

-- name: UnsetModuleConfiguration :exec
DELETE FROM module_configuration
WHERE module = @module AND name = @name;

-- name: GetModuleSecretURL :one
SELECT url
FROM module_secrets
WHERE
  (module IS NULL OR module = @module)
  AND name = @name
ORDER BY module NULLS LAST
LIMIT 1;

-- name: ListModuleSecrets :many
SELECT *
FROM module_secrets
ORDER BY module, name;

-- name: SetModuleSecretURL :exec
INSERT INTO module_secrets (module, name, url)
VALUES ($1, $2, $3)
ON CONFLICT (module, name) DO UPDATE SET url = $3;

-- name: UnsetModuleSecret :exec
DELETE FROM module_secrets
WHERE module = @module AND name = @name;
