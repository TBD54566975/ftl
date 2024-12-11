package timeline

import (
	"time"

	"github.com/alecthomas/types/optional"

	timelinepb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	"github.com/TBD54566975/ftl/internal/model"
)

type DeploymentCreated struct {
	DeploymentKey      model.DeploymentKey
	Time               time.Time
	Language           string
	ModuleName         string
	MinReplicas        int
	ReplacedDeployment optional.Option[model.DeploymentKey]
}

var _ Event = DeploymentCreated{}

func (DeploymentCreated) clientEvent() {}
func (d DeploymentCreated) ToEntry() (*timelinepb.CreateEventsRequest_EventEntry, error) {
	var replaced *string
	if r, ok := d.ReplacedDeployment.Get(); ok {
		repl := r.String()
		replaced = &repl
	}
	return &timelinepb.CreateEventsRequest_EventEntry{
		Entry: &timelinepb.CreateEventsRequest_EventEntry_DeploymentCreated{
			DeploymentCreated: &timelinepb.DeploymentCreatedEvent{
				Key:         d.DeploymentKey.String(),
				Language:    d.Language,
				ModuleName:  d.ModuleName,
				MinReplicas: int32(d.MinReplicas),
				Replaced:    replaced,
			},
		},
	}, nil
}

type DeploymentUpdated struct {
	DeploymentKey   model.DeploymentKey
	Time            time.Time
	MinReplicas     int
	PrevMinReplicas int
}

var _ Event = DeploymentUpdated{}

func (DeploymentUpdated) clientEvent() {}
func (d DeploymentUpdated) ToEntry() (*timelinepb.CreateEventsRequest_EventEntry, error) {
	return &timelinepb.CreateEventsRequest_EventEntry{
		Entry: &timelinepb.CreateEventsRequest_EventEntry_DeploymentUpdated{
			DeploymentUpdated: &timelinepb.DeploymentUpdatedEvent{
				Key:             d.DeploymentKey.String(),
				MinReplicas:     int32(d.MinReplicas),
				PrevMinReplicas: int32(d.PrevMinReplicas),
			},
		},
	}, nil
}
