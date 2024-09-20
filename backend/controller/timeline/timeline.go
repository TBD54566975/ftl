package timeline

import (
	"context"
	"time"

	"github.com/TBD54566975/ftl/backend/controller/encryption"
	"github.com/TBD54566975/ftl/backend/controller/sql/sqltypes"
	"github.com/TBD54566975/ftl/backend/controller/timeline/internal/sql"
	"github.com/TBD54566975/ftl/backend/libdal"
)

type EventType = sql.EventType

// Supported event types.
const (
	EventTypeLog               = sql.EventTypeLog
	EventTypeCall              = sql.EventTypeCall
	EventTypeDeploymentCreated = sql.EventTypeDeploymentCreated
	EventTypeDeploymentUpdated = sql.EventTypeDeploymentUpdated
	EventTypeIngress           = sql.EventTypeIngress
)

// TimelineEvent types.
//
//sumtype:decl
type TimelineEvent interface {
	GetID() int64
	event()
}

type Service struct {
	*libdal.Handle[Service]
	db         sql.Querier
	encryption *encryption.Service
}

func New(ctx context.Context, conn libdal.Connection, encryption *encryption.Service) *Service {
	var s *Service
	s = &Service{
		db:         sql.New(conn),
		encryption: encryption,
		Handle: libdal.New(conn, func(h *libdal.Handle[Service]) *Service {
			return &Service{
				Handle:     h,
				db:         sql.New(h.Connection),
				encryption: s.encryption,
			}
		}),
	}
	return s
}

func (s *Service) DeleteOldEvents(ctx context.Context, eventType EventType, age time.Duration) (int64, error) {
	count, err := s.db.DeleteOldTimelineEvents(ctx, sqltypes.Duration(age), eventType)
	return count, libdal.TranslatePGError(err)
}
