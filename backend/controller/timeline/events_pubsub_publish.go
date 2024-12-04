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
	pubpb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/publisher/v1"
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
	Request       *pubpb.PublishEventRequest
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

func (s *Service) insertPubSubPublishEvent(ctx context.Context, querier sql.Querier, event *PubSubPublishEvent) error {
	pubsubJSON := eventPubSubPublishJSON{
		DurationMS: event.Duration.Milliseconds(),
		Topic:      event.Topic,
		Request:    event.Request,
		Error:      event.Error,
	}

	data, err := json.Marshal(pubsubJSON)
	if err != nil {
		return fmt.Errorf("failed to marshal pubsub event: %w", err)
	}

	var payload ftlencryption.EncryptedTimelineColumn
	err = s.encryption.EncryptJSON(json.RawMessage(data), &payload)
	if err != nil {
		return fmt.Errorf("failed to encrypt cron JSON: %w", err)
	}

	err = libdal.TranslatePGError(querier.InsertTimelinePubsubPublishEvent(ctx, sql.InsertTimelinePubsubPublishEventParams{
		DeploymentKey: event.DeploymentKey,
		RequestKey:    event.RequestKey,
		TimeStamp:     event.Time,
		SourceModule:  event.SourceVerb.Module,
		SourceVerb:    event.SourceVerb.Name,
		Topic:         event.Topic,
		Payload:       payload,
	}))
	if err != nil {
		return fmt.Errorf("failed to insert pubsub publish event: %w", err)
	}
	return err
}
