package observability

import (
	"context"
	"net/http"
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
	sourceVerbKey       = "ftl.source.verb"
	sourceModuleKey     = "ftl.source.module"
	destVerbKey         = "ftl.dest.verb"
	destModuleKey       = "ftl.dest.module"
	callLatency         = "call.latency"
	callRequestCount    = "call.request.count"
	unitMilliseconds    = "ms"
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

		if verb := req.Header().Get(headers.VerbHeader); verb != "" {
			metricsErr := i.recordVerbCallMetrics(ctx, verb, start, req.Header())
			if metricsErr != nil {
				logger.Errorf(metricsErr, "Failed to record metrics for verb: %s", verb)
			}
		}

		if err != nil {
			err = errors.WithStack(err)
			logger.Errorf(err, "Unary RPC failed: %s", req.Spec().Procedure)
			return nil, err
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

func (i *Interceptor) recordVerbCallMetrics(ctx context.Context, verb string, start time.Time, header http.Header) error {
	sourceVerbRef, err := schema.ParseRef(verb)
	if err != nil {
		return errors.WithStack(err)
	}

	caller, err := headers.GetCaller(header)
	if err != nil {
		return errors.WithStack(err)
	}

	attributes := []attribute.KeyValue{
		attribute.String(verbRefKey, verb),
		attribute.String(sourceVerbKey, sourceVerbRef.Name),
		attribute.String(sourceModuleKey, sourceVerbRef.Module),
		attribute.String(destVerbKey, caller.Value().Name),
		attribute.String(destModuleKey, caller.Value().Module),
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
