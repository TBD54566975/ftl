// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: xyz/block/ftl/provisioner/v1beta1/plugin.proto

package provisionerpbconnect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v1beta1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1"
	v1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect.IsAtLeastVersion0_1_0

const (
	// ProvisionerPluginServiceName is the fully-qualified name of the ProvisionerPluginService service.
	ProvisionerPluginServiceName = "xyz.block.ftl.provisioner.v1beta1.ProvisionerPluginService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// ProvisionerPluginServicePingProcedure is the fully-qualified name of the
	// ProvisionerPluginService's Ping RPC.
	ProvisionerPluginServicePingProcedure = "/xyz.block.ftl.provisioner.v1beta1.ProvisionerPluginService/Ping"
	// ProvisionerPluginServiceProvisionProcedure is the fully-qualified name of the
	// ProvisionerPluginService's Provision RPC.
	ProvisionerPluginServiceProvisionProcedure = "/xyz.block.ftl.provisioner.v1beta1.ProvisionerPluginService/Provision"
	// ProvisionerPluginServicePlanProcedure is the fully-qualified name of the
	// ProvisionerPluginService's Plan RPC.
	ProvisionerPluginServicePlanProcedure = "/xyz.block.ftl.provisioner.v1beta1.ProvisionerPluginService/Plan"
	// ProvisionerPluginServiceStatusProcedure is the fully-qualified name of the
	// ProvisionerPluginService's Status RPC.
	ProvisionerPluginServiceStatusProcedure = "/xyz.block.ftl.provisioner.v1beta1.ProvisionerPluginService/Status"
)

// ProvisionerPluginServiceClient is a client for the
// xyz.block.ftl.provisioner.v1beta1.ProvisionerPluginService service.
type ProvisionerPluginServiceClient interface {
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	Provision(context.Context, *connect.Request[v1beta1.ProvisionRequest]) (*connect.Response[v1beta1.ProvisionResponse], error)
	Plan(context.Context, *connect.Request[v1beta1.PlanRequest]) (*connect.Response[v1beta1.PlanResponse], error)
	Status(context.Context, *connect.Request[v1beta1.StatusRequest]) (*connect.Response[v1beta1.StatusResponse], error)
}

// NewProvisionerPluginServiceClient constructs a client for the
// xyz.block.ftl.provisioner.v1beta1.ProvisionerPluginService service. By default, it uses the
// Connect protocol with the binary Protobuf Codec, asks for gzipped responses, and sends
// uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC() or
// connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewProvisionerPluginServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) ProvisionerPluginServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &provisionerPluginServiceClient{
		ping: connect.NewClient[v1.PingRequest, v1.PingResponse](
			httpClient,
			baseURL+ProvisionerPluginServicePingProcedure,
			opts...,
		),
		provision: connect.NewClient[v1beta1.ProvisionRequest, v1beta1.ProvisionResponse](
			httpClient,
			baseURL+ProvisionerPluginServiceProvisionProcedure,
			opts...,
		),
		plan: connect.NewClient[v1beta1.PlanRequest, v1beta1.PlanResponse](
			httpClient,
			baseURL+ProvisionerPluginServicePlanProcedure,
			opts...,
		),
		status: connect.NewClient[v1beta1.StatusRequest, v1beta1.StatusResponse](
			httpClient,
			baseURL+ProvisionerPluginServiceStatusProcedure,
			opts...,
		),
	}
}

// provisionerPluginServiceClient implements ProvisionerPluginServiceClient.
type provisionerPluginServiceClient struct {
	ping      *connect.Client[v1.PingRequest, v1.PingResponse]
	provision *connect.Client[v1beta1.ProvisionRequest, v1beta1.ProvisionResponse]
	plan      *connect.Client[v1beta1.PlanRequest, v1beta1.PlanResponse]
	status    *connect.Client[v1beta1.StatusRequest, v1beta1.StatusResponse]
}

// Ping calls xyz.block.ftl.provisioner.v1beta1.ProvisionerPluginService.Ping.
func (c *provisionerPluginServiceClient) Ping(ctx context.Context, req *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return c.ping.CallUnary(ctx, req)
}

// Provision calls xyz.block.ftl.provisioner.v1beta1.ProvisionerPluginService.Provision.
func (c *provisionerPluginServiceClient) Provision(ctx context.Context, req *connect.Request[v1beta1.ProvisionRequest]) (*connect.Response[v1beta1.ProvisionResponse], error) {
	return c.provision.CallUnary(ctx, req)
}

// Plan calls xyz.block.ftl.provisioner.v1beta1.ProvisionerPluginService.Plan.
func (c *provisionerPluginServiceClient) Plan(ctx context.Context, req *connect.Request[v1beta1.PlanRequest]) (*connect.Response[v1beta1.PlanResponse], error) {
	return c.plan.CallUnary(ctx, req)
}

// Status calls xyz.block.ftl.provisioner.v1beta1.ProvisionerPluginService.Status.
func (c *provisionerPluginServiceClient) Status(ctx context.Context, req *connect.Request[v1beta1.StatusRequest]) (*connect.Response[v1beta1.StatusResponse], error) {
	return c.status.CallUnary(ctx, req)
}

// ProvisionerPluginServiceHandler is an implementation of the
// xyz.block.ftl.provisioner.v1beta1.ProvisionerPluginService service.
type ProvisionerPluginServiceHandler interface {
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	Provision(context.Context, *connect.Request[v1beta1.ProvisionRequest]) (*connect.Response[v1beta1.ProvisionResponse], error)
	Plan(context.Context, *connect.Request[v1beta1.PlanRequest]) (*connect.Response[v1beta1.PlanResponse], error)
	Status(context.Context, *connect.Request[v1beta1.StatusRequest]) (*connect.Response[v1beta1.StatusResponse], error)
}

// NewProvisionerPluginServiceHandler builds an HTTP handler from the service implementation. It
// returns the path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewProvisionerPluginServiceHandler(svc ProvisionerPluginServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	provisionerPluginServicePingHandler := connect.NewUnaryHandler(
		ProvisionerPluginServicePingProcedure,
		svc.Ping,
		opts...,
	)
	provisionerPluginServiceProvisionHandler := connect.NewUnaryHandler(
		ProvisionerPluginServiceProvisionProcedure,
		svc.Provision,
		opts...,
	)
	provisionerPluginServicePlanHandler := connect.NewUnaryHandler(
		ProvisionerPluginServicePlanProcedure,
		svc.Plan,
		opts...,
	)
	provisionerPluginServiceStatusHandler := connect.NewUnaryHandler(
		ProvisionerPluginServiceStatusProcedure,
		svc.Status,
		opts...,
	)
	return "/xyz.block.ftl.provisioner.v1beta1.ProvisionerPluginService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case ProvisionerPluginServicePingProcedure:
			provisionerPluginServicePingHandler.ServeHTTP(w, r)
		case ProvisionerPluginServiceProvisionProcedure:
			provisionerPluginServiceProvisionHandler.ServeHTTP(w, r)
		case ProvisionerPluginServicePlanProcedure:
			provisionerPluginServicePlanHandler.ServeHTTP(w, r)
		case ProvisionerPluginServiceStatusProcedure:
			provisionerPluginServiceStatusHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedProvisionerPluginServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedProvisionerPluginServiceHandler struct{}

func (UnimplementedProvisionerPluginServiceHandler) Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.provisioner.v1beta1.ProvisionerPluginService.Ping is not implemented"))
}

func (UnimplementedProvisionerPluginServiceHandler) Provision(context.Context, *connect.Request[v1beta1.ProvisionRequest]) (*connect.Response[v1beta1.ProvisionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.provisioner.v1beta1.ProvisionerPluginService.Provision is not implemented"))
}

func (UnimplementedProvisionerPluginServiceHandler) Plan(context.Context, *connect.Request[v1beta1.PlanRequest]) (*connect.Response[v1beta1.PlanResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.provisioner.v1beta1.ProvisionerPluginService.Plan is not implemented"))
}

func (UnimplementedProvisionerPluginServiceHandler) Status(context.Context, *connect.Request[v1beta1.StatusRequest]) (*connect.Response[v1beta1.StatusResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.provisioner.v1beta1.ProvisionerPluginService.Status is not implemented"))
}