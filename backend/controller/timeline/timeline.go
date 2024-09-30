package timeline

import (
	"context"
	stdsql "database/sql"
	"fmt"
	"time"

	"github.com/alecthomas/atomic"

	"github.com/TBD54566975/ftl/backend/controller/encryption"
	"github.com/TBD54566975/ftl/backend/controller/observability"
	"github.com/TBD54566975/ftl/backend/controller/sql/sqltypes"
	"github.com/TBD54566975/ftl/backend/controller/timeline/internal/sql"
	"github.com/TBD54566975/ftl/backend/libdal"
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

	maxBatchSize  = 16
	maxBatchDelay = 100 * time.Millisecond
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
	inEvent()
}

type Service struct {
	ctx              context.Context
	conn             *stdsql.DB
	encryption       *encryption.Service
	events           chan InEvent
	lastDroppedError atomic.Value[time.Time]
	lastFailedError  atomic.Value[time.Time]
}

func New(ctx context.Context, conn *stdsql.DB, encryption *encryption.Service) *Service {
	var s *Service
	events := make(chan InEvent, 1000)
	s = &Service{
		ctx:        ctx,
		conn:       conn,
		encryption: encryption,
		events:     events,
	}
	go s.processEvents()
	return s
}

func (s *Service) DeleteOldEvents(ctx context.Context, eventType EventType, age time.Duration) (int64, error) {
	count, err := sql.New(s.conn).DeleteOldTimelineEvents(ctx, sqltypes.Duration(age), eventType)
	return count, libdal.TranslatePGError(err)
}

// EnqueueEvent asynchronously enqueues an event for insertion into the timeline.
func (s *Service) EnqueueEvent(ctx context.Context, event InEvent) {
	select {
	case s.events <- event:
	default:
		if time.Since(s.lastDroppedError.Load()) > 10*time.Second {
			log.FromContext(ctx).Warnf("Dropping event %T due to full queue", event)
			s.lastDroppedError.Store(time.Now())
		}
	}
}

func (s *Service) processEvents() {
	lastFlush := time.Now()
	buffer := make([]InEvent, 0, maxBatchSize)
	for {
		select {
		case event := <-s.events:
			buffer = append(buffer, event)

			if len(buffer) < maxBatchSize || time.Since(lastFlush) < maxBatchDelay {
				continue
			}
			s.flushEvents(buffer)
			buffer = nil

		case <-time.After(maxBatchDelay):
			if len(buffer) == 0 {
				continue
			}
			s.flushEvents(buffer)
			buffer = nil
		}
	}
}

// Flush all events in the buffer to the database in a single transaction.
func (s *Service) flushEvents(events []InEvent) {
	logger := log.FromContext(s.ctx).Scope("timeline")
	tx, err := s.conn.Begin()
	if err != nil {
		logger.Errorf(err, "Failed to start transaction")
		return
	}
	querier := sql.New(tx)
	var lastError error
	failures := 0
	for _, event := range events {
		var err error
		switch e := event.(type) {
		case *Call:
			err = s.insertCallEvent(s.ctx, querier, e)
		case *Log:
			err = s.insertLogEvent(s.ctx, querier, e)
		case *Ingress:
			err = s.insertHTTPIngress(s.ctx, querier, e)
		default:
			panic(fmt.Sprintf("unexpected event type: %T", e))
		}
		if err != nil {
			lastError = err
			failures++
		}
	}
	err = tx.Commit()
	if err != nil {
		failures = len(events)
		lastError = err
	}
	if lastError != nil {
		if time.Since(s.lastFailedError.Load()) > 10*time.Second {
			logger.Errorf(lastError, "Failed to insert %d events, most recent error", failures)
			s.lastFailedError.Store(time.Now())
		}
		observability.Timeline.Failed(s.ctx, failures)
	}
	observability.Timeline.Inserted(s.ctx, len(events)-failures)
}
