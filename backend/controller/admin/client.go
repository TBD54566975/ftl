package admin

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"

	"connectrpc.com/connect"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
)

// Client standardizes an common interface between the AdminService as accessed via gRPC
// and a purely-local variant that doesn't require a running controller to access.
type Client interface {
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

// ShouldUseLocalClient returns whether a local admin client should be used based on the admin service client and the endpoint.
//
// If the controller is not present AND endpoint is local, then a local client should be used
// so that the user does not need to spin up a controller just to run the `ftl config/secret` commands.
//
// If true is returned, use NewLocalClient() to create a local client after setting up config and secret managers for the context.
func ShouldUseLocalClient(ctx context.Context, adminClient ftlv1connect.AdminServiceClient, endpoint *url.URL) (bool, error) {
	isLocal, err := isEndpointLocal(endpoint)
	if err != nil {
		return false, err
	}
	_, err = adminClient.Ping(ctx, connect.NewRequest(&ftlv1.PingRequest{}))
	if isConnectUnavailableError(err) && isLocal {
		return true, nil
	}
	return false, nil
}

func isConnectUnavailableError(err error) bool {
	var connectErr *connect.Error
	if errors.As(err, &connectErr) {
		return connectErr.Code() == connect.CodeUnavailable
	}
	return false
}

func isEndpointLocal(endpoint *url.URL) (bool, error) {
	h := endpoint.Hostname()
	ips, err := net.LookupIP(h)
	if err != nil {
		return false, fmt.Errorf("failed to look up own IP: %w", err)
	}
	for _, netip := range ips {
		if netip.IsLoopback() {
			return true, nil
		}
	}
	return false, nil
}
