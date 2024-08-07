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

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/observability"
)

const (
	asyncCallMeterName                      = "ftl.async_call"
	asyncCallOriginAttr                     = "ftl.async_call.origin"
	asyncCallVerbRefAttr                    = "ftl.async_call.verb.ref"
	asyncCallTimeSinceScheduledAtBucketAttr = "ftl.async_call.time_since_scheduled_at_ms.bucket"
	asyncCallRemainingAttemptsAttr          = "ftl.async_call.remaining_attempts"
	asyncCallExecFailureModeAttr            = "ftl.async_call.execution.failure_mode"
)

type AsyncCallMetrics struct {
	created      metric.Int64Counter
	acquired     metric.Int64Counter
	executed     metric.Int64Counter
	completed    metric.Int64Counter
	msToComplete metric.Int64Histogram
	queueDepth   metric.Int64Gauge
}

func initAsyncCallMetrics() (*AsyncCallMetrics, error) {
	result := &AsyncCallMetrics{
		created:      noop.Int64Counter{},
		acquired:     noop.Int64Counter{},
		executed:     noop.Int64Counter{},
		completed:    noop.Int64Counter{},
		msToComplete: noop.Int64Histogram{},
		queueDepth:   noop.Int64Gauge{},
	}

	var err error
	meter := otel.Meter(asyncCallMeterName)

	signalName := fmt.Sprintf("%s.created", asyncCallMeterName)
	if result.created, err = meter.Int64Counter(signalName, metric.WithUnit("1"),
		metric.WithDescription("the number of times that an async call was created")); err != nil {
		return nil, wrapErr(signalName, err)
	}

	signalName = fmt.Sprintf("%s.acquired", asyncCallMeterName)
	if result.acquired, err = meter.Int64Counter(signalName, metric.WithUnit("1"),
		metric.WithDescription("the number of times that the controller tries acquiring an async call")); err != nil {
		return nil, wrapErr(signalName, err)
	}

	signalName = fmt.Sprintf("%s.executed", asyncCallMeterName)
	if result.executed, err = meter.Int64Counter(signalName, metric.WithUnit("1"),
		metric.WithDescription("the number of times that the controller tries executing an async call")); err != nil {
		return nil, wrapErr(signalName, err)
	}

	signalName = fmt.Sprintf("%s.completed", asyncCallMeterName)
	if result.completed, err = meter.Int64Counter(signalName, metric.WithUnit("1"),
		metric.WithDescription("the number of times that the controller tries completing an async call")); err != nil {
		return nil, wrapErr(signalName, err)
	}

	signalName = fmt.Sprintf("%s.ms_to_complete", asyncCallMeterName)
	if result.msToComplete, err = meter.Int64Histogram(signalName, metric.WithUnit("ms"),
		metric.WithDescription("duration in ms to complete an async call, from the earliest time it was scheduled to execute")); err != nil {
		return nil, wrapErr(signalName, err)
	}

	signalName = fmt.Sprintf("%s.queue_depth", asyncCallMeterName)
	if result.queueDepth, err = meter.Int64Gauge(signalName, metric.WithUnit("1"),
		metric.WithDescription("number of async calls queued up")); err != nil {
		return nil, wrapErr(signalName, err)
	}

	return result, nil
}

func (m *AsyncCallMetrics) Created(ctx context.Context, verb schema.RefKey, origin string, remainingAttempts int64, maybeErr error) {
	attrs := extractRefAttrs(verb, origin)
	attrs = append(attrs, observability.SuccessOrFailureStatusAttr(maybeErr == nil))
	attrs = append(attrs, attribute.Int64(asyncCallRemainingAttemptsAttr, remainingAttempts))

	m.created.Add(ctx, 1, metric.WithAttributes(attrs...))
}

func (m *AsyncCallMetrics) RecordQueueDepth(ctx context.Context, queueDepth int64) {
	m.queueDepth.Record(ctx, queueDepth)
}

func (m *AsyncCallMetrics) Acquired(ctx context.Context, verb schema.RefKey, origin string, scheduledAt time.Time, maybeErr error) {
	attrs := extractAsyncCallAttrs(verb, origin, scheduledAt)
	attrs = append(attrs, observability.SuccessOrFailureStatusAttr(maybeErr == nil))
	m.acquired.Add(ctx, 1, metric.WithAttributes(attrs...))
}

func (m *AsyncCallMetrics) Executed(ctx context.Context, verb schema.RefKey, origin string, scheduledAt time.Time, maybeFailureMode optional.Option[string]) {
	attrs := extractAsyncCallAttrs(verb, origin, scheduledAt)

	failureMode, ok := maybeFailureMode.Get()
	attrs = append(attrs, observability.SuccessOrFailureStatusAttr(!ok))
	if ok {
		attrs = append(attrs, attribute.String(asyncCallExecFailureModeAttr, failureMode))
	}

	m.executed.Add(ctx, 1, metric.WithAttributes(attrs...))
}

func (m *AsyncCallMetrics) Completed(ctx context.Context, verb schema.RefKey, origin string, scheduledAt time.Time, queueDepth int64, maybeErr error) {
	msToComplete := timeSinceMS(scheduledAt)

	attrs := extractRefAttrs(verb, origin)
	attrs = append(attrs, observability.SuccessOrFailureStatusAttr(maybeErr == nil))
	m.msToComplete.Record(ctx, msToComplete, metric.WithAttributes(attrs...))

	attrs = append(attrs, attribute.String(asyncCallTimeSinceScheduledAtBucketAttr, logBucket(8, msToComplete)))
	m.completed.Add(ctx, 1, metric.WithAttributes(attrs...))

	m.queueDepth.Record(ctx, queueDepth)
}

func extractAsyncCallAttrs(verb schema.RefKey, origin string, scheduledAt time.Time) []attribute.KeyValue {
	return append(extractRefAttrs(verb, origin), attribute.String(asyncCallTimeSinceScheduledAtBucketAttr, logBucket(8, timeSinceMS(scheduledAt))))
}

func extractRefAttrs(verb schema.RefKey, origin string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String(observability.ModuleNameAttribute, verb.Module),
		attribute.String(asyncCallVerbRefAttr, verb.String()),
		attribute.String(asyncCallOriginAttr, origin),
	}
}
