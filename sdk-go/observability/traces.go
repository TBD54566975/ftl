package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	// TODO(wb): Get the verb name from the context.
	return otel.GetTracerProvider().Tracer("verbName").Start(ctx, name, opts...)
}
