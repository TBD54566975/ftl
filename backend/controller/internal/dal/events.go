package dal

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/types"

	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/model"
	"github.com/TBD54566975/ftl/backend/controller/internal/sql"
	"github.com/TBD54566975/ftl/backend/controller/internal/sqltypes"
	"github.com/TBD54566975/ftl/backend/schema"
)

type EventType = sql.EventType

// Supported event types.
const (
	EventTypeLog  = sql.EventTypeLog
	EventTypeCall = sql.EventTypeCall
)

// Event types.
//
//sumtype:decl
type Event interface{ event() }

type LogEvent struct {
	DeploymentKey model.DeploymentKey
	RequestKey    types.Option[model.IngressRequestKey]
	Time          time.Time
	Level         int32
	Attributes    map[string]string
	Message       string
	Error         types.Option[string]
}

func (e *LogEvent) event() {}

type CallEvent struct {
	DeploymentKey model.DeploymentKey
	RequestKey    types.Option[model.IngressRequestKey]
	Time          time.Time
	SourceVerb    schema.VerbRef
	DestVerb      schema.VerbRef
	Duration      time.Duration
	Request       []byte
	Response      []byte
	Error         error
}

func (e *CallEvent) event() {}

type eventFilterCall struct {
	sourceModule string
	destModule   string
}

type eventFilter struct {
	level       *log.Level
	calls       []*eventFilterCall
	types       []EventType
	deployments []sqltypes.Key
	requests    []sqltypes.Key
}

type EventFilter func(query *eventFilter)

func FilterLogs(level log.Level) EventFilter {
	return func(query *eventFilter) {
		query.level = &level
	}
}

// FilterCall filters call events between the given modules.
//
// May be called multiple times.
func FilterCall(sourceModule, destModule string) EventFilter {
	return func(query *eventFilter) {
		query.calls = append(query.calls, &eventFilterCall{sourceModule: sourceModule, destModule: destModule})
	}
}

func FilterDeployments(deploymentKeys ...model.DeploymentKey) EventFilter {
	return func(query *eventFilter) {
		for _, key := range deploymentKeys {
			query.deployments = append(query.deployments, sqltypes.Key(key))
		}
	}
}

func FilterRequests(requestKeys ...model.IngressRequestKey) EventFilter {
	return func(query *eventFilter) {
		for _, request := range requestKeys {
			query.requests = append(query.requests, sqltypes.Key(request))
		}
	}
}

func FilterTypes(types ...sql.EventType) EventFilter {
	return func(query *eventFilter) {
		query.types = append(query.types, types...)
	}
}

type eventRow struct {
	DeploymentKey sqltypes.Key
	RequestKey    *sqltypes.Key
	TimeStamp     time.Time
	CustomKey1    types.Option[string]
	CustomKey2    types.Option[string]
	CustomKey3    types.Option[string]
	CustomKey4    types.Option[string]
	Type          sql.EventType
	Payload       []byte
}

// The internal JSON payload of a call event.
type eventCallJSON struct {
	DurationMS int64                `json:"duration_ms"`
	Request    json.RawMessage      `json:"request"`
	Response   json.RawMessage      `json:"response"`
	Error      types.Option[string] `json:"error"`
}

type eventLogJSON struct {
	Message    string               `json:"message"`
	Attributes map[string]string    `json:"attributes"`
	Error      types.Option[string] `json:"error"`
}

func (d *DAL) QueryEvents(ctx context.Context, after, before time.Time, filters ...EventFilter) ([]Event, error) {
	// Build query.
	q := `SELECT d.key AS deployment_key,
				   ir.key AS request_key,
				   e.time_stamp AS time_stamp,
				   e.custom_key_1 AS custom_key_1,
				   e.custom_key_2 AS custom_key_2,
				   e.custom_key_3 AS custom_key_3,
				   e.custom_key_4 AS custom_key_4,
				   e.type AS type,
				   e.payload AS payload
			FROM events e
					 INNER JOIN deployments d on e.deployment_id = d.id
					 LEFT JOIN ingress_requests ir on e.request_id = ir.id
			WHERE time_stamp BETWEEN $1::TIMESTAMPTZ and $2::TIMESTAMPTZ
		`

	filter := eventFilter{}
	for _, f := range filters {
		f(&filter)
	}
	args := []any{after, before}
	index := 3
	if filter.deployments != nil {
		q += fmt.Sprintf(` AND d.key = ANY($%d::UUID[])`, index)
		index++
		args = append(args, filter.deployments)
	}
	if filter.requests != nil {
		q += fmt.Sprintf(` AND ir.key = ANY($%d::UUID[])`, index)
		index++
		args = append(args, filter.requests)
	}
	if filter.types != nil {
		// Why are we double casting? Because I hit "cannot find encode plan" and
		// this works around it: https://github.com/jackc/pgx/issues/338#issuecomment-333399112
		q += fmt.Sprintf(` AND e.type = ANY($%d::text[]::event_type[])`, index)
		index++
		args = append(args, filter.types)
	}
	if filter.level != nil {
		q += fmt.Sprintf(" AND (e.type != 'log' OR (e.type = 'log' AND e.custom_key_1::INT >= $%d::INT))\n", index)
		index++
		args = append(args, *filter.level)
	}
	if len(filter.calls) > 0 {
		q += " AND ("
		for i, call := range filter.calls {
			if i > 0 {
				q += " OR "
			}
			q += fmt.Sprintf("(e.type != 'call' OR (e.type = 'call' AND e.custom_key_1 = $%d AND e.custom_key_3 = $%d))", index, index+1)
			index += 2
			args = append(args, call.sourceModule, call.destModule)
		}
		q += ")\n"
	}

	// Issue query.
	rows, err := d.db.Conn().Query(ctx, q, args...)
	if err != nil {
		return nil, errors.WithStack(translatePGError(err))
	}
	defer rows.Close()

	// Translate events to concrete Go types.
	var out []Event
	for rows.Next() {
		row := eventRow{}
		err := rows.Scan(
			&row.DeploymentKey, &row.RequestKey, &row.TimeStamp,
			&row.CustomKey1, &row.CustomKey2, &row.CustomKey3, &row.CustomKey4,
			&row.Type, &row.Payload,
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		var requestKey types.Option[model.IngressRequestKey]
		if row.RequestKey != nil {
			requestKey = types.Some(model.IngressRequestKey(*row.RequestKey))
		}

		switch row.Type {
		case sql.EventTypeLog:
			var jsonPayload eventLogJSON
			if err := json.Unmarshal(row.Payload, &jsonPayload); err != nil {
				return nil, errors.WithStack(err)
			}
			level, err := strconv.ParseInt(row.CustomKey1.MustGet(), 10, 32)
			if err != nil {
				return nil, errors.Wrapf(err, "invalid log level: %q", row.CustomKey1.MustGet())
			}
			out = append(out, &LogEvent{
				DeploymentKey: model.DeploymentKey(row.DeploymentKey),
				RequestKey:    requestKey,
				Time:          row.TimeStamp,
				Level:         int32(level),
				Attributes:    jsonPayload.Attributes,
				Message:       jsonPayload.Message,
				Error:         jsonPayload.Error,
			})

		case sql.EventTypeCall:
			var jsonPayload eventCallJSON
			if err := json.Unmarshal(row.Payload, &jsonPayload); err != nil {
				return nil, errors.WithStack(err)
			}
			var eventError error
			if e, ok := jsonPayload.Error.Get(); ok {
				eventError = errors.New(e)
			}
			out = append(out, &CallEvent{
				DeploymentKey: model.DeploymentKey(row.DeploymentKey),
				RequestKey:    requestKey,
				Time:          row.TimeStamp,
				SourceVerb:    schema.VerbRef{Module: row.CustomKey1.MustGet(), Name: row.CustomKey2.MustGet()},
				DestVerb:      schema.VerbRef{Module: row.CustomKey3.MustGet(), Name: row.CustomKey4.MustGet()},
				Duration:      time.Duration(jsonPayload.DurationMS) * time.Millisecond,
				Request:       jsonPayload.Request,
				Response:      jsonPayload.Response,
				Error:         eventError,
			})

		default:
			panic("unknown event type: " + row.Type)
		}
	}
	return out, nil
}
