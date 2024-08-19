// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: async_queries.sql

package sql

import (
	"context"
	"encoding/json"
	"time"

	"github.com/TBD54566975/ftl/backend/controller/sql/sqltypes"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/alecthomas/types/optional"
)

const createAsyncCall = `-- name: CreateAsyncCall :one
INSERT INTO async_calls (
  scheduled_at,
  verb,
  origin,
  request,
  remaining_attempts,
  backoff,
  max_backoff,
  catch_verb,
  parent_request_key,
  trace_context,
  cron_job_key
)
VALUES (
  $1::TIMESTAMPTZ,
  $2,
  $3,
  $4,
  $5,
  $6::interval,
  $7::interval,
  $8,
  $9,
  $10::jsonb,
  $11
)
RETURNING id
`

type CreateAsyncCallParams struct {
	ScheduledAt       time.Time
	Verb              schema.RefKey
	Origin            string
	Request           []byte
	RemainingAttempts int32
	Backoff           sqltypes.Duration
	MaxBackoff        sqltypes.Duration
	CatchVerb         optional.Option[schema.RefKey]
	ParentRequestKey  optional.Option[string]
	TraceContext      json.RawMessage
	CronJobKey        optional.Option[model.CronJobKey]
}

func (q *Queries) CreateAsyncCall(ctx context.Context, arg CreateAsyncCallParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, createAsyncCall,
		arg.ScheduledAt,
		arg.Verb,
		arg.Origin,
		arg.Request,
		arg.RemainingAttempts,
		arg.Backoff,
		arg.MaxBackoff,
		arg.CatchVerb,
		arg.ParentRequestKey,
		arg.TraceContext,
		arg.CronJobKey,
	)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const isCronJobPending = `-- name: IsCronJobPending :one
SELECT EXISTS (
    SELECT 1
    FROM async_calls ac
    WHERE ac.cron_job_key = $1::cron_job_key
      AND ac.scheduled_at > $2::TIMESTAMPTZ
      AND ac.state = 'pending'
) AS pending
`

func (q *Queries) IsCronJobPending(ctx context.Context, key model.CronJobKey, startTime time.Time) (bool, error) {
	row := q.db.QueryRowContext(ctx, isCronJobPending, key, startTime)
	var pending bool
	err := row.Scan(&pending)
	return pending, err
}
