package observability

import (
	"context"
	"time"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc/headers"
	"github.com/TBD54566975/ftl/schema"
)

const (
	instrumentationName = "ftl"
	verbRefKey          = "ftl.verb.ref"

	SourceVerbKey    = "ftl.source.verb"   // SourceVerbKey is the key for the source verb.
	SourceModuleKey  = "ftl.source.module" // SourceModuleKey is the key for the source module.
	DestVerbKey      = "ftl.dest.verb"     // DestVerbKey is the key for the destination verb.
	DestModuleKey    = "ftl.dest.module"   // DestModuleKey is the key for the destination module.
	callLatency      = "call.latency"
	callRequestCount = "call.request.count"
	unitMilliseconds = "ms"
)

type Interceptor struct{}

var _ connect.Interceptor = &Interceptor{}

func NewInterceptor() *Interceptor {
	return &Interceptor{}
}

func (i *Interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		start := time.Now()
		logger := log.FromContext(ctx)

		resp, err := next(ctx, req)
		if err != nil {
			err = errors.WithStack(err)
			logger.Errorf(err, "Unary RPC failed: %s", req.Spec().Procedure)
			return nil, err
		}

		callers, err := headers.GetCallers(req.Header())
		if err != nil {
			return nil, errors.WithStack(err)
		}

		err = i.recordVerbCallMetrics(ctx, callers, start)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		return resp, nil
	}
}

func (i *Interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, s connect.Spec) connect.StreamingClientConn {
		return next(ctx, s)
	}
}

func (i *Interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, s connect.StreamingHandlerConn) error {
		return next(ctx, s)
	}
}

func (i *Interceptor) recordVerbCallMetrics(ctx context.Context, callers []*schema.VerbRef, start time.Time) error {
	if len(callers) == 0 {
		return nil // no callers, no metrics
	}
	destRef := callers[len(callers)-1]

	attributes := []attribute.KeyValue{
		attribute.String(verbRefKey, destRef.String()),
		attribute.String(DestVerbKey, destRef.Name),
		attribute.String(DestModuleKey, destRef.Module),
	}

	if len(callers) > 1 {
		sourceRef := callers[len(callers)-2]
		attributes = append(attributes, attribute.String(SourceVerbKey, sourceRef.Name))
		attributes = append(attributes, attribute.String(SourceModuleKey, sourceRef.Module))
	}

	meter := otel.GetMeterProvider().Meter(instrumentationName)

	counter, err := meter.Int64Counter(callRequestCount)
	if err != nil {
		return errors.WithStack(err)
	}
	counter.Add(ctx, 1, metric.WithAttributes(attributes...))

	histogram, err := meter.Int64Histogram(callLatency, metric.WithUnit(unitMilliseconds))
	if err != nil {
		return errors.WithStack(err)
	}
	histogram.Record(ctx, time.Since(start).Milliseconds(), metric.WithAttributes(attributes...))
	return nil
}
