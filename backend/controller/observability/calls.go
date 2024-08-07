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

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/observability"
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
}

func initCallMetrics() (*CallMetrics, error) {
	result := &CallMetrics{
		requests:     noop.Int64Counter{},
		msToComplete: noop.Int64Histogram{},
	}

	var err error
	meter := otel.Meter(callMeterName)

	signalName := fmt.Sprintf("%s.requests", callMeterName)
	if result.requests, err = meter.Int64Counter(signalName, metric.WithUnit("1"),
		metric.WithDescription("the number of times that the FTL controller receives a verb call request")); err != nil {
		return nil, wrapErr(signalName, err)
	}

	signalName = fmt.Sprintf("%s.ms_to_complete", callMeterName)
	if result.msToComplete, err = meter.Int64Histogram(signalName, metric.WithUnit("ms"),
		metric.WithDescription("duration in ms to complete a verb call")); err != nil {
		return nil, wrapErr(signalName, err)
	}

	return result, nil
}

func (m *CallMetrics) Request(ctx context.Context, verb *schemapb.Ref, startTime time.Time, maybeFailureMode optional.Option[string]) {
	attrs := []attribute.KeyValue{
		attribute.String(observability.ModuleNameAttribute, verb.Module),
		attribute.String(callVerbRefAttr, schema.RefFromProto(verb).String()),
	}

	failureMode, ok := maybeFailureMode.Get()
	attrs = append(attrs, observability.SuccessOrFailureStatus(!ok))
	if ok {
		attrs = append(attrs, attribute.String(callFailureModeAttr, failureMode))
	}

	msToComplete := timeSinceMS(startTime)
	m.msToComplete.Record(ctx, msToComplete, metric.WithAttributes(attrs...))

	attrs = append(attrs, attribute.String(callRunTimeBucketAttr, logBucket(2, msToComplete)))
	m.requests.Add(ctx, 1, metric.WithAttributes(attrs...))
}
