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
)

type Log struct {
	DeploymentKey model.DeploymentKey
	RequestKey    optional.Option[model.RequestKey]
	Time          time.Time
	Level         int32
	Attributes    map[string]string
	Message       string
	Error         optional.Option[string]
}

type LogEvent struct {
	ID int64
	Log
}

func (e *LogEvent) GetID() int64 { return e.ID }
func (e *LogEvent) event()       {}

type eventLogJSON struct {
	Message    string                  `json:"message"`
	Attributes map[string]string       `json:"attributes"`
	Error      optional.Option[string] `json:"error,omitempty"`
}

func (s *Service) InsertLogEvent(ctx context.Context, log *Log) error {
	var requestKey optional.Option[string]
	if name, ok := log.RequestKey.Get(); ok {
		requestKey = optional.Some(name.String())
	}

	logJSON := eventLogJSON{
		Message:    log.Message,
		Attributes: log.Attributes,
		Error:      log.Error,
	}

	data, err := json.Marshal(logJSON)
	if err != nil {
		return fmt.Errorf("failed to marshal log event: %w", err)
	}

	var encryptedPayload ftlencryption.EncryptedTimelineColumn
	err = s.encryption.EncryptJSON(json.RawMessage(data), &encryptedPayload)
	if err != nil {
		return fmt.Errorf("failed to encrypt log payload: %w", err)
	}

	return libdal.TranslatePGError(s.db.InsertTimelineLogEvent(ctx, sql.InsertTimelineLogEventParams{
		DeploymentKey: log.DeploymentKey,
		RequestKey:    requestKey,
		TimeStamp:     log.Time,
		Level:         log.Level,
		Payload:       encryptedPayload,
	}))
}
