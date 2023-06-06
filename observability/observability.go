package observability

import (
	"context"

	"github.com/alecthomas/atomic"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types"
	"github.com/bufbuild/connect-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"github.com/TBD54566975/ftl/internal/log"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

type Observability struct {
	registrationFailure atomic.Value[types.Option[error]]
}

type Config struct {
	MetricsExporterConfig `embed:"" prefix:"metrics-"`
	SpanExporterConfig    `embed:"" prefix:"traces-"`
}

func NewObservability() *Observability {
	return &Observability{}
}

var _ ftlv1connect.ObservabilityServiceHandler = (*Observability)(nil)

func Init(ctx context.Context, observabilityServiceClient ftlv1connect.ObservabilityServiceClient, name string, conf Config) {
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(name),
	)

	ftlSpanExporter := NewSpanExporter(ctx, observabilityServiceClient, conf.SpanExporterConfig)

	tp := trace.NewTracerProvider(
		trace.WithBatcher(ftlSpanExporter),

		trace.WithResource(res))

	otel.SetTracerProvider(tp)

	ftlMetricsExporter := NewMetricsExporter(ctx, observabilityServiceClient, conf.MetricsExporterConfig)

	provider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(ftlMetricsExporter, metric.WithInterval(conf.Interval))),
	)

	otel.SetMeterProvider(provider)
}

func (o *Observability) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	var notReady *string
	if err, ok := o.registrationFailure.Load().Get(); ok {
		msg := err.Error()
		notReady = &msg
	}
	return connect.NewResponse(&ftlv1.PingResponse{NotReady: notReady}), nil
}

func (o *Observability) SendTraces(ctx context.Context, req *connect.ClientStream[ftlv1.SendTracesRequest]) (*connect.Response[ftlv1.SendTracesResponse], error) {
	logger := log.FromContext(ctx)
	for req.Receive() {
		logger.Tracef("Traces: %s", req.Msg().Json)
	}
	if err := req.Err(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.WithStack(err))
	}
	return connect.NewResponse(&ftlv1.SendTracesResponse{}), nil
}

func (o *Observability) SendMetrics(ctx context.Context, req *connect.ClientStream[ftlv1.SendMetricsRequest]) (*connect.Response[ftlv1.SendMetricsResponse], error) {
	logger := log.FromContext(ctx)
	for req.Receive() {
		logger.Tracef("Metrics: %s", req.Msg().Json)
	}
	if err := req.Err(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.WithStack(err))
	}
	return connect.NewResponse(&ftlv1.SendMetricsResponse{}), nil
}
