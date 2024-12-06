package timeline

import (
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
)

type PubSubConsumeEvent struct {
	ID       int64
	Duration time.Duration
	PubSubConsume
}

func (e *PubSubConsumeEvent) GetID() int64 { return e.ID }
func (e *PubSubConsumeEvent) event()       {}

type PubSubConsume struct {
	DeploymentKey model.DeploymentKey
	RequestKey    optional.Option[string]
	Time          time.Time
	DestVerb      optional.Option[schema.RefKey]
	Topic         string
	Error         optional.Option[string]
}

func (e *PubSubConsume) toEvent() (Event, error) { //nolint:unparam
	return &PubSubConsumeEvent{
		PubSubConsume: *e,
		Duration:      time.Since(e.Time),
	}, nil
}

type eventPubSubConsumeJSON struct {
	DurationMS int64                   `json:"duration_ms"`
	Topic      string                  `json:"topic"`
	Error      optional.Option[string] `json:"error,omitempty"`
}
