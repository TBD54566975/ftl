package timeline

import (
	"time"

	"github.com/alecthomas/types/optional"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
	timelinepb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
)

type AsyncExecuteEventType string

const (
	AsyncExecuteEventTypeUnkown AsyncExecuteEventType = "unknown"
	AsyncExecuteEventTypeCron   AsyncExecuteEventType = "cron"
	AsyncExecuteEventTypePubSub AsyncExecuteEventType = "pubsub"
)

func asyncExecuteEventTypeToProto(eventType AsyncExecuteEventType) timelinepb.AsyncExecuteEventType {
	switch eventType {
	case AsyncExecuteEventTypeCron:
		return timelinepb.AsyncExecuteEventType_ASYNC_EXECUTE_EVENT_TYPE_CRON
	case AsyncExecuteEventTypePubSub:
		return timelinepb.AsyncExecuteEventType_ASYNC_EXECUTE_EVENT_TYPE_PUBSUB
	case AsyncExecuteEventTypeUnkown:
		return timelinepb.AsyncExecuteEventType_ASYNC_EXECUTE_EVENT_TYPE_UNSPECIFIED

	default:
		panic("unknown async execute event type")
	}
}

type AsyncExecute struct {
	DeploymentKey model.DeploymentKey
	RequestKey    optional.Option[string]
	EventType     AsyncExecuteEventType
	Verb          schema.Ref
	Time          time.Time
	Error         optional.Option[string]
}

var _ Event = AsyncExecute{}

func (AsyncExecute) clientEvent() {}
func (a AsyncExecute) ToEntry() (*timelinepb.CreateEventsRequest_EventEntry, error) {
	return &timelinepb.CreateEventsRequest_EventEntry{
		Entry: &timelinepb.CreateEventsRequest_EventEntry_AsyncExecute{
			AsyncExecute: &timelinepb.AsyncExecuteEvent{
				DeploymentKey:  a.DeploymentKey.String(),
				RequestKey:     a.RequestKey.Ptr(),
				Timestamp:      timestamppb.New(a.Time),
				Error:          a.Error.Ptr(),
				Duration:       durationpb.New(time.Since(a.Time)),
				VerbRef:        (&a.Verb).ToProto().(*schemapb.Ref), //nolint:forcetypeassert
				AsyncEventType: asyncExecuteEventTypeToProto(a.EventType),
			},
		},
	}, nil
}
