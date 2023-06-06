package observability

import (
	"context"

	"github.com/TBD54566975/ftl/internal/rpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	verb, _ := rpc.VerbFromContext(ctx)
	return otel.GetTracerProvider().Tracer(verb.Name).Start(ctx, name, opts...)
}
