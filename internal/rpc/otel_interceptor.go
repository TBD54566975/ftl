package rpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/TBD54566975/ftl/internal/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

const (
	otelFtlRequestKey = "ftl.requestKey"
	otelFtlVerbRef    = "ftl.verb.ref"
	otelFtlVerbModule = "ftl.verb.module"
	otelFtlEventName  = "ftl.message"
)

func OtelInterceptor() connect.Interceptor {
	return &otelInterceptor{}
}

type otelInterceptor struct{}

func getAttributes(ctx context.Context, rpcSystemKey string) []attribute.KeyValue {
	logger := log.FromContext(ctx)
	attributes := []attribute.KeyValue{
		semconv.RPCSystemKey.String(rpcSystemKey),
	}
	requestKey, err := RequestKeyFromContext(ctx)
	if err != nil {
		logger.Debugf("failed to get request key: %s", err)
	}
	if key, ok := requestKey.Get(); ok {
		attributes = append(attributes, attribute.String(otelFtlRequestKey, key.String()))
	}
	if verb, ok := VerbFromContext(ctx); ok {
		attributes = append(attributes, attribute.String(otelFtlVerbRef, verb.String()))
		attributes = append(attributes, attribute.String(otelFtlVerbModule, verb.Module))
	}
	return attributes
}

func (i *otelInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		requestStartTime := time.Now()
		isClient := request.Spec().IsClient
		name := strings.TrimLeft(request.Spec().Procedure, "/")

		spanKind := trace.SpanKindClient
		requestSpan, responseSpan := semconv.MessageTypeSent, semconv.MessageTypeReceived
		if !isClient {
			spanKind = trace.SpanKindServer
			requestSpan, responseSpan = semconv.MessageTypeReceived, semconv.MessageTypeSent
		}

		attributes := getAttributes(ctx, request.Peer().Protocol)
		traceOpts := []trace.SpanStartOption{
			trace.WithAttributes(attributes...),
			trace.WithSpanKind(spanKind),
		}
		tracer := otel.GetTracerProvider().Tracer(request.Spec().Procedure)
		ctx, span := tracer.Start(ctx, name, traceOpts...)
		defer span.End()

		var requestSize int
		if request != nil {
			if msg, ok := request.Any().(proto.Message); ok {
				requestSize = proto.Size(msg)
			}
		}

		span.AddEvent(otelFtlEventName,
			trace.WithAttributes(
				requestSpan,
				semconv.MessageIDKey.Int(1),
				semconv.MessageUncompressedSizeKey.Int(requestSize),
			),
		)
		statusCode := attribute.Int64("rpc.grpc.status_code", 0)
		response, err := next(ctx, request)
		if err != nil {
			statusCode = attribute.Int64(string(statusCode.Key), int64(connect.CodeOf(err)))
		}
		attributes = append(attributes, statusCode)
		var responseSize int
		if err == nil {
			if msg, ok := response.Any().(proto.Message); ok {
				responseSize = proto.Size(msg)
			}
		}
		span.AddEvent(otelFtlEventName,
			trace.WithAttributes(
				responseSpan,
				semconv.MessageIDKey.Int(1),
				semconv.MessageUncompressedSizeKey.Int(responseSize),
			),
		)
		span.SetAttributes(attributes...)

		instrumentation, err := createInstruments(otel.GetMeterProvider().Meter("ftl.rpc.unary"))
		if err != nil {
			return nil, fmt.Errorf("failed to create instruments: %w", err)
		}
		instrumentation.duration.Record(ctx, time.Since(requestStartTime).Milliseconds(), metric.WithAttributes(attributes...))
		instrumentation.requestSize.Record(ctx, int64(requestSize), metric.WithAttributes(attributes...))
		instrumentation.responseSize.Record(ctx, int64(responseSize), metric.WithAttributes(attributes...))
		return response, err
	}
}

func (i *otelInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		return next(ctx, spec)
	}
}

func (i *otelInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		return next(ctx, conn)
	}
}

type instruments struct {
	duration     metric.Int64Histogram
	requestSize  metric.Int64Histogram
	responseSize metric.Int64Histogram
}

func createInstruments(meter metric.Meter) (instruments, error) {
	duration, err := meter.Int64Histogram("duration", metric.WithUnit("ms"))
	if err != nil {
		return instruments{}, fmt.Errorf("failed to create duration metric: %w", err)
	}
	requestSize, err := meter.Int64Histogram("request.size", metric.WithUnit("By"))
	if err != nil {
		return instruments{}, fmt.Errorf("failed to create request size metric: %w", err)
	}
	responseSize, err := meter.Int64Histogram("response.size", metric.WithUnit("By"))
	if err != nil {
		return instruments{}, fmt.Errorf("failed to create response size metric: %w", err)
	}
	return instruments{
		duration:     duration,
		requestSize:  requestSize,
		responseSize: responseSize,
	}, nil
}
