package timeline

import (
	"time"

	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/result"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
	timelinepb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
)

type Call struct {
	DeploymentKey    model.DeploymentKey
	RequestKey       model.RequestKey
	ParentRequestKey optional.Option[model.RequestKey]
	StartTime        time.Time
	DestVerb         *schema.Ref
	Callers          []*schema.Ref
	Request          *ftlv1.CallRequest
	Response         result.Result[*ftlv1.CallResponse]
}

func (Call) clientEvent() {}
func (c Call) ToReq() (*timelinepb.CreateEventRequest, error) {
	requestKey := c.RequestKey.String()

	var respError *string
	var responseBody []byte
	var stack *string
	resp, err := c.Response.Result()
	if err != nil {
		errStr := err.Error()
		respError = &errStr
	} else {
		responseBody = resp.GetBody()
		if callError := resp.GetError(); callError != nil {
			respError = optional.Some(callError.Message).Ptr()
			stack = callError.Stack
		}
	}
	var sourceVerb *schemapb.Ref
	if len(c.Callers) > 0 {
		sourceVerb = c.Callers[0].ToProto().(*schemapb.Ref) //nolint:forcetypeassert
	}

	return &timelinepb.CreateEventRequest{
		Entry: &timelinepb.CreateEventRequest_Call{
			Call: &timelinepb.CallEvent{
				RequestKey:         &requestKey,
				DeploymentKey:      c.DeploymentKey.String(),
				Timestamp:          timestamppb.New(c.StartTime),
				Response:           string(responseBody),
				Error:              respError,
				SourceVerbRef:      sourceVerb,
				DestinationVerbRef: c.DestVerb.ToProto().(*schemapb.Ref), //nolint:forcetypeassert
				Duration:           durationpb.New(time.Since(c.StartTime)),
				Request:            string(c.Request.GetBody()),
				Stack:              stack,
			},
		},
	}, nil
}