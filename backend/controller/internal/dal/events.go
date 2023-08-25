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
	EventTypeLog        = sql.EventTypeLog
	EventTypeCall       = sql.EventTypeCall
	EventTypeDeployment = sql.EventTypeDeployment
)

type DeploymentEventType string

const (
	DeploymentCreated  DeploymentEventType = "created"
	DeploymentUpdated  DeploymentEventType = "updated"
	DeploymentReplaced DeploymentEventType = "replaced"
)

// Event types.
//
//sumtype:decl
type Event interface{ event() }

type LogEvent struct {
	DeploymentName model.DeploymentName
	RequestKey     types.Option[model.IngressRequestKey]
	Time           time.Time
	Level          int32
	Attributes     map[string]string
	Message        string
	Error          types.Option[string]
}

func (e *LogEvent) event() {}

type CallEvent struct {
	DeploymentName model.DeploymentName
	RequestKey     types.Option[model.IngressRequestKey]
	Time           time.Time
	SourceVerb     types.Option[schema.VerbRef]
	DestVerb       schema.VerbRef
	Duration       time.Duration
	Request        []byte
	Response       []byte
	Error          types.Option[string]
}

func (e *CallEvent) event() {}

type DeploymentEvent struct {
	DeploymentName     model.DeploymentName
	Time               time.Time
	Type               DeploymentEventType
	Language           string
	ModuleName         string
	MinReplicas        int
	ReplacedDeployment types.Option[model.DeploymentName]
}

func (e *DeploymentEvent) event() {}

type eventFilterCall struct {
	sourceModule types.Option[string]
	destModule   string
	destVerb     types.Option[string]
}

type eventFilter struct {
	level       *log.Level
	calls       []*eventFilterCall
	types       []EventType
	deployments []model.DeploymentName
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
func FilterCall(sourceModule types.Option[string], destModule string, destVerb types.Option[string]) EventFilter {
	return func(query *eventFilter) {
		query.calls = append(query.calls, &eventFilterCall{sourceModule: sourceModule, destModule: destModule, destVerb: destVerb})
	}
}

func FilterDeployments(deploymentNames ...model.DeploymentName) EventFilter {
	return func(query *eventFilter) {
		for _, name := range deploymentNames {
			query.deployments = append(query.deployments, name)
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

// The internal JSON payload of a call event.
type eventCallJSON struct {
	DurationMS int64                `json:"duration_ms"`
	Request    json.RawMessage      `json:"request"`
	Response   json.RawMessage      `json:"response"`
	Error      types.Option[string] `json:"error,omitempty"`
}

type eventLogJSON struct {
	Message    string               `json:"message"`
	Attributes map[string]string    `json:"attributes"`
	Error      types.Option[string] `json:"error,omitempty"`
}

type eventDeploymentJSON struct {
	MinReplicas        int                                `json:"min_replicas"`
	ReplacedDeployment types.Option[model.DeploymentName] `json:"replaced,omitempty"`
}

func (d *DAL) QueryEvents(ctx context.Context, after, before time.Time, filters ...EventFilter) ([]Event, error) {
	// Build query.
	q := `SELECT d.name AS deployment_name,
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
		q += fmt.Sprintf(` AND d.name = ANY($%d::TEXT[])`, index)
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
			if sourceModule, ok := call.sourceModule.Get(); ok {
				q += fmt.Sprintf("(e.type != 'call' OR (e.type = 'call' AND e.custom_key_1 = $%d AND e.custom_key_3 = $%d))", index, index+1)
				args = append(args, sourceModule, call.destModule)
				index += 2
			} else if destVerb, ok := call.destVerb.Get(); ok {
				q += fmt.Sprintf("(e.type != 'call' OR (e.type = 'call' AND e.custom_key_3 = $%d AND e.custom_key_4 = $%d))", index, index+1)
				args = append(args, call.destModule, destVerb)
				index++
			} else {
				q += fmt.Sprintf("(e.type != 'call' OR (e.type = 'call' AND e.custom_key_3 = $%d))", index)
				args = append(args, call.destModule)
				index++
			}
		}
		q += ")\n"
	}

	q += " ORDER BY time_stamp ASC"

	// Issue query.
	rows, err := d.db.Conn().Query(ctx, q, args...)
	if err != nil {
		return nil, errors.WithStack(translatePGError(err))
	}
	defer rows.Close()

	// Translate events to concrete Go types.
	var out []Event
	for rows.Next() {
		row := sql.GetEventsRow{}
		err := rows.Scan(
			&row.DeploymentName, &row.RequestKey, &row.Event.TimeStamp,
			&row.Event.CustomKey1, &row.Event.CustomKey2, &row.Event.CustomKey3, &row.Event.CustomKey4,
			&row.Event.Type, &row.Event.Payload,
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		var requestKey types.Option[model.IngressRequestKey]
		if key, ok := row.RequestKey.Get(); ok {
			requestKey = types.Some(model.IngressRequestKey(key))
		}
		switch row.Event.Type {
		case sql.EventTypeLog:
			var jsonPayload eventLogJSON
			if err := json.Unmarshal(row.Event.Payload, &jsonPayload); err != nil {
				return nil, errors.WithStack(err)
			}
			level, err := strconv.ParseInt(row.Event.CustomKey1.MustGet(), 10, 32)
			if err != nil {
				return nil, errors.Wrapf(err, "invalid log level: %q", row.Event.CustomKey1.MustGet())
			}
			out = append(out, &LogEvent{
				DeploymentName: row.DeploymentName,
				RequestKey:     requestKey,
				Time:           row.Event.TimeStamp,
				Level:          int32(level),
				Attributes:     jsonPayload.Attributes,
				Message:        jsonPayload.Message,
				Error:          jsonPayload.Error,
			})

		case sql.EventTypeCall:
			var jsonPayload eventCallJSON
			if err := json.Unmarshal(row.Event.Payload, &jsonPayload); err != nil {
				return nil, errors.WithStack(err)
			}
			var sourceVerb types.Option[schema.VerbRef]
			sourceModule, smok := row.Event.CustomKey1.Get()
			sourceName, snok := row.Event.CustomKey2.Get()
			if smok && snok {
				sourceVerb = types.Some(schema.VerbRef{Module: sourceModule, Name: sourceName})
			}
			out = append(out, &CallEvent{
				DeploymentName: row.DeploymentName,
				RequestKey:     requestKey,
				Time:           row.Event.TimeStamp,
				SourceVerb:     sourceVerb,
				DestVerb:       schema.VerbRef{Module: row.Event.CustomKey3.MustGet(), Name: row.Event.CustomKey4.MustGet()},
				Duration:       time.Duration(jsonPayload.DurationMS) * time.Millisecond,
				Request:        jsonPayload.Request,
				Response:       jsonPayload.Response,
				Error:          jsonPayload.Error,
			})
		case sql.EventTypeDeployment:
			var jsonPayload eventDeploymentJSON
			if err := json.Unmarshal(row.Event.Payload, &jsonPayload); err != nil {
				return nil, errors.WithStack(err)
			}
			out = append(out, &DeploymentEvent{
				DeploymentName:     row.DeploymentName,
				Time:               row.Event.TimeStamp,
				Type:               DeploymentEventType(row.Event.CustomKey1.MustGet()),
				Language:           row.Event.CustomKey2.MustGet(),
				ModuleName:         row.Event.CustomKey3.MustGet(),
				MinReplicas:        jsonPayload.MinReplicas,
				ReplacedDeployment: jsonPayload.ReplacedDeployment,
			})
		default:
			panic("unknown event type: " + row.Event.Type)
		}
	}
	return out, nil
}
