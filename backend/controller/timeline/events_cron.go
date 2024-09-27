package timeline

import (
	"context"
	"encoding/json"
	"time"

	"github.com/alecthomas/types/optional"

	ftlencryption "github.com/TBD54566975/ftl/backend/controller/encryption/api"
	"github.com/TBD54566975/ftl/backend/controller/timeline/internal/sql"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
)

type CronScheduledEvent struct {
	ID            int64
	DeploymentKey model.DeploymentKey
	Verb          schema.Ref

	Time        time.Time
	ScheduledAt time.Time
	Schedule    string
	Error       optional.Option[string]
}

func (e *CronScheduledEvent) GetID() int64 { return e.ID }
func (e *CronScheduledEvent) event()       {}

type eventCronScheduledJSON struct {
	ScheduledAt time.Time               `json:"scheduled_at"`
	Schedule    string                  `json:"schedule"`
	Error       optional.Option[string] `json:"error,omitempty"`
}

func (s *Service) InsertCronScheduledEvent(ctx context.Context, event *CronScheduledEvent) {
	logger := log.FromContext(ctx)

	cronJSON := eventCronScheduledJSON{
		ScheduledAt: event.ScheduledAt,
		Schedule:    event.Schedule,
		Error:       event.Error,
	}

	data, err := json.Marshal(cronJSON)
	if err != nil {
		logger.Errorf(err, "failed to marshal cron JSON")
		return
	}

	var payload ftlencryption.EncryptedTimelineColumn
	err = s.encryption.EncryptJSON(json.RawMessage(data), &payload)
	if err != nil {
		logger.Errorf(err, "failed to encrypt cron JSON")
		return
	}

	err = libdal.TranslatePGError(s.db.InsertTimelineCronScheduledEvent(ctx, sql.InsertTimelineCronScheduledEventParams{
		DeploymentKey: event.DeploymentKey,
		TimeStamp:     event.Time,
		Module:        event.Verb.Module,
		Verb:          event.Verb.Name,
		Payload:       payload,
	}))
	if err != nil {
		logger.Errorf(err, "failed to insert cron event")
	}
}
