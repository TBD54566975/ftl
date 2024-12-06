package timeline

import (
	"context"
	stdsql "database/sql"

	"github.com/TBD54566975/ftl/backend/controller/encryption"
	"github.com/TBD54566975/ftl/backend/controller/timeline/internal/sql"
)

type EventType = sql.EventType

// Supported event types.
const (
	EventTypeLog               = sql.EventTypeLog
	EventTypeCall              = sql.EventTypeCall
	EventTypeDeploymentCreated = sql.EventTypeDeploymentCreated
	EventTypeDeploymentUpdated = sql.EventTypeDeploymentUpdated
	EventTypeIngress           = sql.EventTypeIngress
	EventTypeCronScheduled     = sql.EventTypeCronScheduled
	EventTypeAsyncExecute      = sql.EventTypeAsyncExecute
	EventTypePubSubPublish     = sql.EventTypePubsubPublish
	EventTypePubSubConsume     = sql.EventTypePubsubConsume
)

// Event types.
//
//sumtype:decl
type Event interface {
	GetID() int64
	event()
}

// InEvent is a marker interface for events that are inserted into the timeline.
type InEvent interface {
	toEvent() (Event, error)
}

type Service struct {
	ctx        context.Context
	conn       *stdsql.DB
	encryption *encryption.Service
}

func New(ctx context.Context, conn *stdsql.DB, encryption *encryption.Service) *Service {
	s := &Service{
		ctx:        ctx,
		conn:       conn,
		encryption: encryption,
	}
	return s
}
