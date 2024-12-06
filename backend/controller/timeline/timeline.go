package timeline

import (
	"context"
	stdsql "database/sql"
	"time"

	"github.com/alecthomas/atomic"

	"github.com/TBD54566975/ftl/backend/controller/encryption"
	"github.com/TBD54566975/ftl/backend/controller/timeline/internal/sql"
	"github.com/TBD54566975/ftl/internal/log"
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
	ctx              context.Context
	conn             *stdsql.DB
	encryption       *encryption.Service
	events           chan Event
	lastDroppedError atomic.Value[time.Time]
	lastFailedError  atomic.Value[time.Time]
}

func New(ctx context.Context, conn *stdsql.DB, encryption *encryption.Service) *Service {
	var s *Service
	events := make(chan Event, 1000)
	s = &Service{
		ctx:        ctx,
		conn:       conn,
		encryption: encryption,
		events:     events,
	}
	return s
}

// EnqueueEvent asynchronously enqueues an event for insertion into the timeline.
func (s *Service) EnqueueEvent(ctx context.Context, inEvent InEvent) {
	event, err := inEvent.toEvent()
	if err != nil {
		log.FromContext(ctx).Warnf("Failed to convert event to event: %v", err)
		return
	}
	select {
	case s.events <- event:
	default:
		if time.Since(s.lastDroppedError.Load()) > 10*time.Second {
			log.FromContext(ctx).Warnf("Dropping event %T due to full queue", event)
			s.lastDroppedError.Store(time.Now())
		}
	}
}
