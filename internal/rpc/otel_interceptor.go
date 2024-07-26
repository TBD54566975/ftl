package rpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/TBD54566975/ftl/internal/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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

var instruments instrumentation

func init() {
	instruments = createInstruments(otel.GetMeterProvider().Meter("ftl.rpc.unary"))
}

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

		statusCodeKey := fmt.Sprintf("rpc.%s.status_code", request.Peer().Protocol)
		statusCode := attribute.Int64(statusCodeKey, 0)
		response, err := next(ctx, request)
		if err != nil {
			statusCode = attribute.Int64(statusCodeKey, int64(connect.CodeOf(err)))
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

		instruments.duration.Record(ctx, time.Since(requestStartTime).Milliseconds(), metric.WithAttributes(attributes...))
		instruments.requestSize.Record(ctx, int64(requestSize), metric.WithAttributes(attributes...))
		instruments.requestsPerRPC.Record(ctx, 1, metric.WithAttributes(attributes...))
		instruments.responseSize.Record(ctx, int64(responseSize), metric.WithAttributes(attributes...))
		instruments.responsesPerRPC.Record(ctx, 1, metric.WithAttributes(attributes...))
		return response, err
	}
}

func (i *otelInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		requestStartTime := time.Now()
		name := strings.TrimLeft(spec.Procedure, "/")

		attributes := getAttributes(ctx, "grpc")
		traceOpts := []trace.SpanStartOption{
			trace.WithAttributes(attributes...),
			trace.WithSpanKind(trace.SpanKindClient),
		}
		tracer := otel.GetTracerProvider().Tracer(spec.Procedure)
		ctx, span := tracer.Start(ctx, name, traceOpts...) // nolint:spancheck
		conn := next(ctx, spec)

		// TODO: add conn.RequestHeader() to span.SetAttributes(...)
		state := &streamingState{
			spec:        spec,
			protocol:    conn.Peer().Protocol,
			attributes:  attributes,
			receiveSize: instruments.responseSize,
			sendSize:    instruments.requestSize,
		}

		return &streamingClientInterceptor{ // nolint:spancheck
			StreamingClientConn: conn,
			ctx:                 ctx,
			state:               state,
			onClose: func() {
				statusCodeKey := fmt.Sprintf("rpc.%s.status_code", conn.Peer().Protocol)
				statusCode := attribute.Int64(statusCodeKey, 0)
				if state.error != nil {
					statusCode = attribute.Int64(statusCodeKey, int64(connect.CodeOf(state.error)))
					span.SetStatus(codes.Error, state.error.Error())
				}
				// TODO: add header attributes
				state.attributes = append(state.attributes, statusCode)
				span.SetAttributes(state.attributes...)
				span.End()
				instruments.requestsPerRPC.Record(ctx, state.sentCounter, metric.WithAttributes(state.attributes...))
				instruments.responsesPerRPC.Record(ctx, state.receivedCounter, metric.WithAttributes(state.attributes...))
				instruments.duration.Record(ctx, time.Since(requestStartTime).Milliseconds(), metric.WithAttributes(state.attributes...))
			},
		}
	}
}

func (i *otelInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		return next(ctx, conn)
	}
}

type instrumentation struct {
	duration        metric.Int64Histogram
	requestSize     metric.Int64Histogram
	responseSize    metric.Int64Histogram
	requestsPerRPC  metric.Int64Histogram
	responsesPerRPC metric.Int64Histogram
}

func createInstruments(meter metric.Meter) instrumentation {
	duration, err := meter.Int64Histogram("duration", metric.WithUnit("ms"))
	if err != nil {
		panic(fmt.Errorf("failed to create duration metric: %w", err))
	}
	requestSize, err := meter.Int64Histogram("request.size", metric.WithUnit("By"))
	if err != nil {
		panic(fmt.Errorf("failed to create request size metric: %w", err))
	}
	responseSize, err := meter.Int64Histogram("response.size", metric.WithUnit("By"))
	if err != nil {
		panic(fmt.Errorf("failed to create response size metric: %w", err))
	}
	requestsPerRPC, err := meter.Int64Histogram("requests_per_rpc", metric.WithUnit("1"))
	if err != nil {
		panic(fmt.Errorf("failed to create requests per rpc metric: %w", err))
	}
	responsesPerRPC, err := meter.Int64Histogram("responses_per_rpc", metric.WithUnit("1"))
	if err != nil {
		panic(fmt.Errorf("failed to create responses per rpc metric: %w", err))
	}
	return instrumentation{
		duration:        duration,
		requestSize:     requestSize,
		responseSize:    responseSize,
		requestsPerRPC:  requestsPerRPC,
		responsesPerRPC: responsesPerRPC,
	}
}

type streamingState struct {
	spec            connect.Spec
	protocol        string
	attributes      []attribute.KeyValue
	error           error
	sentCounter     int64
	receivedCounter int64
	receiveSize     metric.Int64Histogram
	sendSize        metric.Int64Histogram
}

type streamingClientInterceptor struct {
	connect.StreamingClientConn
	mu      sync.Mutex
	ctx     context.Context
	state   *streamingState
	onClose func()
}

func (s *streamingClientInterceptor) Receive(msg any) error {
	err := s.StreamingClientConn.Receive(msg)
	s.mu.Lock()
	defer s.mu.Unlock()
	if errors.Is(err, io.EOF) {
		return err // nolint:wrapcheck
	}
	s.state.receivedCounter++
	if err != nil {
		s.state.error = err
		statusCodeKey := fmt.Sprintf("rpc.%s.status_code", s.Peer().Protocol)
		statusCode := attribute.Int64(statusCodeKey, int64(connect.CodeOf(err)))
		s.state.attributes = append(s.state.attributes, statusCode)
	}
	attrs := append(s.state.attributes, []attribute.KeyValue{ // nolint:gocritic
		semconv.MessageTypeReceived,
		semconv.MessageIDKey.Int64(s.state.receivedCounter),
	}...)
	if protomsg, ok := msg.(proto.Message); ok {
		size := proto.Size(protomsg)
		attrs = append(attrs, semconv.MessageUncompressedSizeKey.Int(size))
		s.state.receiveSize.Record(s.ctx, int64(size), metric.WithAttributes(attrs...))
	}

	span := trace.SpanFromContext(s.ctx)
	span.AddEvent(otelFtlEventName, trace.WithAttributes(attrs...))
	return err // nolint:wrapcheck
}

func (s *streamingClientInterceptor) Send(msg any) error {
	err := s.StreamingClientConn.Send(msg)
	s.mu.Lock()
	defer s.mu.Unlock()
	if errors.Is(err, io.EOF) {
		return err // nolint:wrapcheck
	}
	s.state.sentCounter++
	if err != nil {
		s.state.error = err
		statusCodeKey := fmt.Sprintf("rpc.%s.status_code", s.Peer().Protocol)
		statusCode := attribute.Int64(statusCodeKey, int64(connect.CodeOf(err)))
		s.state.attributes = append(s.state.attributes, statusCode)
	}
	attrs := append(s.state.attributes, []attribute.KeyValue{ // nolint:gocritic
		semconv.MessageTypeSent,
		semconv.MessageIDKey.Int64(s.state.sentCounter),
	}...)
	if protomsg, ok := msg.(proto.Message); ok {
		size := proto.Size(protomsg)
		attrs = append(attrs, semconv.MessageUncompressedSizeKey.Int(size))
		s.state.sendSize.Record(s.ctx, int64(size), metric.WithAttributes(attrs...))
	}

	span := trace.SpanFromContext(s.ctx)
	span.AddEvent(otelFtlEventName, trace.WithAttributes(attrs...))
	return err // nolint:wrapcheck
}

func (s *streamingClientInterceptor) Close() error {
	err := s.StreamingClientConn.CloseResponse()
	s.onClose()
	return err // nolint:wrapcheck
}
