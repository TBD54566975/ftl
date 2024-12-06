package timeline

import (
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
)

type CronScheduledEvent struct {
	ID       int64
	Duration time.Duration
	CronScheduled
}

func (e *CronScheduledEvent) GetID() int64 { return e.ID }
func (e *CronScheduledEvent) event()       {}

type CronScheduled struct {
	DeploymentKey model.DeploymentKey
	Verb          schema.Ref

	Time        time.Time
	ScheduledAt time.Time
	Schedule    string
	Error       optional.Option[string]
}

func (e *CronScheduled) toEvent() (Event, error) { //nolint:unparam
	return &CronScheduledEvent{
		CronScheduled: *e,
		Duration:      time.Since(e.Time),
	}, nil
}

type eventCronScheduledJSON struct {
	DurationMS  int64                   `json:"duration_ms"`
	ScheduledAt time.Time               `json:"scheduled_at"`
	Schedule    string                  `json:"schedule"`
	Error       optional.Option[string] `json:"error,omitempty"`
}
