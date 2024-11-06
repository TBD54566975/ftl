package timeline

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"

	ftlencryption "github.com/TBD54566975/ftl/backend/controller/encryption/api"
	"github.com/TBD54566975/ftl/backend/controller/timeline/internal/sql"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
)

type AsyncExecuteEvent struct {
	ID       int64
	Duration time.Duration
	AsyncExecute
}

func (e *AsyncExecuteEvent) GetID() int64 { return e.ID }
func (e *AsyncExecuteEvent) event()       {}

type AsyncExecuteEventType string

const (
	AsyncExecuteEventTypeUnkown AsyncExecuteEventType = "unknown"
	AsyncExecuteEventTypeCron   AsyncExecuteEventType = "cron"
	AsyncExecuteEventTypePubSub AsyncExecuteEventType = "pubsub"
)

type AsyncExecute struct {
	DeploymentKey model.DeploymentKey
	RequestKey    optional.Option[string]
	EventType     AsyncExecuteEventType
	Verb          schema.Ref
	Time          time.Time
	Error         optional.Option[string]
}

func (e *AsyncExecute) toEvent() (Event, error) { //nolint:unparam
	return &AsyncExecuteEvent{
		AsyncExecute: *e,
		Duration:     time.Since(e.Time),
	}, nil
}

type eventAsyncExecuteJSON struct {
	DurationMS int64                   `json:"duration_ms"`
	EventType  AsyncExecuteEventType   `json:"event_type"`
	Error      optional.Option[string] `json:"error,omitempty"`
}

func (s *Service) insertAsyncExecuteEvent(ctx context.Context, querier sql.Querier, event *AsyncExecuteEvent) error {
	asyncJSON := eventAsyncExecuteJSON{
		DurationMS: event.Duration.Milliseconds(),
		EventType:  event.EventType,
		Error:      event.Error,
	}

	data, err := json.Marshal(asyncJSON)
	if err != nil {
		return fmt.Errorf("failed to marshal async execute event: %w", err)
	}

	var payload ftlencryption.EncryptedTimelineColumn
	err = s.encryption.EncryptJSON(json.RawMessage(data), &payload)
	if err != nil {
		return fmt.Errorf("failed to encrypt cron JSON: %w", err)
	}

	err = libdal.TranslatePGError(querier.InsertTimelineAsyncExecuteEvent(ctx, sql.InsertTimelineAsyncExecuteEventParams{
		DeploymentKey: event.DeploymentKey,
		RequestKey:    event.RequestKey,
		TimeStamp:     event.Time,
		Module:        event.Verb.Module,
		Verb:          event.Verb.Name,
		Payload:       payload,
	}))
	if err != nil {
		return fmt.Errorf("failed to insert async execute event: %w", err)
	}
	return err
}
