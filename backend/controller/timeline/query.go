package timeline

import (
	"context"
	stdsql "database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/TBD54566975/ftl/backend/controller/timeline/internal/sql"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/alecthomas/types/optional"
)

type eventFilterCall struct {
	sourceModule optional.Option[string]
	destModule   string
	destVerb     optional.Option[string]
}

type eventFilter struct {
	level        *log.Level
	calls        []*eventFilterCall
	types        []EventType
	deployments  []model.DeploymentKey
	requests     []string
	newerThan    time.Time
	olderThan    time.Time
	idHigherThan int64
	idLowerThan  int64
	descending   bool
}

type eventRow struct {
	sql.Timeline
	DeploymentKey model.DeploymentKey
	RequestKey    optional.Option[model.RequestKey]
}

type TimelineFilter func(query *eventFilter)

func FilterLogLevel(level log.Level) TimelineFilter {
	return func(query *eventFilter) {
		query.level = &level
	}
}

// FilterCall filters call events between the given modules.
//
// May be called multiple times.
func FilterCall(sourceModule optional.Option[string], destModule string, destVerb optional.Option[string]) TimelineFilter {
	return func(query *eventFilter) {
		query.calls = append(query.calls, &eventFilterCall{sourceModule: sourceModule, destModule: destModule, destVerb: destVerb})
	}
}

func FilterDeployments(deploymentKeys ...model.DeploymentKey) TimelineFilter {
	return func(query *eventFilter) {
		query.deployments = append(query.deployments, deploymentKeys...)
	}
}

func FilterRequests(requestKeys ...model.RequestKey) TimelineFilter {
	return func(query *eventFilter) {
		for _, request := range requestKeys {
			query.requests = append(query.requests, request.String())
		}
	}
}

func FilterTypes(types ...sql.EventType) TimelineFilter {
	return func(query *eventFilter) {
		query.types = append(query.types, types...)
	}
}

// FilterTimeRange filters events between the given times, inclusive.
//
// Either maybe be zero to indicate no upper or lower bound.
func FilterTimeRange(olderThan, newerThan time.Time) TimelineFilter {
	return func(query *eventFilter) {
		query.newerThan = newerThan
		query.olderThan = olderThan
	}
}

// FilterIDRange filters events between the given IDs, inclusive.
func FilterIDRange(higherThan, lowerThan int64) TimelineFilter {
	return func(query *eventFilter) {
		query.idHigherThan = higherThan
		query.idLowerThan = lowerThan
	}
}

// FilterDescending returns events in descending order.
func FilterDescending() TimelineFilter {
	return func(query *eventFilter) {
		query.descending = true
	}
}

func (s *Service) QueryTimeline(ctx context.Context, limit int, filters ...TimelineFilter) ([]TimelineEvent, error) {
	if limit < 1 {
		return nil, fmt.Errorf("limit must be >= 1, got %d", limit)
	}

	// Build query.
	q := `SELECT e.id AS id,
				e.deployment_id,
				ir.key AS request_key,
				e.time_stamp,
				e.custom_key_1,
				e.custom_key_2,
				e.custom_key_3,
				e.custom_key_4,
				e.type,
				e.payload
			FROM timeline e
					 LEFT JOIN requests ir on e.request_id = ir.id
			WHERE true -- The "true" is to simplify the ANDs below.
		`

	filter := eventFilter{}
	for _, f := range filters {
		f(&filter)
	}
	var args []any
	index := 1
	param := func(v any) int {
		args = append(args, v)
		index++
		return index - 1
	}
	if !filter.olderThan.IsZero() {
		q += fmt.Sprintf(" AND e.time_stamp <= $%d::TIMESTAMPTZ", param(filter.olderThan))
	}
	if !filter.newerThan.IsZero() {
		q += fmt.Sprintf(" AND e.time_stamp >= $%d::TIMESTAMPTZ", param(filter.newerThan))
	}
	if filter.idHigherThan != 0 {
		q += fmt.Sprintf(" AND e.id >= $%d::BIGINT", param(filter.idHigherThan))
	}
	if filter.idLowerThan != 0 {
		q += fmt.Sprintf(" AND e.id <= $%d::BIGINT", param(filter.idLowerThan))
	}
	deploymentKeys := map[int64]model.DeploymentKey{}
	// TODO: We should probably make it mandatory to provide at least one deployment.
	deploymentQuery := `SELECT id, key FROM deployments`
	deploymentArgs := []any{}
	if len(filter.deployments) != 0 {
		// Unfortunately, if we use a join here, PG will do a sequential scan on
		// timeline and deployments, making a 7ms query into a 700ms query.
		// https://www.pgexplain.dev/plan/ecd44488-6060-4ad1-a9b4-49d092c3de81
		deploymentQuery += ` WHERE key = ANY($1::TEXT[])`
		deploymentArgs = append(deploymentArgs, filter.deployments)
	}
	rows, err := s.Handle.Connection.QueryContext(ctx, deploymentQuery, deploymentArgs...)
	if err != nil {
		return nil, libdal.TranslatePGError(err)
	}
	defer rows.Close() // nolint:errcheck
	deploymentIDs := []int64{}
	for rows.Next() {
		var id int64
		var key model.DeploymentKey
		if err := rows.Scan(&id, &key); err != nil {
			return nil, err
		}

		deploymentIDs = append(deploymentIDs, id)
		deploymentKeys[id] = key
	}

	q += fmt.Sprintf(` AND e.deployment_id = ANY($%d::BIGINT[])`, param(deploymentIDs))

	if filter.requests != nil {
		q += fmt.Sprintf(` AND ir.key = ANY($%d::TEXT[])`, param(filter.requests))
	}
	if filter.types != nil {
		// Why are we double casting? Because I hit "cannot find encode plan" and
		// this works around it: https://github.com/jackc/pgx/issues/338#issuecomment-333399112
		q += fmt.Sprintf(` AND e.type = ANY($%d::text[]::event_type[])`, param(filter.types))
	}
	if filter.level != nil {
		q += fmt.Sprintf(" AND (e.type != 'log' OR (e.type = 'log' AND e.custom_key_1::INT >= $%d::INT))\n", param(*filter.level))
	}
	if len(filter.calls) > 0 {
		q += " AND ("
		for i, call := range filter.calls {
			if i > 0 {
				q += " OR "
			}
			if sourceModule, ok := call.sourceModule.Get(); ok {
				q += fmt.Sprintf("(e.type != 'call' OR (e.type = 'call' AND e.custom_key_1 = $%d AND e.custom_key_3 = $%d))", param(sourceModule), param(call.destModule))
			} else if destVerb, ok := call.destVerb.Get(); ok {
				q += fmt.Sprintf("(e.type != 'call' OR (e.type = 'call' AND e.custom_key_3 = $%d AND e.custom_key_4 = $%d))", param(call.destModule), param(destVerb))
			} else {
				q += fmt.Sprintf("(e.type != 'call' OR (e.type = 'call' AND e.custom_key_3 = $%d))", param(call.destModule))
			}
		}
		q += ")\n"
	}

	if filter.descending {
		q += " ORDER BY e.time_stamp DESC, e.id DESC"
	} else {
		q += " ORDER BY e.time_stamp ASC, e.id ASC"
	}

	q += fmt.Sprintf(" LIMIT %d", limit)

	// Issue query.
	rows, err = s.Handle.Connection.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", q, libdal.TranslatePGError(err))
	}
	defer rows.Close()

	events, err := s.transformRowsToTimelineEvents(deploymentKeys, rows)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (s *Service) transformRowsToTimelineEvents(deploymentKeys map[int64]model.DeploymentKey, rows *stdsql.Rows) ([]TimelineEvent, error) {
	var out []TimelineEvent
	for rows.Next() {
		row := eventRow{}
		var deploymentID int64
		err := rows.Scan(
			&row.ID, &deploymentID, &row.RequestKey, &row.TimeStamp,
			&row.CustomKey1, &row.CustomKey2, &row.CustomKey3, &row.CustomKey4,
			&row.Type, &row.Payload,
		)
		if err != nil {
			return nil, err
		}

		row.DeploymentKey = deploymentKeys[deploymentID]

		switch row.Type {
		case sql.EventTypeLog:
			var jsonPayload eventLogJSON
			if err := s.encryption.DecryptJSON(&row.Payload, &jsonPayload); err != nil {
				return nil, fmt.Errorf("failed to decrypt log event: %w", err)
			}

			level, err := strconv.ParseInt(row.CustomKey1.MustGet(), 10, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid log level: %q: %w", row.CustomKey1.MustGet(), err)
			}
			out = append(out, &LogEvent{
				ID:            row.ID,
				DeploymentKey: row.DeploymentKey,
				RequestKey:    row.RequestKey,
				Time:          row.TimeStamp,
				Level:         int32(level),
				Attributes:    jsonPayload.Attributes,
				Message:       jsonPayload.Message,
				Error:         jsonPayload.Error,
			})

		case sql.EventTypeCall:
			var jsonPayload eventCallJSON
			if err := s.encryption.DecryptJSON(&row.Payload, &jsonPayload); err != nil {
				return nil, fmt.Errorf("failed to decrypt call event: %w", err)
			}
			var sourceVerb optional.Option[schema.Ref]
			sourceModule, smok := row.CustomKey1.Get()
			sourceName, snok := row.CustomKey2.Get()
			if smok && snok {
				sourceVerb = optional.Some(schema.Ref{Module: sourceModule, Name: sourceName})
			}
			out = append(out, &CallEvent{
				ID:            row.ID,
				DeploymentKey: row.DeploymentKey,
				RequestKey:    row.RequestKey,
				Time:          row.TimeStamp,
				SourceVerb:    sourceVerb,
				DestVerb:      schema.Ref{Module: row.CustomKey3.MustGet(), Name: row.CustomKey4.MustGet()},
				Duration:      time.Duration(jsonPayload.DurationMS) * time.Millisecond,
				Request:       jsonPayload.Request,
				Response:      jsonPayload.Response,
				Error:         jsonPayload.Error,
				Stack:         jsonPayload.Stack,
			})

		case sql.EventTypeDeploymentCreated:
			var jsonPayload eventDeploymentCreatedJSON
			if err := s.encryption.DecryptJSON(&row.Payload, &jsonPayload); err != nil {
				return nil, fmt.Errorf("failed to decrypt deployment created event: %w", err)
			}
			out = append(out, &DeploymentCreatedEvent{
				ID:                 row.ID,
				DeploymentKey:      row.DeploymentKey,
				Time:               row.TimeStamp,
				Language:           row.CustomKey1.MustGet(),
				ModuleName:         row.CustomKey2.MustGet(),
				MinReplicas:        jsonPayload.MinReplicas,
				ReplacedDeployment: jsonPayload.ReplacedDeployment,
			})

		case sql.EventTypeDeploymentUpdated:
			var jsonPayload eventDeploymentUpdatedJSON
			if err := s.encryption.DecryptJSON(&row.Payload, &jsonPayload); err != nil {
				return nil, fmt.Errorf("failed to decrypt deployment updated event: %w", err)
			}
			out = append(out, &DeploymentUpdatedEvent{
				ID:              row.ID,
				DeploymentKey:   row.DeploymentKey,
				Time:            row.TimeStamp,
				MinReplicas:     jsonPayload.MinReplicas,
				PrevMinReplicas: jsonPayload.PrevMinReplicas,
			})

		case sql.EventTypeIngress:
			var jsonPayload eventIngressJSON
			if err := s.encryption.DecryptJSON(&row.Payload, &jsonPayload); err != nil {
				return nil, fmt.Errorf("failed to decrypt ingress event: %w", err)
			}
			out = append(out, &DeploymentUpdatedEvent{
				ID:            row.ID,
				DeploymentKey: row.DeploymentKey,
				Time:          row.TimeStamp,
			})

		default:
			panic("unknown event type: " + row.Type)
		}
	}
	return out, nil
}