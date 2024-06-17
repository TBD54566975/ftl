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
VALUES ($1, $2, $3);

-- name: UnsetModuleConfiguration :exec
DELETE FROM module_configuration
WHERE module = @module AND name = @name;
