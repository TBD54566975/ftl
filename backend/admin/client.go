package admin

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"time"

	"connectrpc.com/connect"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
)

// Client standardizes an common interface between the AdminService as accessed via gRPC
// and a purely-local variant that doesn't require a running controller to access.
type Client interface {
	Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error)

	// List configuration.
	ConfigList(ctx context.Context, req *connect.Request[ftlv1.ConfigListRequest]) (*connect.Response[ftlv1.ConfigListResponse], error)

	// Get a config value.
	ConfigGet(ctx context.Context, req *connect.Request[ftlv1.ConfigGetRequest]) (*connect.Response[ftlv1.ConfigGetResponse], error)

	// Set a config value.
	ConfigSet(ctx context.Context, req *connect.Request[ftlv1.ConfigSetRequest]) (*connect.Response[ftlv1.ConfigSetResponse], error)

	// Unset a config value.
	ConfigUnset(ctx context.Context, req *connect.Request[ftlv1.ConfigUnsetRequest]) (*connect.Response[ftlv1.ConfigUnsetResponse], error)

	// List secrets.
	SecretsList(ctx context.Context, req *connect.Request[ftlv1.SecretsListRequest]) (*connect.Response[ftlv1.SecretsListResponse], error)

	// Get a secret.
	SecretGet(ctx context.Context, req *connect.Request[ftlv1.SecretGetRequest]) (*connect.Response[ftlv1.SecretGetResponse], error)

	// Set a secret.
	SecretSet(ctx context.Context, req *connect.Request[ftlv1.SecretSetRequest]) (*connect.Response[ftlv1.SecretSetResponse], error)

	// Unset a secret.
	SecretUnset(ctx context.Context, req *connect.Request[ftlv1.SecretUnsetRequest]) (*connect.Response[ftlv1.SecretUnsetResponse], error)

	// MapConfigsForModule combines all configuration values visible to the module.
	// Local values take precedence.
	MapConfigsForModule(ctx context.Context, req *connect.Request[ftlv1.MapConfigsForModuleRequest]) (*connect.Response[ftlv1.MapConfigsForModuleResponse], error)

	// MapSecretsForModule combines all secrets visible to the module.
	// Local values take precedence.
	MapSecretsForModule(ctx context.Context, req *connect.Request[ftlv1.MapSecretsForModuleRequest]) (*connect.Response[ftlv1.MapSecretsForModuleResponse], error)
}

// ShouldUseLocalClient returns whether a local admin client should be used based on the admin service client and the endpoint.
//
// If the service is not present AND endpoint is local, then a local client should be used
// so that the user does not need to spin up a cluster just to run the `ftl config/secret` commands.
//
// If true is returned, use NewLocalClient() to create a local client after setting up config and secret managers for the context.
func ShouldUseLocalClient(ctx context.Context, adminClient ftlv1connect.AdminServiceClient, endpoint *url.URL) (bool, error) {
	isLocal, err := isEndpointLocal(endpoint)
	if err != nil {
		return false, err
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
	defer cancel()
	_, err = adminClient.Ping(timeoutCtx, connect.NewRequest(&ftlv1.PingRequest{}))
	if isConnectUnavailableError(err) && isLocal {
		return true, nil
	}
	return false, nil
}

func isConnectUnavailableError(err error) bool {
	var connectErr *connect.Error
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	} else if errors.As(err, &connectErr) {
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
