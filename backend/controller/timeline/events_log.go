package timeline

import (
	"time"

	"github.com/alecthomas/types/optional"

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

func (l *Log) toEvent() (Event, error) { return &LogEvent{Log: *l}, nil } //nolint:unparam

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
