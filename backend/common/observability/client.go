package observability

import (
	"context"

	"github.com/alecthomas/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"

	"github.com/TBD54566975/ftl/backend/common/log"
)

type Config struct {
	LogLevel   log.Level `default:"error" help:"OTEL log level."`
	ExportOTEL bool      `help:"Export metrics to OTEL."`
}

func Init(ctx context.Context, name string, config Config) error {
	logger := log.FromContext(ctx)
	if !config.ExportOTEL {
		logger.Tracef("OTEL metrics export is disabled")
		return nil
	}

	OTELLogger := NewOtelLogger(logger)
	otel.SetLogger(OTELLogger)
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) { logger.Errorf(err, "OTEL") }))

	OTELMetricExporter, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	meterProvider := metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(OTELMetricExporter)))
	otel.SetMeterProvider(meterProvider)

	return nil
}
