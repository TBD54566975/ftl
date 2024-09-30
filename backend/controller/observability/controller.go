package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	controllerPollingOperation = "ftl.controller.poll"
	operation                  = "operation"
)

type ControllerTracing struct {
	polling trace.Tracer
}

func initControllerTracing() *ControllerTracing {
	provider := otel.GetTracerProvider()
	result := &ControllerTracing{
		polling: provider.Tracer(controllerPollingOperation),
	}

	return result
}

func (m *ControllerTracing) BeginSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	attrs := []attribute.KeyValue{
		attribute.String(operation, name),
	}
	return m.polling.Start(ctx, controllerPollingOperation, trace.WithAttributes(attrs...))
}
