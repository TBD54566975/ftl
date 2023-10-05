// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: xyz/block/ftl/v1/console/console.proto

package pbconsoleconnect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	console "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/console"
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
	// ConsoleServiceName is the fully-qualified name of the ConsoleService service.
	ConsoleServiceName = "xyz.block.ftl.v1.console.ConsoleService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// ConsoleServicePingProcedure is the fully-qualified name of the ConsoleService's Ping RPC.
	ConsoleServicePingProcedure = "/xyz.block.ftl.v1.console.ConsoleService/Ping"
	// ConsoleServiceGetModulesProcedure is the fully-qualified name of the ConsoleService's GetModules
	// RPC.
	ConsoleServiceGetModulesProcedure = "/xyz.block.ftl.v1.console.ConsoleService/GetModules"
	// ConsoleServiceStreamEventsProcedure is the fully-qualified name of the ConsoleService's
	// StreamEvents RPC.
	ConsoleServiceStreamEventsProcedure = "/xyz.block.ftl.v1.console.ConsoleService/StreamEvents"
	// ConsoleServiceGetEventsProcedure is the fully-qualified name of the ConsoleService's GetEvents
	// RPC.
	ConsoleServiceGetEventsProcedure = "/xyz.block.ftl.v1.console.ConsoleService/GetEvents"
)

// ConsoleServiceClient is a client for the xyz.block.ftl.v1.console.ConsoleService service.
type ConsoleServiceClient interface {
	// Ping service for readiness.
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	GetModules(context.Context, *connect.Request[console.GetModulesRequest]) (*connect.Response[console.GetModulesResponse], error)
	StreamEvents(context.Context, *connect.Request[console.StreamEventsRequest]) (*connect.ServerStreamForClient[console.StreamEventsResponse], error)
	GetEvents(context.Context, *connect.Request[console.EventsQuery]) (*connect.Response[console.GetEventsResponse], error)
}

// NewConsoleServiceClient constructs a client for the xyz.block.ftl.v1.console.ConsoleService
// service. By default, it uses the Connect protocol with the binary Protobuf Codec, asks for
// gzipped responses, and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply
// the connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewConsoleServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) ConsoleServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &consoleServiceClient{
		ping: connect.NewClient[v1.PingRequest, v1.PingResponse](
			httpClient,
			baseURL+ConsoleServicePingProcedure,
			connect.WithIdempotency(connect.IdempotencyNoSideEffects),
			connect.WithClientOptions(opts...),
		),
		getModules: connect.NewClient[console.GetModulesRequest, console.GetModulesResponse](
			httpClient,
			baseURL+ConsoleServiceGetModulesProcedure,
			opts...,
		),
		streamEvents: connect.NewClient[console.StreamEventsRequest, console.StreamEventsResponse](
			httpClient,
			baseURL+ConsoleServiceStreamEventsProcedure,
			opts...,
		),
		getEvents: connect.NewClient[console.EventsQuery, console.GetEventsResponse](
			httpClient,
			baseURL+ConsoleServiceGetEventsProcedure,
			opts...,
		),
	}
}

// consoleServiceClient implements ConsoleServiceClient.
type consoleServiceClient struct {
	ping         *connect.Client[v1.PingRequest, v1.PingResponse]
	getModules   *connect.Client[console.GetModulesRequest, console.GetModulesResponse]
	streamEvents *connect.Client[console.StreamEventsRequest, console.StreamEventsResponse]
	getEvents    *connect.Client[console.EventsQuery, console.GetEventsResponse]
}

// Ping calls xyz.block.ftl.v1.console.ConsoleService.Ping.
func (c *consoleServiceClient) Ping(ctx context.Context, req *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return c.ping.CallUnary(ctx, req)
}

// GetModules calls xyz.block.ftl.v1.console.ConsoleService.GetModules.
func (c *consoleServiceClient) GetModules(ctx context.Context, req *connect.Request[console.GetModulesRequest]) (*connect.Response[console.GetModulesResponse], error) {
	return c.getModules.CallUnary(ctx, req)
}

// StreamEvents calls xyz.block.ftl.v1.console.ConsoleService.StreamEvents.
func (c *consoleServiceClient) StreamEvents(ctx context.Context, req *connect.Request[console.StreamEventsRequest]) (*connect.ServerStreamForClient[console.StreamEventsResponse], error) {
	return c.streamEvents.CallServerStream(ctx, req)
}

// GetEvents calls xyz.block.ftl.v1.console.ConsoleService.GetEvents.
func (c *consoleServiceClient) GetEvents(ctx context.Context, req *connect.Request[console.EventsQuery]) (*connect.Response[console.GetEventsResponse], error) {
	return c.getEvents.CallUnary(ctx, req)
}

// ConsoleServiceHandler is an implementation of the xyz.block.ftl.v1.console.ConsoleService
// service.
type ConsoleServiceHandler interface {
	// Ping service for readiness.
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	GetModules(context.Context, *connect.Request[console.GetModulesRequest]) (*connect.Response[console.GetModulesResponse], error)
	StreamEvents(context.Context, *connect.Request[console.StreamEventsRequest], *connect.ServerStream[console.StreamEventsResponse]) error
	GetEvents(context.Context, *connect.Request[console.EventsQuery]) (*connect.Response[console.GetEventsResponse], error)
}

// NewConsoleServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewConsoleServiceHandler(svc ConsoleServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	consoleServicePingHandler := connect.NewUnaryHandler(
		ConsoleServicePingProcedure,
		svc.Ping,
		connect.WithIdempotency(connect.IdempotencyNoSideEffects),
		connect.WithHandlerOptions(opts...),
	)
	consoleServiceGetModulesHandler := connect.NewUnaryHandler(
		ConsoleServiceGetModulesProcedure,
		svc.GetModules,
		opts...,
	)
	consoleServiceStreamEventsHandler := connect.NewServerStreamHandler(
		ConsoleServiceStreamEventsProcedure,
		svc.StreamEvents,
		opts...,
	)
	consoleServiceGetEventsHandler := connect.NewUnaryHandler(
		ConsoleServiceGetEventsProcedure,
		svc.GetEvents,
		opts...,
	)
	return "/xyz.block.ftl.v1.console.ConsoleService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case ConsoleServicePingProcedure:
			consoleServicePingHandler.ServeHTTP(w, r)
		case ConsoleServiceGetModulesProcedure:
			consoleServiceGetModulesHandler.ServeHTTP(w, r)
		case ConsoleServiceStreamEventsProcedure:
			consoleServiceStreamEventsHandler.ServeHTTP(w, r)
		case ConsoleServiceGetEventsProcedure:
			consoleServiceGetEventsHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedConsoleServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedConsoleServiceHandler struct{}

func (UnimplementedConsoleServiceHandler) Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.console.ConsoleService.Ping is not implemented"))
}

func (UnimplementedConsoleServiceHandler) GetModules(context.Context, *connect.Request[console.GetModulesRequest]) (*connect.Response[console.GetModulesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.console.ConsoleService.GetModules is not implemented"))
}

func (UnimplementedConsoleServiceHandler) StreamEvents(context.Context, *connect.Request[console.StreamEventsRequest], *connect.ServerStream[console.StreamEventsResponse]) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.console.ConsoleService.StreamEvents is not implemented"))
}

func (UnimplementedConsoleServiceHandler) GetEvents(context.Context, *connect.Request[console.EventsQuery]) (*connect.Response[console.GetEventsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.console.ConsoleService.GetEvents is not implemented"))
}
