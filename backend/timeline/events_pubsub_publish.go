package timeline

import (
	"time"

	"github.com/alecthomas/types/optional"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	deployment "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/deployment/v1"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
	timelinepb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
)

type PubSubPublish struct {
	DeploymentKey model.DeploymentKey
	RequestKey    optional.Option[string]
	Time          time.Time
	SourceVerb    schema.Ref
	Topic         string
	// Should this just be request body?
	Request *deployment.PublishEventRequest
	Error   optional.Option[string]
}

var _ Event = PubSubPublish{}

func (PubSubPublish) clientEvent() {}
func (p PubSubPublish) ToReq() (*timelinepb.CreateEventRequest, error) {
	return &timelinepb.CreateEventRequest{
		Entry: &timelinepb.CreateEventRequest_PubsubPublish{
			PubsubPublish: &timelinepb.PubSubPublishEvent{
				DeploymentKey: p.DeploymentKey.String(),
				RequestKey:    p.RequestKey.Ptr(),
				VerbRef:       (&p.SourceVerb).ToProto().(*schemapb.Ref), //nolint:forcetypeassert
				Timestamp:     timestamppb.New(p.Time),
				Duration:      durationpb.New(time.Since(p.Time)),
				Topic:         p.Topic,
				Request:       string(p.Request.Body),
				Error:         p.Error.Ptr(),
			},
		},
	}, nil
}
