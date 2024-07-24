package observability

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/TBD54566975/ftl/internal/log"
)

const schemaURL = semconv.SchemaURL

type ExportOTELFlag bool

func (e *ExportOTELFlag) UnmarshalText(text []byte) error {
	// Default behaviour of Kong is to use strconv.ParseBool, but we want to be less strict.
	v := strings.ToLower(string(text))
	*e = ExportOTELFlag(!(v == "false" || v == "0" || v == "no" || v == ""))
	return nil
}

type Config struct {
	LogLevel   log.Level      `default:"error" help:"OTEL log level." env:"FTL_O11Y_LOG_LEVEL"`
	ExportOTEL ExportOTELFlag `help:"Export observability data to OTEL." env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
}

func Init(ctx context.Context, serviceName, serviceVersion string, config Config) error {
	logger := log.FromContext(ctx)
	if !config.ExportOTEL {
		logger.Tracef("OTEL export is disabled, set OTEL_EXPORTER_OTLP_ENDPOINT to enable")
		return nil
	}

	logger.Debugf("OTEL is enabled, exporting to %s at log level %s", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"), os.Getenv("FTL_O11Y_LOG_LEVEL"))

	otelLogger := NewOtelLogger(logger, config.LogLevel)
	otel.SetLogger(otelLogger)
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) { logger.Errorf(err, "OTEL") }))

	res, err := resource.Merge(resource.Default(),
		resource.NewWithAttributes(
			schemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		))
	if err != nil {
		return fmt.Errorf("failed to create OTEL resource: %w", err)
	}

	otelMetricExporter, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		return fmt.Errorf("failed to create OTEL metric exporter: %w", err)
	}

	meterProvider := metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(otelMetricExporter)), metric.WithResource(res))
	otel.SetMeterProvider(meterProvider)

	otelTraceExporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return fmt.Errorf("failed to create OTEL trace exporter: %w", err)
	}
	traceProvider := trace.NewTracerProvider(trace.WithBatcher(otelTraceExporter), trace.WithResource(res))
	otel.SetTracerProvider(traceProvider)

	return nil
}
