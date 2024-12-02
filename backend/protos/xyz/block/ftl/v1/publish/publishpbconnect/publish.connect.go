// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: xyz/block/ftl/v1/publish/publish.proto

package publishpbconnect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	publish "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/publish"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect.IsAtLeastVersion1_7_0

const (
	// PublishServiceName is the fully-qualified name of the PublishService service.
	PublishServiceName = "xyz.block.ftl.v1.publish.PublishService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// PublishServicePingProcedure is the fully-qualified name of the PublishService's Ping RPC.
	PublishServicePingProcedure = "/xyz.block.ftl.v1.publish.PublishService/Ping"
	// PublishServicePublishEventProcedure is the fully-qualified name of the PublishService's
	// PublishEvent RPC.
	PublishServicePublishEventProcedure = "/xyz.block.ftl.v1.publish.PublishService/PublishEvent"
)

// PublishServiceClient is a client for the xyz.block.ftl.v1.publish.PublishService service.
type PublishServiceClient interface {
	// Ping service for readiness.
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	// Publish a message to a topic.
	PublishEvent(context.Context, *connect.Request[publish.PublishEventRequest]) (*connect.Response[publish.PublishEventResponse], error)
}

// NewPublishServiceClient constructs a client for the xyz.block.ftl.v1.publish.PublishService
// service. By default, it uses the Connect protocol with the binary Protobuf Codec, asks for
// gzipped responses, and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply
// the connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewPublishServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) PublishServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &publishServiceClient{
		ping: connect.NewClient[v1.PingRequest, v1.PingResponse](
			httpClient,
			baseURL+PublishServicePingProcedure,
			connect.WithIdempotency(connect.IdempotencyNoSideEffects),
			connect.WithClientOptions(opts...),
		),
		publishEvent: connect.NewClient[publish.PublishEventRequest, publish.PublishEventResponse](
			httpClient,
			baseURL+PublishServicePublishEventProcedure,
			opts...,
		),
	}
}

// publishServiceClient implements PublishServiceClient.
type publishServiceClient struct {
	ping         *connect.Client[v1.PingRequest, v1.PingResponse]
	publishEvent *connect.Client[publish.PublishEventRequest, publish.PublishEventResponse]
}

// Ping calls xyz.block.ftl.v1.publish.PublishService.Ping.
func (c *publishServiceClient) Ping(ctx context.Context, req *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return c.ping.CallUnary(ctx, req)
}

// PublishEvent calls xyz.block.ftl.v1.publish.PublishService.PublishEvent.
func (c *publishServiceClient) PublishEvent(ctx context.Context, req *connect.Request[publish.PublishEventRequest]) (*connect.Response[publish.PublishEventResponse], error) {
	return c.publishEvent.CallUnary(ctx, req)
}

// PublishServiceHandler is an implementation of the xyz.block.ftl.v1.publish.PublishService
// service.
type PublishServiceHandler interface {
	// Ping service for readiness.
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	// Publish a message to a topic.
	PublishEvent(context.Context, *connect.Request[publish.PublishEventRequest]) (*connect.Response[publish.PublishEventResponse], error)
}

// NewPublishServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewPublishServiceHandler(svc PublishServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	publishServicePingHandler := connect.NewUnaryHandler(
		PublishServicePingProcedure,
		svc.Ping,
		connect.WithIdempotency(connect.IdempotencyNoSideEffects),
		connect.WithHandlerOptions(opts...),
	)
	publishServicePublishEventHandler := connect.NewUnaryHandler(
		PublishServicePublishEventProcedure,
		svc.PublishEvent,
		opts...,
	)
	return "/xyz.block.ftl.v1.publish.PublishService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case PublishServicePingProcedure:
			publishServicePingHandler.ServeHTTP(w, r)
		case PublishServicePublishEventProcedure:
			publishServicePublishEventHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedPublishServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedPublishServiceHandler struct{}

func (UnimplementedPublishServiceHandler) Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.publish.PublishService.Ping is not implemented"))
}

func (UnimplementedPublishServiceHandler) PublishEvent(context.Context, *connect.Request[publish.PublishEventRequest]) (*connect.Response[publish.PublishEventResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.publish.PublishService.PublishEvent is not implemented"))
}
