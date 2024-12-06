package timeline

import (
	"encoding/json"
	"time"

	"github.com/alecthomas/types/optional"

	deployment "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/deployment/v1"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
)

type PubSubPublishEvent struct {
	ID       int64
	Duration time.Duration
	Request  json.RawMessage
	PubSubPublish
}

func (e *PubSubPublishEvent) GetID() int64 { return e.ID }
func (e *PubSubPublishEvent) event()       {}

type PubSubPublish struct {
	DeploymentKey model.DeploymentKey
	RequestKey    optional.Option[string]
	Time          time.Time
	SourceVerb    schema.Ref
	Topic         string
	Request       *deployment.PublishEventRequest
	Error         optional.Option[string]
}

func (e *PubSubPublish) toEvent() (Event, error) { //nolint:unparam
	return &PubSubPublishEvent{
		PubSubPublish: *e,
		Duration:      time.Since(e.Time),
	}, nil
}

type eventPubSubPublishJSON struct {
	DurationMS int64                   `json:"duration_ms"`
	Topic      string                  `json:"topic"`
	Request    json.RawMessage         `json:"request"`
	Error      optional.Option[string] `json:"error,omitempty"`
}
