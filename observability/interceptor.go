package observability

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/TBD54566975/ftl/internal/log"
)

const ftlDirectRoutingHeader = "FTL-Direct"
const ftlVerbHeader = "FTL-Verb"

const (
	instrumentationName = "ftl"
	verbKey             = "ftl.verb"
	durationFormat      = "%s.duration"
	requestCountFormat  = "%s.request.count"
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

		if verbRef := req.Header().Get(ftlVerbHeader); verbRef != "" {
			verb, module, err := parseRef(verbRef)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			meter := otel.GetMeterProvider().Meter(
				instrumentationName,
				metric.WithInstrumentationAttributes(attribute.String("ftl.verbRef", verbRef)),
				metric.WithInstrumentationAttributes(attribute.String("ftl.verb", verb)),
				metric.WithInstrumentationAttributes(attribute.String("ftl.module", module)),
			)

			counter, err := meter.Int64Counter(fmt.Sprintf(durationFormat, verbRef))
			if err != nil {
				return nil, errors.WithStack(err)
			}
			counter.Add(ctx, 1)

			histogram, err := meter.Int64Histogram(fmt.Sprintf(durationFormat, verbRef), metric.WithUnit(unitMilliseconds))
			if err != nil {
				return nil, errors.WithStack(err)
			}
			histogram.Record(ctx, time.Since(start).Milliseconds())
		}

		if err != nil {
			err = errors.WithStack(err)
			logger.Errorf(err, "Unary RPC failed: %s: %s", req.Spec().Procedure)
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
		return nil
	}
}

func parseRef(s string) (string, string, error) {
	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		return "", "", errors.Errorf("invalid reference %q", s)
	}
	return parts[0], parts[1], nil
}
