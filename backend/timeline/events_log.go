package timeline

import (
	"time"

	"github.com/alecthomas/types/optional"
	"google.golang.org/protobuf/types/known/timestamppb"

	timelinepb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1"
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

var _ Event = Log{}

func (Log) clientEvent() {}
func (l Log) ToReq() (*timelinepb.CreateEventRequest, error) {
	var requestKey *string
	if r, ok := l.RequestKey.Get(); ok {
		key := r.String()
		requestKey = &key
	}
	return &timelinepb.CreateEventRequest{
		Entry: &timelinepb.CreateEventRequest_Log{
			Log: &timelinepb.LogEvent{
				DeploymentKey: l.DeploymentKey.String(),
				RequestKey:    requestKey,
				Timestamp:     timestamppb.New(l.Time),
				LogLevel:      l.Level,
				Attributes:    l.Attributes,
				Message:       l.Message,
				Error:         l.Error.Ptr(),
			},
		},
	}, nil
}