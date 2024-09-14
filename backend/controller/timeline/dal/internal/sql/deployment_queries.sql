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
