package rpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"connectrpc.com/connect"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"

	"github.com/TBD54566975/ftl/internal/log"
)

const (
	otelFtlRequestKeyAttr            = attribute.Key("ftl.request_key")
	otelFtlVerbChainAttr             = attribute.Key("ftl.verb_chain")
	otelFtlVerbRefAttr               = attribute.Key("ftl.verb.ref")
	otelFtlVerbModuleAttr            = attribute.Key("ftl.verb.module")
	otelMessageSentSizesAttr         = attribute.Key("ftl.rpc.message.sent.sizes_bytes")
	otelMessageReceivedSizesAttr     = attribute.Key("ftl.rpc.message.received.sizes_bytes")
	otelRPCDurationMetricName        = "ftl.rpc.duration_ms"
	otelRPCRequestSizeMetricName     = "ftl.rpc.request.size_bytes"
	otelRPCRequestsPerRPCMetricName  = "ftl.rpc.request.count_per_rpc"
	otelRPCResponseSizeMetricName    = "ftl.rpc.response.size_bytes"
	otelRPCResponsesPerRPCMetricName = "ftl.rpc.response.count_per_rpc"
)

func CustomOtelInterceptor() connect.Interceptor {
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

func getAttributes(ctx context.Context) []attribute.KeyValue {
	logger := log.FromContext(ctx)
	attributes := []attribute.KeyValue{}
	requestKey, err := RequestKeyFromContext(ctx)
	if err != nil {
		logger.Warnf("failed to get request key: %s", err)
	}
	if key, ok := requestKey.Get(); ok {
		attributes = append(attributes, otelFtlRequestKeyAttr.String(key.String()))
	}
	if verb, ok := VerbFromContext(ctx); ok {
		attributes = append(attributes, otelFtlVerbRefAttr.String(verb.String()))
		attributes = append(attributes, otelFtlVerbModuleAttr.String(verb.Module))
	}
	if verbs, ok := VerbsFromContext(ctx); ok {
		verbStrings := make([]string, len(verbs))
		for i, v := range verbs {
			verbStrings[i] = v.String()
		}
		attributes = append(attributes, otelFtlVerbChainAttr.StringSlice(verbStrings))
	}
	return attributes
}

func (i *otelInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		requestStartTime := time.Now()
		isClient := request.Spec().IsClient

		requestSizesAttr := otelMessageSentSizesAttr
		responseSizesAttr := otelMessageReceivedSizesAttr
		if !isClient {
			requestSizesAttr = otelMessageReceivedSizesAttr
			responseSizesAttr = otelMessageSentSizesAttr
		}

		attributes := getAttributes(ctx)
		requestSize := 0
		if request != nil {
			if msg, ok := request.Any().(proto.Message); ok {
				requestSize = proto.Size(msg)
			}
		}

		response, err := next(ctx, request)
		responseSize := 0
		if err == nil {
			if msg, ok := response.Any().(proto.Message); ok {
				responseSize = proto.Size(msg)
			}
		}

		span := trace.SpanFromContext(ctx)
		duration := time.Since(requestStartTime).Milliseconds()
		span.SetAttributes(append(attributes,
			attribute.Int64(otelRPCRequestsPerRPCMetricName, 1),
			attribute.Int64(otelRPCRequestSizeMetricName, int64(requestSize)),
			attribute.Int64(otelRPCResponsesPerRPCMetricName, 1),
			attribute.Int64(otelRPCResponseSizeMetricName, int64(responseSize)),
			attribute.Int64(otelRPCDurationMetricName, duration),
			requestSizesAttr.Int64Slice([]int64{int64(requestSize)}),
			responseSizesAttr.Int64Slice([]int64{int64(responseSize)}),
		)...)
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
		attributes := getAttributes(ctx)
		conn := next(ctx, spec)

		instruments := getInstruments(spec.IsClient)
		state := &streamingState{
			spec:              spec,
			protocol:          conn.Peer().Protocol,
			attributes:        attributes,
			receiveSizeMetric: instruments.responseSize,
			sendSizeMetric:    instruments.requestSize,
			receiveSizes:      []int64{},
			sendSizes:         []int64{},
		}

		span := trace.SpanFromContext(ctx)
		return &streamingClientInterceptor{ // nolint:spancheck
			StreamingClientConn: conn,
			receive: func(msg any, conn connect.StreamingClientConn) error {
				return state.receive(ctx, msg, conn)
			},
			send: func(msg any, conn connect.StreamingClientConn) error {
				return state.send(ctx, msg, conn)
			},
			onClose: func() {
				duration := time.Since(requestStartTime).Milliseconds()
				span.SetAttributes(append(state.attributes,
					attribute.Int64(otelRPCRequestsPerRPCMetricName, state.sentCounter),
					attribute.Int64(otelRPCResponsesPerRPCMetricName, state.receivedCounter),
					attribute.Int64(otelRPCDurationMetricName, duration),
					otelMessageSentSizesAttr.Int64Slice(state.sendSizes),
					otelMessageReceivedSizesAttr.Int64Slice(state.receiveSizes),
				)...)
				if state.error != nil {
					span.SetStatus(codes.Error, state.error.Error())
				}
				span.End()
				instruments.requestsPerRPC.Record(ctx, state.sentCounter, metric.WithAttributes(state.attributes...))
				instruments.responsesPerRPC.Record(ctx, state.receivedCounter, metric.WithAttributes(state.attributes...))
				instruments.duration.Record(ctx, duration, metric.WithAttributes(state.attributes...))
			},
		}
	}
}

func (i *otelInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		requestStartTime := time.Now()
		attributes := getAttributes(ctx)
		instruments := getInstruments(conn.Spec().IsClient)
		state := &streamingState{
			spec:              conn.Spec(),
			protocol:          conn.Peer().Protocol,
			attributes:        attributes,
			receiveSizeMetric: instruments.requestSize,
			sendSizeMetric:    instruments.responseSize,
			receiveSizes:      []int64{},
			sendSizes:         []int64{},
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
		duration := time.Since(requestStartTime).Milliseconds()
		span := trace.SpanFromContext(ctx)
		span.SetAttributes(append(state.attributes,
			attribute.Int64(otelRPCRequestsPerRPCMetricName, state.receivedCounter),
			attribute.Int64(otelRPCResponsesPerRPCMetricName, state.sentCounter),
			attribute.Int64(otelRPCDurationMetricName, duration),
			otelMessageSentSizesAttr.Int64Slice(state.sendSizes),
			otelMessageReceivedSizesAttr.Int64Slice(state.receiveSizes),
		)...)
		instruments.requestsPerRPC.Record(ctx, state.receivedCounter, metric.WithAttributes(state.attributes...))
		instruments.responsesPerRPC.Record(ctx, state.sentCounter, metric.WithAttributes(state.attributes...))
		instruments.duration.Record(ctx, duration, metric.WithAttributes(state.attributes...))
		return err
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
	duration, err := meter.Int64Histogram(otelRPCDurationMetricName, metric.WithUnit("ms"), metric.WithDescription("Duration of the RPC call"))
	if err != nil {
		panic(fmt.Errorf("failed to create duration metric: %w", err))
	}
	requestSize, err := meter.Int64Histogram(otelRPCRequestSizeMetricName, metric.WithUnit("By"), metric.WithDescription("Size of the request payload"))
	if err != nil {
		panic(fmt.Errorf("failed to create request size metric: %w", err))
	}
	responseSize, err := meter.Int64Histogram(otelRPCResponseSizeMetricName, metric.WithUnit("By"), metric.WithDescription("Size of the response payload"))
	if err != nil {
		panic(fmt.Errorf("failed to create response size metric: %w", err))
	}
	requestsPerRPC, err := meter.Int64Histogram(otelRPCRequestsPerRPCMetricName, metric.WithUnit("1"), metric.WithDescription("Number of requests made in the RPC call"))
	if err != nil {
		panic(fmt.Errorf("failed to create requests per rpc metric: %w", err))
	}
	responsesPerRPC, err := meter.Int64Histogram(otelRPCResponsesPerRPCMetricName, metric.WithUnit("1"), metric.WithDescription("Number of responses received in the RPC call"))
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
	statusCodeKey := fmt.Sprintf("ftl.rpc.%s.status_code", protocol)
	statusCode := attribute.Int64(statusCodeKey, 0)
	if err != nil {
		statusCode = attribute.Int64(statusCodeKey, int64(connect.CodeOf(err)))
	}
	return statusCode
}

// streamingState stores the ongoing metrics for streaming interceptors.
type streamingState struct {
	mu                sync.Mutex
	spec              connect.Spec
	protocol          string
	attributes        []attribute.KeyValue
	error             error
	sentCounter       int64
	receivedCounter   int64
	receiveSizeMetric metric.Int64Histogram
	sendSizeMetric    metric.Int64Histogram
	receiveSizes      []int64
	sendSizes         []int64
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
	if err != nil {
		s.error = err
		s.attributes = append(s.attributes, statusCodeAttribute(s.protocol, err))
	}
	s.receivedCounter++
	if protomsg, ok := msg.(proto.Message); ok {
		size := proto.Size(protomsg)
		s.receiveSizes = append(s.receiveSizes, int64(size))
		s.receiveSizeMetric.Record(ctx, int64(size), metric.WithAttributes(s.attributes...))
	}
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
	if protomsg, ok := msg.(proto.Message); ok {
		size := proto.Size(protomsg)
		s.sendSizes = append(s.sendSizes, int64(size))
		s.sendSizeMetric.Record(ctx, int64(size), metric.WithAttributes(s.attributes...))
	}
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
