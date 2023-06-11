package observability

import (
	"context"
	"net"
	"net/url"
	"time"

	"github.com/alecthomas/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"github.com/TBD54566975/ftl/internal/log"
)

type Config struct {
	Endpoint   *url.URL      `help:"FTL observability endpoint." env:"FTL_OBSERVABILITY_ENDPOINT" placeholder:"URL"`
	Interval   time.Duration `default:"30s" help:"Interval to export metrics." env:"FTL_METRICS_INTERVAL"`
	LogLevel   log.Level     `default:"error" help:"OTEL log level."`
	ExportOTEL bool          `help:"Export metrics to OTEL."`
}

func Init(ctx context.Context, name string, config Config) error {
	if config.Endpoint == nil {
		return errors.Errorf("FTL_OBSERVABILITY_ENDPOINT is required")
	}
	_, _, err := net.SplitHostPort(config.Endpoint.Host)
	if err != nil {
		return errors.WithStack(err)
	}

	logger := log.FromContext(ctx).Sub("otel", config.LogLevel)
	otelLogger := NewOtelLogger(logger)
	otel.SetLogger(otelLogger)
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) { logger.Errorf(err, "OTEL") }))

	ftlExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithEndpoint(config.Endpoint.Host), otlpmetricgrpc.WithInsecure())
	if err != nil {
		return errors.WithStack(err)
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(name),
	)

	metricOptions := []metric.Option{
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(ftlExporter, metric.WithInterval(config.Interval))),
	}
	if config.ExportOTEL {
		otelExporter, err := otlpmetricgrpc.New(ctx)
		if err != nil {
			return errors.WithStack(err)
		}
		metricOptions = append(metricOptions, metric.WithReader(metric.NewPeriodicReader(otelExporter, metric.WithInterval(config.Interval))))
	} else {
		log.FromContext(ctx).Warnf("OTEL metrics export is disabled")
	}

	meterProvider := metric.NewMeterProvider(metricOptions...)
	otel.SetMeterProvider(meterProvider)

	tp := trace.NewTracerProvider(
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(name),
		)),
	)

	// Set the global tracer provider
	otel.SetTracerProvider(tp)
	return nil
}
