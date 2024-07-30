package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/observability"
	"github.com/alecthomas/types/optional"
)

const (
	asyncCallMeterName                = "ftl.async_call"
	asyncCallOriginAttr               = "ftl.async_call.origin"
	asyncCallVerbRefAttr              = "ftl.async_call.verb.ref"
	asyncCallTimeSinceScheduledAtAttr = "ftl.async_call.time_since_scheduled_at_ms"
	asyncCallExecFailureModeAttr      = "ftl.async_call.execution.failure_mode"
)

type AsyncCallMetrics struct {
	meter     metric.Meter
	acquired  metric.Int64Counter
	executed  metric.Int64Counter
	completed metric.Int64Counter
}

func initAsyncCallMetrics() (*AsyncCallMetrics, error) {
	result := &AsyncCallMetrics{}
	var errs error
	var err error

	result.meter = otel.Meter(asyncCallMeterName)

	counterName := fmt.Sprintf("%s.acquired", asyncCallMeterName)
	if result.acquired, err = result.meter.Int64Counter(
		counterName,
		metric.WithUnit("1"),
		metric.WithDescription("the number of times that the controller tries acquiring an async call")); err != nil {
		errs = handleInitCounterError(errs, err, counterName)
		result.acquired = noop.Int64Counter{}
	}

	counterName = fmt.Sprintf("%s.executed", asyncCallMeterName)
	if result.executed, err = result.meter.Int64Counter(
		counterName,
		metric.WithUnit("1"),
		metric.WithDescription("the number of times that the controller tries executing an async call")); err != nil {
		errs = handleInitCounterError(errs, err, counterName)
		result.executed = noop.Int64Counter{}
	}

	counterName = fmt.Sprintf("%s.completed", asyncCallMeterName)
	if result.completed, err = result.meter.Int64Counter(
		counterName,
		metric.WithUnit("1"),
		metric.WithDescription("the number of times that the controller tries completing an async call")); err != nil {
		errs = handleInitCounterError(errs, err, counterName)
		result.completed = noop.Int64Counter{}
	}

	return result, errs
}

func (m *AsyncCallMetrics) Acquired(ctx context.Context, verb schema.RefKey, origin string, scheduledAt time.Time, maybeErr error) {
	m.acquired.Add(ctx, 1, metric.WithAttributes(extractAsyncCallAndMaybeErrAttrs(verb, origin, scheduledAt, maybeErr)...))
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
	m.completed.Add(ctx, 1, metric.WithAttributes(extractAsyncCallAndMaybeErrAttrs(verb, origin, scheduledAt, maybeErr)...))
}

func extractAsyncCallAndMaybeErrAttrs(verb schema.RefKey, origin string, scheduledAt time.Time, maybeErr error) []attribute.KeyValue {
	attrs := extractAsyncCallAttrs(verb, origin, scheduledAt)
	return append(attrs, attribute.Bool(observability.StatusSucceededAttribute, maybeErr == nil))
}

func extractAsyncCallAttrs(verb schema.RefKey, origin string, scheduledAt time.Time) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String(observability.ModuleNameAttribute, verb.Module),
		attribute.String(asyncCallVerbRefAttr, verb.String()),
		attribute.String(asyncCallOriginAttr, origin),
		attribute.Int64(asyncCallTimeSinceScheduledAtAttr, time.Since(scheduledAt).Milliseconds()),
	}
}
