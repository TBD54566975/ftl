package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/observability"
	"github.com/alecthomas/types/optional"
)

const (
	ingressMeterName         = "ftl.ingress"
	ingressMethodAttr        = "ftl.ingress.method"
	ingressPathAttr          = "ftl.ingress.path"
	ingressVerbRefAttr       = "ftl.ingress.verb.ref"
	ingressFailureModeAttr   = "ftl.ingress.failure_mode"
	ingressRunTimeBucketAttr = "ftl.ingress.run_time_ms.bucket"
)

type IngressMetrics struct {
	requests     metric.Int64Counter
	msToComplete metric.Int64Histogram
}

func initIngressMetrics() (*IngressMetrics, error) {
	result := &IngressMetrics{
		requests:     noop.Int64Counter{},
		msToComplete: noop.Int64Histogram{},
	}

	var err error
	meter := otel.Meter(ingressMeterName)

	signalName := fmt.Sprintf("%s.requests", ingressMeterName)
	if result.requests, err = meter.Int64Counter(signalName, metric.WithUnit("1"),
		metric.WithDescription("the number of ingress requests that the FTL controller receives")); err != nil {
		return nil, wrapErr(signalName, err)
	}

	signalName = fmt.Sprintf("%s.ms_to_complete", ingressMeterName)
	if result.msToComplete, err = meter.Int64Histogram(signalName, metric.WithUnit("ms"),
		metric.WithDescription("duration in ms to complete an ingress request")); err != nil {
		return nil, wrapErr(signalName, err)
	}

	return result, nil
}

func (m *IngressMetrics) Request(ctx context.Context, method string, path string, verb optional.Option[*schemapb.Ref], startTime time.Time, failureMode optional.Option[string]) {
	attrs := []attribute.KeyValue{
		attribute.String(ingressMethodAttr, method),
		attribute.String(ingressPathAttr, path),
	}
	if v, ok := verb.Get(); ok {
		attrs = append(attrs,
			attribute.String(observability.ModuleNameAttribute, v.Module),
			attribute.String(ingressVerbRefAttr, schema.RefFromProto(v).String()))
	}
	if f, ok := failureMode.Get(); ok {
		attrs = append(attrs, attribute.String(ingressFailureModeAttr, f))
	}

	msToComplete := timeSinceMS(startTime)
	m.msToComplete.Record(ctx, msToComplete, metric.WithAttributes(attrs...))

	attrs = append(attrs, attribute.String(ingressRunTimeBucketAttr, logBucket(2, msToComplete)))
	m.requests.Add(ctx, 1, metric.WithAttributes(attrs...))
}
