package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/TBD54566975/ftl/backend/common/rpc"
)

func TracerWithVerb(ctx context.Context) trace.Tracer {
	verb, ok := rpc.VerbFromContext(ctx)
	if !ok {
		panic("traces: no verb in context")
	}
	return otel.GetTracerProvider().Tracer(verb.Name)
}

func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return TracerWithVerb(ctx).Start(ctx, name, opts...)
}
