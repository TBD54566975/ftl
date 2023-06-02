package observability

import (
	"context"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

type tracerProviderKey struct{}

// ContextWithTracerProvider returns a context with an otel TracerProvider
func ContextWithTracerProvider(ctx context.Context, tp *trace.TracerProvider) context.Context {
	return context.WithValue(ctx, tracerProviderKey{}, tp)
}

// TracerProviderFromContext returns the otel TracerProvider or panics
func TracerProviderFromContext(ctx context.Context) *trace.TracerProvider {
	value := ctx.Value(tracerProviderKey{})
	if value == nil {
		panic("no otel TracerProvider in context")
	}
	return value.(*trace.TracerProvider) //nolint:forcetypeassert
}

type meterProviderKey struct{}

// ContextWithMeterProvider returns a context with an otel MeterProvider
func ContextWithMeterProvider(ctx context.Context, mp metric.MeterProvider) context.Context {
	return context.WithValue(ctx, meterProviderKey{}, mp)
}

// MeterProviderFromContext returns the otel MeterProvider or panics
func MeterProviderFromContext(ctx context.Context) metric.MeterProvider {
	value := ctx.Value(meterProviderKey{})
	if value == nil {
		panic("no otel MeterProvider in context")
	}
	return value.(metric.MeterProvider) //nolint:forcetypeassert
}
