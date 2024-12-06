// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: xyz/block/ftl/console/v1/console.proto

package pbconsoleconnect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v11 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/console/v1"
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
	// ConsoleServiceName is the fully-qualified name of the ConsoleService service.
	ConsoleServiceName = "xyz.block.ftl.console.v1.ConsoleService"
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
	ConsoleServicePingProcedure = "/xyz.block.ftl.console.v1.ConsoleService/Ping"
	// ConsoleServiceGetModulesProcedure is the fully-qualified name of the ConsoleService's GetModules
	// RPC.
	ConsoleServiceGetModulesProcedure = "/xyz.block.ftl.console.v1.ConsoleService/GetModules"
	// ConsoleServiceStreamModulesProcedure is the fully-qualified name of the ConsoleService's
	// StreamModules RPC.
	ConsoleServiceStreamModulesProcedure = "/xyz.block.ftl.console.v1.ConsoleService/StreamModules"
	// ConsoleServiceGetConfigProcedure is the fully-qualified name of the ConsoleService's GetConfig
	// RPC.
	ConsoleServiceGetConfigProcedure = "/xyz.block.ftl.console.v1.ConsoleService/GetConfig"
	// ConsoleServiceSetConfigProcedure is the fully-qualified name of the ConsoleService's SetConfig
	// RPC.
	ConsoleServiceSetConfigProcedure = "/xyz.block.ftl.console.v1.ConsoleService/SetConfig"
	// ConsoleServiceGetSecretProcedure is the fully-qualified name of the ConsoleService's GetSecret
	// RPC.
	ConsoleServiceGetSecretProcedure = "/xyz.block.ftl.console.v1.ConsoleService/GetSecret"
	// ConsoleServiceSetSecretProcedure is the fully-qualified name of the ConsoleService's SetSecret
	// RPC.
	ConsoleServiceSetSecretProcedure = "/xyz.block.ftl.console.v1.ConsoleService/SetSecret"
)

// ConsoleServiceClient is a client for the xyz.block.ftl.console.v1.ConsoleService service.
type ConsoleServiceClient interface {
	// Ping service for readiness.
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	GetModules(context.Context, *connect.Request[v11.GetModulesRequest]) (*connect.Response[v11.GetModulesResponse], error)
	StreamModules(context.Context, *connect.Request[v11.StreamModulesRequest]) (*connect.ServerStreamForClient[v11.StreamModulesResponse], error)
	GetConfig(context.Context, *connect.Request[v11.GetConfigRequest]) (*connect.Response[v11.GetConfigResponse], error)
	SetConfig(context.Context, *connect.Request[v11.SetConfigRequest]) (*connect.Response[v11.SetConfigResponse], error)
	GetSecret(context.Context, *connect.Request[v11.GetSecretRequest]) (*connect.Response[v11.GetSecretResponse], error)
	SetSecret(context.Context, *connect.Request[v11.SetSecretRequest]) (*connect.Response[v11.SetSecretResponse], error)
}

// NewConsoleServiceClient constructs a client for the xyz.block.ftl.console.v1.ConsoleService
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
		getModules: connect.NewClient[v11.GetModulesRequest, v11.GetModulesResponse](
			httpClient,
			baseURL+ConsoleServiceGetModulesProcedure,
			opts...,
		),
		streamModules: connect.NewClient[v11.StreamModulesRequest, v11.StreamModulesResponse](
			httpClient,
			baseURL+ConsoleServiceStreamModulesProcedure,
			opts...,
		),
		getConfig: connect.NewClient[v11.GetConfigRequest, v11.GetConfigResponse](
			httpClient,
			baseURL+ConsoleServiceGetConfigProcedure,
			opts...,
		),
		setConfig: connect.NewClient[v11.SetConfigRequest, v11.SetConfigResponse](
			httpClient,
			baseURL+ConsoleServiceSetConfigProcedure,
			opts...,
		),
		getSecret: connect.NewClient[v11.GetSecretRequest, v11.GetSecretResponse](
			httpClient,
			baseURL+ConsoleServiceGetSecretProcedure,
			opts...,
		),
		setSecret: connect.NewClient[v11.SetSecretRequest, v11.SetSecretResponse](
			httpClient,
			baseURL+ConsoleServiceSetSecretProcedure,
			opts...,
		),
	}
}

// consoleServiceClient implements ConsoleServiceClient.
type consoleServiceClient struct {
	ping          *connect.Client[v1.PingRequest, v1.PingResponse]
	getModules    *connect.Client[v11.GetModulesRequest, v11.GetModulesResponse]
	streamModules *connect.Client[v11.StreamModulesRequest, v11.StreamModulesResponse]
	getConfig     *connect.Client[v11.GetConfigRequest, v11.GetConfigResponse]
	setConfig     *connect.Client[v11.SetConfigRequest, v11.SetConfigResponse]
	getSecret     *connect.Client[v11.GetSecretRequest, v11.GetSecretResponse]
	setSecret     *connect.Client[v11.SetSecretRequest, v11.SetSecretResponse]
}

// Ping calls xyz.block.ftl.console.v1.ConsoleService.Ping.
func (c *consoleServiceClient) Ping(ctx context.Context, req *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return c.ping.CallUnary(ctx, req)
}

// GetModules calls xyz.block.ftl.console.v1.ConsoleService.GetModules.
func (c *consoleServiceClient) GetModules(ctx context.Context, req *connect.Request[v11.GetModulesRequest]) (*connect.Response[v11.GetModulesResponse], error) {
	return c.getModules.CallUnary(ctx, req)
}

// StreamModules calls xyz.block.ftl.console.v1.ConsoleService.StreamModules.
func (c *consoleServiceClient) StreamModules(ctx context.Context, req *connect.Request[v11.StreamModulesRequest]) (*connect.ServerStreamForClient[v11.StreamModulesResponse], error) {
	return c.streamModules.CallServerStream(ctx, req)
}

// GetConfig calls xyz.block.ftl.console.v1.ConsoleService.GetConfig.
func (c *consoleServiceClient) GetConfig(ctx context.Context, req *connect.Request[v11.GetConfigRequest]) (*connect.Response[v11.GetConfigResponse], error) {
	return c.getConfig.CallUnary(ctx, req)
}

// SetConfig calls xyz.block.ftl.console.v1.ConsoleService.SetConfig.
func (c *consoleServiceClient) SetConfig(ctx context.Context, req *connect.Request[v11.SetConfigRequest]) (*connect.Response[v11.SetConfigResponse], error) {
	return c.setConfig.CallUnary(ctx, req)
}

// GetSecret calls xyz.block.ftl.console.v1.ConsoleService.GetSecret.
func (c *consoleServiceClient) GetSecret(ctx context.Context, req *connect.Request[v11.GetSecretRequest]) (*connect.Response[v11.GetSecretResponse], error) {
	return c.getSecret.CallUnary(ctx, req)
}

// SetSecret calls xyz.block.ftl.console.v1.ConsoleService.SetSecret.
func (c *consoleServiceClient) SetSecret(ctx context.Context, req *connect.Request[v11.SetSecretRequest]) (*connect.Response[v11.SetSecretResponse], error) {
	return c.setSecret.CallUnary(ctx, req)
}

// ConsoleServiceHandler is an implementation of the xyz.block.ftl.console.v1.ConsoleService
// service.
type ConsoleServiceHandler interface {
	// Ping service for readiness.
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	GetModules(context.Context, *connect.Request[v11.GetModulesRequest]) (*connect.Response[v11.GetModulesResponse], error)
	StreamModules(context.Context, *connect.Request[v11.StreamModulesRequest], *connect.ServerStream[v11.StreamModulesResponse]) error
	GetConfig(context.Context, *connect.Request[v11.GetConfigRequest]) (*connect.Response[v11.GetConfigResponse], error)
	SetConfig(context.Context, *connect.Request[v11.SetConfigRequest]) (*connect.Response[v11.SetConfigResponse], error)
	GetSecret(context.Context, *connect.Request[v11.GetSecretRequest]) (*connect.Response[v11.GetSecretResponse], error)
	SetSecret(context.Context, *connect.Request[v11.SetSecretRequest]) (*connect.Response[v11.SetSecretResponse], error)
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
	consoleServiceStreamModulesHandler := connect.NewServerStreamHandler(
		ConsoleServiceStreamModulesProcedure,
		svc.StreamModules,
		opts...,
	)
	consoleServiceGetConfigHandler := connect.NewUnaryHandler(
		ConsoleServiceGetConfigProcedure,
		svc.GetConfig,
		opts...,
	)
	consoleServiceSetConfigHandler := connect.NewUnaryHandler(
		ConsoleServiceSetConfigProcedure,
		svc.SetConfig,
		opts...,
	)
	consoleServiceGetSecretHandler := connect.NewUnaryHandler(
		ConsoleServiceGetSecretProcedure,
		svc.GetSecret,
		opts...,
	)
	consoleServiceSetSecretHandler := connect.NewUnaryHandler(
		ConsoleServiceSetSecretProcedure,
		svc.SetSecret,
		opts...,
	)
	return "/xyz.block.ftl.console.v1.ConsoleService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case ConsoleServicePingProcedure:
			consoleServicePingHandler.ServeHTTP(w, r)
		case ConsoleServiceGetModulesProcedure:
			consoleServiceGetModulesHandler.ServeHTTP(w, r)
		case ConsoleServiceStreamModulesProcedure:
			consoleServiceStreamModulesHandler.ServeHTTP(w, r)
		case ConsoleServiceGetConfigProcedure:
			consoleServiceGetConfigHandler.ServeHTTP(w, r)
		case ConsoleServiceSetConfigProcedure:
			consoleServiceSetConfigHandler.ServeHTTP(w, r)
		case ConsoleServiceGetSecretProcedure:
			consoleServiceGetSecretHandler.ServeHTTP(w, r)
		case ConsoleServiceSetSecretProcedure:
			consoleServiceSetSecretHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedConsoleServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedConsoleServiceHandler struct{}

func (UnimplementedConsoleServiceHandler) Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.console.v1.ConsoleService.Ping is not implemented"))
}

func (UnimplementedConsoleServiceHandler) GetModules(context.Context, *connect.Request[v11.GetModulesRequest]) (*connect.Response[v11.GetModulesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.console.v1.ConsoleService.GetModules is not implemented"))
}

func (UnimplementedConsoleServiceHandler) StreamModules(context.Context, *connect.Request[v11.StreamModulesRequest], *connect.ServerStream[v11.StreamModulesResponse]) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.console.v1.ConsoleService.StreamModules is not implemented"))
}

func (UnimplementedConsoleServiceHandler) GetConfig(context.Context, *connect.Request[v11.GetConfigRequest]) (*connect.Response[v11.GetConfigResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.console.v1.ConsoleService.GetConfig is not implemented"))
}

func (UnimplementedConsoleServiceHandler) SetConfig(context.Context, *connect.Request[v11.SetConfigRequest]) (*connect.Response[v11.SetConfigResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.console.v1.ConsoleService.SetConfig is not implemented"))
}

func (UnimplementedConsoleServiceHandler) GetSecret(context.Context, *connect.Request[v11.GetSecretRequest]) (*connect.Response[v11.GetSecretResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.console.v1.ConsoleService.GetSecret is not implemented"))
}

func (UnimplementedConsoleServiceHandler) SetSecret(context.Context, *connect.Request[v11.SetSecretRequest]) (*connect.Response[v11.SetSecretResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.console.v1.ConsoleService.SetSecret is not implemented"))
}
