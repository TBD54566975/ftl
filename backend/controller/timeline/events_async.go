package timeline

import (
	"time"

	"github.com/alecthomas/types/optional"

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
