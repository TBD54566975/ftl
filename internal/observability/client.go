package observability

import (
	"context"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
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

	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(otelTraceExporter),
		sdktrace.WithResource(res),
		sdktrace.WithIDGenerator(newTraceIDGenerator()),
	)
	otel.SetTracerProvider(traceProvider)

	return nil
}

type traceIDGenerator struct {
	sync.Mutex
	randSource *rand.Rand
}

var _ sdktrace.IDGenerator = &traceIDGenerator{}

func (t *traceIDGenerator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	logger := log.FromContext(ctx)
	t.Lock()
	defer t.Unlock()

	var tid trace.TraceID
	if rk, err := rpc.RequestKeyFromContext(ctx); err == nil {
		if k, ok := rk.Get(); ok {
			hash := sha256.Sum256([]byte(k.Payload.Key))
			copy(tid[:], hash[:16])
			logger.Debugf("Mapping requestKey to traceId %v --> %s", rk, hex.EncodeToString(hash[:16]))
		}
	}
	if !tid.IsValid() {
		tid = trace.TraceID{}
		_, _ = t.randSource.Read(tid[:])
	}
	sid := trace.SpanID{}
	_, _ = t.randSource.Read(sid[:])
	return tid, sid
}

func (t *traceIDGenerator) NewSpanID(ctx context.Context, traceID trace.TraceID) trace.SpanID {
	t.Lock()
	defer t.Unlock()
	sid := trace.SpanID{}
	_, _ = t.randSource.Read(sid[:])
	return sid
}

func newTraceIDGenerator() sdktrace.IDGenerator {
	tig := &traceIDGenerator{}
	var rngSeed int64
	err := binary.Read(crand.Reader, binary.LittleEndian, &rngSeed)
	if err != nil {
		panic("failed to seed TraceID generator:" + err.Error())
	}
	tig.randSource = rand.New(rand.NewSource(rngSeed)) //nolint:gosec
	return tig
}
