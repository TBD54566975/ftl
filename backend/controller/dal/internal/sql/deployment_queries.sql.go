// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: deployment_queries.sql

package sql

import (
	"context"

	"github.com/TBD54566975/ftl/backend/controller/encryption/api"
	"github.com/TBD54566975/ftl/internal/model"
)

const insertTimelineDeploymentCreatedEvent = `-- name: InsertTimelineDeploymentCreatedEvent :exec
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
    WHERE deployments.key = $1::deployment_key
  ),
  'deployment_created',
  $2::TEXT,
  $3::TEXT,
  $4
)
`

type InsertTimelineDeploymentCreatedEventParams struct {
	DeploymentKey model.DeploymentKey
	Language      string
	ModuleName    string
	Payload       api.EncryptedTimelineColumn
}

func (q *Queries) InsertTimelineDeploymentCreatedEvent(ctx context.Context, arg InsertTimelineDeploymentCreatedEventParams) error {
	_, err := q.db.ExecContext(ctx, insertTimelineDeploymentCreatedEvent,
		arg.DeploymentKey,
		arg.Language,
		arg.ModuleName,
		arg.Payload,
	)
	return err
}

const insertTimelineDeploymentUpdatedEvent = `-- name: InsertTimelineDeploymentUpdatedEvent :exec
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
    WHERE deployments.key = $1::deployment_key
  ),
  'deployment_updated',
  $2::TEXT,
  $3::TEXT,
  $4
)
`

type InsertTimelineDeploymentUpdatedEventParams struct {
	DeploymentKey model.DeploymentKey
	Language      string
	ModuleName    string
	Payload       api.EncryptedTimelineColumn
}

func (q *Queries) InsertTimelineDeploymentUpdatedEvent(ctx context.Context, arg InsertTimelineDeploymentUpdatedEventParams) error {
	_, err := q.db.ExecContext(ctx, insertTimelineDeploymentUpdatedEvent,
		arg.DeploymentKey,
		arg.Language,
		arg.ModuleName,
		arg.Payload,
	)
	return err
}