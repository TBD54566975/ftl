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

func (e *PubSubConsume) toEvent() (Event, error) {
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

func (s *Service) insertPubSubConsumeEvent(ctx context.Context, querier sql.Querier, event *PubSubConsumeEvent) error {
	pubsubJSON := eventPubSubConsumeJSON{
		DurationMS: event.Duration.Milliseconds(),
		Topic:      event.Topic,
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

	destModule := optional.None[string]()
	destVerb := optional.None[string]()
	if dv, ok := event.DestVerb.Get(); ok {
		destModule = optional.Some(dv.Module)
		destVerb = optional.Some(dv.Name)
	}

	err = libdal.TranslatePGError(querier.InsertTimelinePubsubConsumeEvent(ctx, sql.InsertTimelinePubsubConsumeEventParams{
		DeploymentKey: event.DeploymentKey,
		RequestKey:    event.RequestKey,
		TimeStamp:     event.Time,
		DestModule:    destModule,
		DestVerb:      destVerb,
		Topic:         event.Topic,
		Payload:       payload,
	}))
	if err != nil {
		return fmt.Errorf("failed to insert pubsub consume event: %w", err)
	}
	return err
}
