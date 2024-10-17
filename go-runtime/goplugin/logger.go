package goplugin

import (
	"connectrpc.com/connect"

	langpb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/language"
	"github.com/TBD54566975/ftl/internal/log"
)

type streamLogSink struct {
	stream *connect.ServerStream[langpb.BuildEvent]
}

var _ log.Sink = streamLogSink{}

func newLoggerForStream(level log.Level, stream *connect.ServerStream[langpb.BuildEvent]) *log.Logger {
	return log.New(level, streamLogSink{stream: stream})
}

func (s streamLogSink) Log(e log.Entry) error {
	return s.stream.Send(&langpb.BuildEvent{
		Event: &langpb.BuildEvent_LogMessage{
			LogMessage: &langpb.LogMessage{
				Level:   langpb.LogLevelToProto(e.Level),
				Message: e.Message,
			},
		},
	})
}
