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

type CronScheduled struct {
	DeploymentKey model.DeploymentKey
	Verb          schema.Ref
	Time          time.Time
	ScheduledAt   time.Time
	Schedule      string
	Error         optional.Option[string]
}

var _ Event = CronScheduled{}

func (e CronScheduled) clientEvent() {}
func (e CronScheduled) ToEntry() (*timelinepb.CreateEventsRequest_EventEntry, error) {
	return &timelinepb.CreateEventsRequest_EventEntry{
		Entry: &timelinepb.CreateEventsRequest_EventEntry_CronScheduled{
			CronScheduled: &timelinepb.CronScheduledEvent{
				DeploymentKey: e.DeploymentKey.String(),
				VerbRef:       (&e.Verb).ToProto(), //nolint:forcetypeassert
				Timestamp:     timestamppb.New(e.Time),
				ScheduledAt:   timestamppb.New(e.ScheduledAt),
				Schedule:      e.Schedule,
				Error:         e.Error.Ptr(),
				Duration:      durationpb.New(time.Since(e.Time)),
			},
		},
	}, nil
}
