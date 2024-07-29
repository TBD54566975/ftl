package rpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/TBD54566975/ftl/internal/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
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

var clientInstruments instrumentation
var serverInstruments instrumentation

func init() {
	clientInstruments = createInstruments(otel.GetMeterProvider().Meter("ftl.rpc.client"))
	serverInstruments = createInstruments(otel.GetMeterProvider().Meter("ftl.rpc.server"))
}

func getInstruments(isClient bool) instrumentation {
	if isClient {
		return clientInstruments
	}
	return serverInstruments
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
		ctx = propagateOtelHeaders(ctx, request.Spec().IsClient, request.Header())
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

		response, err := next(ctx, request)
		attributes = append(attributes, statusCodeAttribute(request.Peer().Protocol, err))
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
		instruments := getInstruments(isClient)
		instruments.duration.Record(
			ctx,
			time.Since(requestStartTime).Milliseconds(),
			metric.WithAttributes(attributes...))
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

		instruments := getInstruments(spec.IsClient)
		state := &streamingState{
			spec:        spec,
			protocol:    conn.Peer().Protocol,
			attributes:  attributes,
			receiveSize: instruments.responseSize,
			sendSize:    instruments.requestSize,
		}

		return &streamingClientInterceptor{ // nolint:spancheck
			StreamingClientConn: conn,
			receive: func(msg any, conn connect.StreamingClientConn) error {
				return state.receive(ctx, msg, conn)
			},
			send: func(msg any, conn connect.StreamingClientConn) error {
				return state.send(ctx, msg, conn)
			},
			onClose: func() {
				state.attributes = append(
					state.attributes,
					statusCodeAttribute(conn.Peer().Protocol, state.error))
				span.SetAttributes(state.attributes...)
				if state.error != nil {
					span.SetStatus(codes.Error, state.error.Error())
				}
				span.End()
				instruments.requestsPerRPC.Record(
					ctx,
					state.sentCounter,
					metric.WithAttributes(state.attributes...))
				instruments.responsesPerRPC.Record(
					ctx,
					state.receivedCounter,
					metric.WithAttributes(state.attributes...))
				instruments.duration.Record(ctx,
					time.Since(requestStartTime).Milliseconds(),
					metric.WithAttributes(state.attributes...))
			},
		}
	}
}

func (i *otelInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		ctx = propagateOtelHeaders(ctx, conn.Spec().IsClient, conn.RequestHeader())
		requestStartTime := time.Now()
		name := strings.TrimLeft(conn.Spec().Procedure, "/")

		attributes := getAttributes(ctx, "grpc")
		traceOpts := []trace.SpanStartOption{
			trace.WithAttributes(attributes...),
			trace.WithSpanKind(trace.SpanKindServer),
		}
		tracer := otel.GetTracerProvider().Tracer(conn.Spec().Procedure)
		ctx, span := tracer.Start(ctx, name, traceOpts...)
		defer span.End()

		instruments := getInstruments(conn.Spec().IsClient)
		state := &streamingState{
			spec:        conn.Spec(),
			protocol:    conn.Peer().Protocol,
			attributes:  attributes,
			receiveSize: instruments.responseSize,
			sendSize:    instruments.requestSize,
		}
		streamingHandler := &streamingHandlerInterceptor{
			StreamingHandlerConn: conn,
			receive: func(msg any, conn connect.StreamingHandlerConn) error {
				return state.receive(ctx, msg, conn)
			},
			send: func(msg any, conn connect.StreamingHandlerConn) error {
				return state.send(ctx, msg, conn)
			},
		}
		err := next(ctx, streamingHandler)
		state.attributes = append(
			state.attributes,
			statusCodeAttribute(conn.Peer().Protocol, err))
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.SetAttributes(state.attributes...)
		instruments.requestsPerRPC.Record(
			ctx,
			state.receivedCounter,
			metric.WithAttributes(state.attributes...))
		instruments.responsesPerRPC.Record(
			ctx,
			state.sentCounter,
			metric.WithAttributes(state.attributes...))
		instruments.duration.Record(ctx,
			time.Since(requestStartTime).Milliseconds(),
			metric.WithAttributes(state.attributes...))
		return err
	}
}

func propagateOtelHeaders(ctx context.Context, isClient bool, header http.Header) context.Context {
	if isClient {
		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(header))
	} else {
		ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(header))
	}
	return ctx
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

func statusCodeAttribute(protocol string, err error) attribute.KeyValue {
	statusCodeKey := fmt.Sprintf("rpc.%s.status_code", protocol)
	statusCode := attribute.Int64(statusCodeKey, 0)
	if err != nil {
		statusCode = attribute.Int64(statusCodeKey, int64(connect.CodeOf(err)))
	}
	return statusCode
}

// streamingState stores the ongoing metrics for streaming interceptors.
type streamingState struct {
	mu              sync.Mutex
	spec            connect.Spec
	protocol        string
	attributes      []attribute.KeyValue
	error           error
	sentCounter     int64
	receivedCounter int64
	receiveSize     metric.Int64Histogram
	sendSize        metric.Int64Histogram
}

// streamingSenderReceiver encapsulates either a StreamingClientConn or a StreamingHandlerConn.
type streamingSenderReceiver interface {
	Receive(msg any) error
	Send(msg any) error
}

func (s *streamingState) receive(ctx context.Context, msg any, conn streamingSenderReceiver) error {
	err := conn.Receive(msg)
	s.mu.Lock()
	defer s.mu.Unlock()
	if errors.Is(err, io.EOF) {
		return err // nolint:wrapcheck
	}
	s.receivedCounter++
	if err != nil {
		s.error = err
		s.attributes = append(s.attributes, statusCodeAttribute(s.protocol, err))
	}
	attrs := append(s.attributes, []attribute.KeyValue{ // nolint:gocritic
		semconv.MessageTypeReceived,
		semconv.MessageIDKey.Int64(s.receivedCounter),
	}...)
	if protomsg, ok := msg.(proto.Message); ok {
		size := proto.Size(protomsg)
		attrs = append(attrs, semconv.MessageUncompressedSizeKey.Int(size))
		s.receiveSize.Record(ctx, int64(size), metric.WithAttributes(attrs...))
	}

	span := trace.SpanFromContext(ctx)
	span.AddEvent(otelFtlEventName, trace.WithAttributes(attrs...))
	return err // nolint:wrapcheck
}

func (s *streamingState) send(ctx context.Context, msg any, conn streamingSenderReceiver) error {
	err := conn.Send(msg)
	s.mu.Lock()
	defer s.mu.Unlock()
	if errors.Is(err, io.EOF) {
		return err // nolint:wrapcheck
	}
	s.sentCounter++
	if err != nil {
		s.error = err
		s.attributes = append(s.attributes, statusCodeAttribute(s.protocol, err))
	}
	attrs := append(s.attributes, []attribute.KeyValue{ // nolint:gocritic
		semconv.MessageTypeSent,
		semconv.MessageIDKey.Int64(s.sentCounter),
	}...)
	if protomsg, ok := msg.(proto.Message); ok {
		size := proto.Size(protomsg)
		attrs = append(attrs, semconv.MessageUncompressedSizeKey.Int(size))
		s.sendSize.Record(ctx, int64(size), metric.WithAttributes(attrs...))
	}

	span := trace.SpanFromContext(ctx)
	span.AddEvent(otelFtlEventName, trace.WithAttributes(attrs...))
	return err // nolint:wrapcheck
}

type streamingClientInterceptor struct {
	connect.StreamingClientConn
	receive func(msg any, conn connect.StreamingClientConn) error
	send    func(any, connect.StreamingClientConn) error
	onClose func()
}

func (s *streamingClientInterceptor) Receive(msg any) error {
	return s.receive(msg, s.StreamingClientConn)
}

func (s *streamingClientInterceptor) Send(msg any) error {
	return s.send(msg, s.StreamingClientConn)
}

func (s *streamingClientInterceptor) Close() error {
	err := s.StreamingClientConn.CloseResponse()
	s.onClose()
	return err // nolint:wrapcheck
}

type streamingHandlerInterceptor struct {
	connect.StreamingHandlerConn
	receive func(any, connect.StreamingHandlerConn) error
	send    func(any, connect.StreamingHandlerConn) error
}

func (s *streamingHandlerInterceptor) Receive(msg any) error {
	return s.receive(msg, s.StreamingHandlerConn)
}

func (s *streamingHandlerInterceptor) Send(msg any) error {
	return s.send(msg, s.StreamingHandlerConn)
}
