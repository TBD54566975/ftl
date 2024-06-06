package admin

import (
	"context"
	"errors"
	"net"
	"net/url"

	"connectrpc.com/connect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
)

type CmdClient interface {
	Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error)

	// List configuration.
	ConfigList(ctx context.Context, req *connect.Request[ftlv1.ListConfigRequest]) (*connect.Response[ftlv1.ListConfigResponse], error)

	// Get a config value.
	ConfigGet(ctx context.Context, req *connect.Request[ftlv1.GetConfigRequest]) (*connect.Response[ftlv1.GetConfigResponse], error)

	// Set a config value.
	ConfigSet(ctx context.Context, req *connect.Request[ftlv1.SetConfigRequest]) (*connect.Response[ftlv1.SetConfigResponse], error)

	// Unset a config value.
	ConfigUnset(ctx context.Context, req *connect.Request[ftlv1.UnsetConfigRequest]) (*connect.Response[ftlv1.UnsetConfigResponse], error)

	// List secrets.
	SecretsList(ctx context.Context, req *connect.Request[ftlv1.ListSecretsRequest]) (*connect.Response[ftlv1.ListSecretsResponse], error)

	// Get a secret.
	SecretGet(ctx context.Context, req *connect.Request[ftlv1.GetSecretRequest]) (*connect.Response[ftlv1.GetSecretResponse], error)

	// Set a secret.
	SecretSet(ctx context.Context, req *connect.Request[ftlv1.SetSecretRequest]) (*connect.Response[ftlv1.SetSecretResponse], error)

	// Unset a secret.
	SecretUnset(ctx context.Context, req *connect.Request[ftlv1.UnsetSecretRequest]) (*connect.Response[ftlv1.UnsetSecretResponse], error)
}

// NewCmdClient takes the service client and endpoint flag received by the cmd interface
// and returns an appropriate interface for the cmd library to use.
//
// If the controller is not present AND endpoint is local, then inject a purely-local
// implementation of the interface so that the user does not need to spin up a controller
// just to run the `ftl config/secret` commands. Otherwise, return back the gRPC client.
func NewCmdClient(ctx context.Context, adminClient ftlv1connect.AdminServiceClient, endpoint *url.URL) (CmdClient, error) {
	_, err := adminClient.Ping(ctx, connect.NewRequest(&ftlv1.PingRequest{}))
	if isConnectUnavailableError(err) && isEndpointLocal(endpoint) {
		return newLocalCmdClient(ctx), nil
	}
	return adminClient, nil
}

func isConnectUnavailableError(err error) bool {
	var connectErr *connect.Error
	if errors.As(err, &connectErr) {
		return connectErr.Code() == connect.CodeUnavailable
	}
	return false
}

func isEndpointLocal(endpoint *url.URL) bool {
	h := endpoint.Hostname()
	ips, err := net.LookupIP(h)
	if err != nil {
		panic(err.Error())
	}
	for _, netip := range ips {
		if netip.IsLoopback() {
			return true
		}
	}
	return false
}
