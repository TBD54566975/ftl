// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: queries.sql

package sql

import (
	"context"
	"time"

	"github.com/TBD54566975/ftl/backend/controller/sql/sqltypes"
	"github.com/TBD54566975/ftl/internal/encryption"
	"github.com/alecthomas/types/optional"
)

const deleteOldTimelineEvents = `-- name: DeleteOldTimelineEvents :one
WITH deleted AS (
    DELETE FROM timeline
    WHERE time_stamp < (NOW() AT TIME ZONE 'utc') - $1::INTERVAL
      AND type = $2
    RETURNING 1
)
SELECT COUNT(*)
FROM deleted
`

func (q *Queries) DeleteOldTimelineEvents(ctx context.Context, timeout sqltypes.Duration, type_ EventType) (int64, error) {
	row := q.db.QueryRowContext(ctx, deleteOldTimelineEvents, timeout, type_)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const dummyQueryTimeline = `-- name: DummyQueryTimeline :one
SELECT id, time_stamp, deployment_id, request_id, type, custom_key_1, custom_key_2, custom_key_3, custom_key_4, payload, parent_request_id FROM timeline WHERE id = $1
`

// This is a dummy query to ensure that the Timeline model is generated.
func (q *Queries) DummyQueryTimeline(ctx context.Context, id int64) (Timeline, error) {
	row := q.db.QueryRowContext(ctx, dummyQueryTimeline, id)
	var i Timeline
	err := row.Scan(
		&i.ID,
		&i.TimeStamp,
		&i.DeploymentID,
		&i.RequestID,
		&i.Type,
		&i.CustomKey1,
		&i.CustomKey2,
		&i.CustomKey3,
		&i.CustomKey4,
		&i.Payload,
		&i.ParentRequestID,
	)
	return i, err
}

const insertTimelineCallEvent = `-- name: InsertTimelineCallEvent :exec
INSERT INTO timeline (
  deployment_id,
  request_id,
  parent_request_id,
  time_stamp,
  type,
  custom_key_1,
  custom_key_2,
  custom_key_3,
  custom_key_4,
  payload
)
VALUES (
  (SELECT id FROM deployments WHERE deployments.key = $1::deployment_key),
  (CASE
      WHEN $2::TEXT IS NULL THEN NULL
      ELSE (SELECT id FROM requests ir WHERE ir.key = $2::TEXT)
    END),
  (CASE
      WHEN $3::TEXT IS NULL THEN NULL
      ELSE (SELECT id FROM requests ir WHERE ir.key = $3::TEXT)
    END),
  $4::TIMESTAMPTZ,
  'call',
  $5::TEXT,
  $6::TEXT,
  $7::TEXT,
  $8::TEXT,
  $9
)
`

type InsertTimelineCallEventParams struct {
	DeploymentKey    interface{}
	RequestKey       optional.Option[string]
	ParentRequestKey optional.Option[string]
	TimeStamp        time.Time
	SourceModule     optional.Option[string]
	SourceVerb       optional.Option[string]
	DestModule       string
	DestVerb         string
	Payload          encryption.EncryptedTimelineColumn
}

func (q *Queries) InsertTimelineCallEvent(ctx context.Context, arg InsertTimelineCallEventParams) error {
	_, err := q.db.ExecContext(ctx, insertTimelineCallEvent,
		arg.DeploymentKey,
		arg.RequestKey,
		arg.ParentRequestKey,
		arg.TimeStamp,
		arg.SourceModule,
		arg.SourceVerb,
		arg.DestModule,
		arg.DestVerb,
		arg.Payload,
	)
	return err
}

const insertTimelineLogEvent = `-- name: InsertTimelineLogEvent :exec
INSERT INTO timeline (
  deployment_id,
  request_id,
  time_stamp,
  custom_key_1,
  type,
  payload
)
VALUES (
  (SELECT id FROM deployments d WHERE d.key = $1::deployment_key LIMIT 1),
  (
    CASE
      WHEN $2::TEXT IS NULL THEN NULL
      ELSE (SELECT id FROM requests ir WHERE ir.key = $2::TEXT LIMIT 1)
    END
  ),
  $3::TIMESTAMPTZ,
  $4::INT,
  'log',
  $5
)
`

type InsertTimelineLogEventParams struct {
	DeploymentKey interface{}
	RequestKey    optional.Option[string]
	TimeStamp     time.Time
	Level         int32
	Payload       encryption.EncryptedTimelineColumn
}

func (q *Queries) InsertTimelineLogEvent(ctx context.Context, arg InsertTimelineLogEventParams) error {
	_, err := q.db.ExecContext(ctx, insertTimelineLogEvent,
		arg.DeploymentKey,
		arg.RequestKey,
		arg.TimeStamp,
		arg.Level,
		arg.Payload,
	)
	return err
}
