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
	asyncCallMeterName                = "ftl.async_call"
	asyncCallOriginAttr               = "ftl.async_call.origin"
	asyncCallVerbRefAttr              = "ftl.async_call.verb.ref"
	asyncCallTimeSinceScheduledAtAttr = "ftl.async_call.time_since_scheduled_at_ms"
	asyncCallExecFailureModeAttr      = "ftl.async_call.execution.failure_mode"
)

type AsyncCallMetrics struct {
	acquired     metric.Int64Counter
	executed     metric.Int64Counter
	completed    metric.Int64Counter
	msToComplete metric.Int64Histogram
}

func initAsyncCallMetrics() (*AsyncCallMetrics, error) {
	result := &AsyncCallMetrics{
		acquired:     noop.Int64Counter{},
		executed:     noop.Int64Counter{},
		completed:    noop.Int64Counter{},
		msToComplete: noop.Int64Histogram{},
	}

	var err error
	meter := otel.Meter(asyncCallMeterName)

	signalName := fmt.Sprintf("%s.acquired", asyncCallMeterName)
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

	return result, nil
}

func wrapErr(signalName string, err error) error {
	return fmt.Errorf("failed to create %q signal: %w", signalName, err)
}

func (m *AsyncCallMetrics) Acquired(ctx context.Context, verb schema.RefKey, origin string, scheduledAt time.Time, maybeErr error) {
	attrs := extractAsyncCallAttrs(verb, origin, scheduledAt)
	attrs = append(attrs, attribute.Bool(observability.StatusSucceededAttribute, maybeErr == nil))
	m.acquired.Add(ctx, 1, metric.WithAttributes(attrs...))
}

func (m *AsyncCallMetrics) Executed(ctx context.Context, verb schema.RefKey, origin string, scheduledAt time.Time, maybeFailureMode optional.Option[string]) {
	attrs := extractAsyncCallAttrs(verb, origin, scheduledAt)

	failureMode, ok := maybeFailureMode.Get()
	attrs = append(attrs, attribute.Bool(observability.StatusSucceededAttribute, !ok))
	if ok {
		attrs = append(attrs, attribute.String(asyncCallExecFailureModeAttr, failureMode))
	}

	m.executed.Add(ctx, 1, metric.WithAttributes(attrs...))
}

func (m *AsyncCallMetrics) Completed(ctx context.Context, verb schema.RefKey, origin string, scheduledAt time.Time, maybeErr error) {
	msToComplete := timeSinceMS(scheduledAt)

	attrs := extractRefAttrs(verb, origin)
	attrs = append(attrs, attribute.Bool(observability.StatusSucceededAttribute, maybeErr == nil))
	m.msToComplete.Record(ctx, msToComplete, metric.WithAttributes(attrs...))

	attrs = append(attrs, attribute.Int64(asyncCallTimeSinceScheduledAtAttr, msToComplete))
	m.completed.Add(ctx, 1, metric.WithAttributes(attrs...))
}

func extractAsyncCallAttrs(verb schema.RefKey, origin string, scheduledAt time.Time) []attribute.KeyValue {
	return append(extractRefAttrs(verb, origin), attribute.Int64(asyncCallTimeSinceScheduledAtAttr, timeSinceMS(scheduledAt)))
}

func extractRefAttrs(verb schema.RefKey, origin string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String(observability.ModuleNameAttribute, verb.Module),
		attribute.String(asyncCallVerbRefAttr, verb.String()),
		attribute.String(asyncCallOriginAttr, origin),
	}
}
