// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: xyz/block/ftl/v1/module.proto

package ftlv1connect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
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
	// ModuleServiceName is the fully-qualified name of the ModuleService service.
	ModuleServiceName = "xyz.block.ftl.v1.ModuleService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// ModuleServicePingProcedure is the fully-qualified name of the ModuleService's Ping RPC.
	ModuleServicePingProcedure = "/xyz.block.ftl.v1.ModuleService/Ping"
	// ModuleServiceGetModuleContextProcedure is the fully-qualified name of the ModuleService's
	// GetModuleContext RPC.
	ModuleServiceGetModuleContextProcedure = "/xyz.block.ftl.v1.ModuleService/GetModuleContext"
	// ModuleServiceAcquireLeaseProcedure is the fully-qualified name of the ModuleService's
	// AcquireLease RPC.
	ModuleServiceAcquireLeaseProcedure = "/xyz.block.ftl.v1.ModuleService/AcquireLease"
	// ModuleServicePublishEventProcedure is the fully-qualified name of the ModuleService's
	// PublishEvent RPC.
	ModuleServicePublishEventProcedure = "/xyz.block.ftl.v1.ModuleService/PublishEvent"
)

// ModuleServiceClient is a client for the xyz.block.ftl.v1.ModuleService service.
type ModuleServiceClient interface {
	// Ping service for readiness.
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	// Get configuration state for the module
	GetModuleContext(context.Context, *connect.Request[v1.ModuleContextRequest]) (*connect.ServerStreamForClient[v1.ModuleContextResponse], error)
	// Acquire (and renew) a lease for a deployment.
	//
	// Returns ResourceExhausted if the lease is held.
	AcquireLease(context.Context) *connect.BidiStreamForClient[v1.AcquireLeaseRequest, v1.AcquireLeaseResponse]
	// Publish an event to a topic.
	PublishEvent(context.Context, *connect.Request[v1.PublishEventRequest]) (*connect.Response[v1.PublishEventResponse], error)
}

// NewModuleServiceClient constructs a client for the xyz.block.ftl.v1.ModuleService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewModuleServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) ModuleServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &moduleServiceClient{
		ping: connect.NewClient[v1.PingRequest, v1.PingResponse](
			httpClient,
			baseURL+ModuleServicePingProcedure,
			connect.WithIdempotency(connect.IdempotencyNoSideEffects),
			connect.WithClientOptions(opts...),
		),
		getModuleContext: connect.NewClient[v1.ModuleContextRequest, v1.ModuleContextResponse](
			httpClient,
			baseURL+ModuleServiceGetModuleContextProcedure,
			opts...,
		),
		acquireLease: connect.NewClient[v1.AcquireLeaseRequest, v1.AcquireLeaseResponse](
			httpClient,
			baseURL+ModuleServiceAcquireLeaseProcedure,
			opts...,
		),
		publishEvent: connect.NewClient[v1.PublishEventRequest, v1.PublishEventResponse](
			httpClient,
			baseURL+ModuleServicePublishEventProcedure,
			opts...,
		),
	}
}

// moduleServiceClient implements ModuleServiceClient.
type moduleServiceClient struct {
	ping             *connect.Client[v1.PingRequest, v1.PingResponse]
	getModuleContext *connect.Client[v1.ModuleContextRequest, v1.ModuleContextResponse]
	acquireLease     *connect.Client[v1.AcquireLeaseRequest, v1.AcquireLeaseResponse]
	publishEvent     *connect.Client[v1.PublishEventRequest, v1.PublishEventResponse]
}

// Ping calls xyz.block.ftl.v1.ModuleService.Ping.
func (c *moduleServiceClient) Ping(ctx context.Context, req *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return c.ping.CallUnary(ctx, req)
}

// GetModuleContext calls xyz.block.ftl.v1.ModuleService.GetModuleContext.
func (c *moduleServiceClient) GetModuleContext(ctx context.Context, req *connect.Request[v1.ModuleContextRequest]) (*connect.ServerStreamForClient[v1.ModuleContextResponse], error) {
	return c.getModuleContext.CallServerStream(ctx, req)
}

// AcquireLease calls xyz.block.ftl.v1.ModuleService.AcquireLease.
func (c *moduleServiceClient) AcquireLease(ctx context.Context) *connect.BidiStreamForClient[v1.AcquireLeaseRequest, v1.AcquireLeaseResponse] {
	return c.acquireLease.CallBidiStream(ctx)
}

// PublishEvent calls xyz.block.ftl.v1.ModuleService.PublishEvent.
func (c *moduleServiceClient) PublishEvent(ctx context.Context, req *connect.Request[v1.PublishEventRequest]) (*connect.Response[v1.PublishEventResponse], error) {
	return c.publishEvent.CallUnary(ctx, req)
}

// ModuleServiceHandler is an implementation of the xyz.block.ftl.v1.ModuleService service.
type ModuleServiceHandler interface {
	// Ping service for readiness.
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	// Get configuration state for the module
	GetModuleContext(context.Context, *connect.Request[v1.ModuleContextRequest], *connect.ServerStream[v1.ModuleContextResponse]) error
	// Acquire (and renew) a lease for a deployment.
	//
	// Returns ResourceExhausted if the lease is held.
	AcquireLease(context.Context, *connect.BidiStream[v1.AcquireLeaseRequest, v1.AcquireLeaseResponse]) error
	// Publish an event to a topic.
	PublishEvent(context.Context, *connect.Request[v1.PublishEventRequest]) (*connect.Response[v1.PublishEventResponse], error)
}

// NewModuleServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewModuleServiceHandler(svc ModuleServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	moduleServicePingHandler := connect.NewUnaryHandler(
		ModuleServicePingProcedure,
		svc.Ping,
		connect.WithIdempotency(connect.IdempotencyNoSideEffects),
		connect.WithHandlerOptions(opts...),
	)
	moduleServiceGetModuleContextHandler := connect.NewServerStreamHandler(
		ModuleServiceGetModuleContextProcedure,
		svc.GetModuleContext,
		opts...,
	)
	moduleServiceAcquireLeaseHandler := connect.NewBidiStreamHandler(
		ModuleServiceAcquireLeaseProcedure,
		svc.AcquireLease,
		opts...,
	)
	moduleServicePublishEventHandler := connect.NewUnaryHandler(
		ModuleServicePublishEventProcedure,
		svc.PublishEvent,
		opts...,
	)
	return "/xyz.block.ftl.v1.ModuleService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case ModuleServicePingProcedure:
			moduleServicePingHandler.ServeHTTP(w, r)
		case ModuleServiceGetModuleContextProcedure:
			moduleServiceGetModuleContextHandler.ServeHTTP(w, r)
		case ModuleServiceAcquireLeaseProcedure:
			moduleServiceAcquireLeaseHandler.ServeHTTP(w, r)
		case ModuleServicePublishEventProcedure:
			moduleServicePublishEventHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedModuleServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedModuleServiceHandler struct{}

func (UnimplementedModuleServiceHandler) Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ModuleService.Ping is not implemented"))
}

func (UnimplementedModuleServiceHandler) GetModuleContext(context.Context, *connect.Request[v1.ModuleContextRequest], *connect.ServerStream[v1.ModuleContextResponse]) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ModuleService.GetModuleContext is not implemented"))
}

func (UnimplementedModuleServiceHandler) AcquireLease(context.Context, *connect.BidiStream[v1.AcquireLeaseRequest, v1.AcquireLeaseResponse]) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ModuleService.AcquireLease is not implemented"))
}

func (UnimplementedModuleServiceHandler) PublishEvent(context.Context, *connect.Request[v1.PublishEventRequest]) (*connect.Response[v1.PublishEventResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ModuleService.PublishEvent is not implemented"))
}