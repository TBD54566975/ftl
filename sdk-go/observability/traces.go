package observability

import (
	"context"
	"encoding/json"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"
	"github.com/jpillora/backoff"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

var ErrDroppedTraceEvent = errors.New("observability trace event dropped")

type SpanExporterConfig struct {
	TracesBuffer int `default:"1048576" help:"Number of traces to buffer before dropping."`
}

type SpanExporter struct {
	client ftlv1connect.ObservabilityServiceClient
	queue  chan *ftlv1.SendTracesRequest
}

func NewSpanExporter(ctx context.Context, client ftlv1connect.ObservabilityServiceClient, config SpanExporterConfig) *SpanExporter {
	e := &SpanExporter{
		client: client,
		queue:  make(chan *ftlv1.SendTracesRequest, config.TracesBuffer),
	}
	go rpc.RetryStreamingClientStream(ctx, backoff.Backoff{}, e.client.SendTraces, e.sendLoop)
	return e
}

var _ trace.SpanExporter = (*SpanExporter)(nil)

func (s *SpanExporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	spanStubs := tracetest.SpanStubsFromReadOnlySpans(spans)
	data, err := json.Marshal(spanStubs)
	if err != nil {
		return errors.WithStack(err)
	}

	select {
	case s.queue <- &ftlv1.SendTracesRequest{Json: data}:
	default:
		return errors.WithStack(ErrDroppedMetricEvent)
	}
	return nil
}

func (s *SpanExporter) Shutdown(ctx context.Context) error {
	close(s.queue)
	return nil
}

func (s *SpanExporter) sendLoop(ctx context.Context, stream *connect.ClientStreamForClient[ftlv1.SendTracesRequest, ftlv1.SendTracesResponse]) error {
	logger := log.FromContext(ctx)
	logger.Infof("Traces send loop started")
	for {
		select {
		case <-ctx.Done():
			return errors.WithStack(context.Cause(ctx))

		case event, ok := <-s.queue:
			if !ok {
				return nil
			}
			logger.Infof("%s", event.Json)
			if err := stream.Send(event); err != nil {
				select {
				case s.queue <- event:
				default:
					log.FromContext(ctx).Errorf(errors.WithStack(ErrDroppedMetricEvent), "traces queue full while handling error")
				}
				return errors.WithStack(err)
			}
		}
	}
}
