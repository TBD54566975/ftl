package observability

import (
	"context"
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"

	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/observability"
)

const (
	callMeterName         = "ftl.call"
	callVerbRefAttr       = "ftl.call.verb.ref"
	callFailureModeAttr   = "ftl.call.failure_mode"
	callRunTimeBucketAttr = "ftl.call.run_time_ms.bucket"
)

type CallMetrics struct {
	requests     metric.Int64Counter
	msToComplete metric.Int64Histogram
	callTracer   trace.Tracer
}

func initCallMetrics() *CallMetrics {
	provider := otel.GetTracerProvider()
	result := &CallMetrics{
		requests:     noop.Int64Counter{},
		msToComplete: noop.Int64Histogram{},
		callTracer:   provider.Tracer(callMeterName),
	}

	var err error
	meter := otel.Meter(callMeterName)

	signalName := fmt.Sprintf("%s.requests", callMeterName)
	if result.requests, err = meter.Int64Counter(signalName, metric.WithUnit("1"),
		metric.WithDescription("the number of times that the FTL controller receives a verb call request")); err != nil {
		observability.FatalError(signalName, err)
	}

	signalName = fmt.Sprintf("%s.ms_to_complete", callMeterName)
	if result.msToComplete, err = meter.Int64Histogram(signalName, metric.WithUnit("ms"),
		metric.WithDescription("duration in ms to complete a verb call")); err != nil {
		observability.FatalError(signalName, err)
	}

	return result
}

func (m *CallMetrics) BeginSpan(ctx context.Context, verb *schemapb.Ref) (context.Context, trace.Span) {
	attrs := []attribute.KeyValue{
		attribute.String(callVerbRefAttr, schema.RefFromProto(verb).String()),
	}
	return observability.AddSpanToLogger(m.callTracer.Start(ctx, callMeterName, trace.WithAttributes(attrs...)))
}
func (m *CallMetrics) Request(ctx context.Context, verb *schemapb.Ref, startTime time.Time, maybeFailureMode optional.Option[string]) {
	attrs := []attribute.KeyValue{
		attribute.String(observability.ModuleNameAttribute, verb.Module),
		attribute.String(callVerbRefAttr, schema.RefFromProto(verb).String()),
	}

	failureMode, ok := maybeFailureMode.Get()
	attrs = append(attrs, observability.SuccessOrFailureStatusAttr(!ok))
	if ok {
		attrs = append(attrs, attribute.String(callFailureModeAttr, failureMode))
	}

	msToComplete := observability.TimeSinceMS(startTime)
	m.msToComplete.Record(ctx, msToComplete, metric.WithAttributes(attrs...))

	attrs = append(attrs, attribute.String(callRunTimeBucketAttr, observability.LogBucket(4, msToComplete, optional.Some(3), optional.Some(7))))
	m.requests.Add(ctx, 1, metric.WithAttributes(attrs...))
}
