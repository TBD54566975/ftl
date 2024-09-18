package timeline

import (
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/internal/model"
)

type DeploymentCreatedEvent struct {
	ID                 int64
	DeploymentKey      model.DeploymentKey
	Time               time.Time
	Language           string
	ModuleName         string
	MinReplicas        int
	ReplacedDeployment optional.Option[model.DeploymentKey]
}

func (e *DeploymentCreatedEvent) GetID() int64 { return e.ID }
func (e *DeploymentCreatedEvent) event()       {}

type eventDeploymentUpdatedJSON struct {
	MinReplicas     int `json:"min_replicas"`
	PrevMinReplicas int `json:"prev_min_replicas"`
}

type DeploymentUpdatedEvent struct {
	ID              int64
	DeploymentKey   model.DeploymentKey
	Time            time.Time
	MinReplicas     int
	PrevMinReplicas int
}

func (e *DeploymentUpdatedEvent) GetID() int64 { return e.ID }
func (e *DeploymentUpdatedEvent) event()       {}

type eventDeploymentCreatedJSON struct {
	MinReplicas        int                                  `json:"min_replicas"`
	ReplacedDeployment optional.Option[model.DeploymentKey] `json:"replaced,omitempty"`
}
