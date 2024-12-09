package timeline

import (
	"time"

	"github.com/alecthomas/types/optional"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	timelinepb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
)

type PubSubConsume struct {
	DeploymentKey model.DeploymentKey
	RequestKey    optional.Option[string]
	Time          time.Time
	DestVerb      optional.Option[schema.RefKey]
	Topic         string
	Error         optional.Option[string]
}

var _ Event = PubSubConsume{}

func (PubSubConsume) clientEvent() {}
func (p PubSubConsume) ToEntry() (*timelinepb.CreateEventsRequest_EventEntry, error) {
	var destModule, destVerb *string
	if ref, ok := p.DestVerb.Get(); ok {
		destModule = &ref.Module
		destVerb = &ref.Name
	}
	return &timelinepb.CreateEventsRequest_EventEntry{
		Entry: &timelinepb.CreateEventsRequest_EventEntry_PubsubConsume{
			PubsubConsume: &timelinepb.PubSubConsumeEvent{
				DeploymentKey:  p.DeploymentKey.String(),
				RequestKey:     p.RequestKey.Ptr(),
				Timestamp:      timestamppb.New(p.Time),
				Topic:          p.Topic,
				Error:          p.Error.Ptr(),
				DestVerbModule: destModule,
				DestVerbName:   destVerb,
				Duration:       durationpb.New(time.Since(p.Time)),
			},
		},
	}, nil
}
