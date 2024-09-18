package timeline

import (
	"context"
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"

	ftlencryption "github.com/TBD54566975/ftl/backend/controller/encryption/api"
	"github.com/TBD54566975/ftl/backend/controller/timeline/internal/sql"
	"github.com/TBD54566975/ftl/backend/libdal"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/internal/model"
)

type LogEvent struct {
	ID            int64
	DeploymentKey model.DeploymentKey
	RequestKey    optional.Option[model.RequestKey]
	Time          time.Time
	Level         int32
	Attributes    map[string]string
	Message       string
	Error         optional.Option[string]
}

func (e *LogEvent) GetID() int64 { return e.ID }
func (e *LogEvent) event()       {}

type eventLogJSON struct {
	Message    string                  `json:"message"`
	Attributes map[string]string       `json:"attributes"`
	Error      optional.Option[string] `json:"error,omitempty"`
}

type Log struct {
	DeploymentKey model.DeploymentKey
	RequestKey    optional.Option[model.RequestKey]
	Msg           *ftlv1.StreamDeploymentLogsRequest
}

func (s *Service) RecordLog(ctx context.Context, log *Log) error {
	err := s.InsertLogEvent(ctx, &LogEvent{
		RequestKey:    log.RequestKey,
		DeploymentKey: log.DeploymentKey,
		Time:          log.Msg.TimeStamp.AsTime(),
		Level:         log.Msg.LogLevel,
		Attributes:    log.Msg.Attributes,
		Message:       log.Msg.Message,
		Error:         optional.Ptr(log.Msg.Error),
	})
	if err != nil {
		return fmt.Errorf("failed to insert log event: %w", err)
	}
	return nil
}

func (s *Service) InsertLogEvent(ctx context.Context, log *LogEvent) error {
	var requestKey optional.Option[string]
	if name, ok := log.RequestKey.Get(); ok {
		requestKey = optional.Some(name.String())
	}

	payload := map[string]any{
		"message":    log.Message,
		"attributes": log.Attributes,
		"error":      log.Error,
	}
	var encryptedPayload ftlencryption.EncryptedTimelineColumn
	err := s.encryption.EncryptJSON(payload, &encryptedPayload)
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
