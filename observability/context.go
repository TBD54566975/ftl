package observability

import (
	"context"

	"go.opentelemetry.io/otel/metric"
)

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
