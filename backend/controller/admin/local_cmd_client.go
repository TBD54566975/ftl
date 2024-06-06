package admin

import (
	"context"

	"connectrpc.com/connect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/common/configuration"
)

// localCmdClient reads and writes to local projectconfig files without making any network
// calls. It allows us to interface with local ftl-project.toml files without needing to
// start a controller.
type localCmdClient struct {
	as *AdminService
}

func newLocalCmdClient(ctx context.Context) *localCmdClient {
	cm := configuration.ConfigFromContext(ctx)
	sm := configuration.SecretsFromContext(ctx)
	return &localCmdClient{NewAdminService(cm, sm)}
}

// Ping will always return a healthy response because localCmdClient is purely local.
func (l *localCmdClient) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (l *localCmdClient) ConfigList(ctx context.Context, req *connect.Request[ftlv1.ListConfigRequest]) (*connect.Response[ftlv1.ListConfigResponse], error) {
	return l.as.ConfigList(ctx, req)
}

func (l *localCmdClient) ConfigGet(ctx context.Context, req *connect.Request[ftlv1.GetConfigRequest]) (*connect.Response[ftlv1.GetConfigResponse], error) {
	return l.as.ConfigGet(ctx, req)
}

func (l *localCmdClient) ConfigSet(ctx context.Context, req *connect.Request[ftlv1.SetConfigRequest]) (*connect.Response[ftlv1.SetConfigResponse], error) {
	return l.as.ConfigSet(ctx, req)
}

func (l *localCmdClient) ConfigUnset(ctx context.Context, req *connect.Request[ftlv1.UnsetConfigRequest]) (*connect.Response[ftlv1.UnsetConfigResponse], error) {
	return l.as.ConfigUnset(ctx, req)
}

func (l *localCmdClient) SecretsList(ctx context.Context, req *connect.Request[ftlv1.ListSecretsRequest]) (*connect.Response[ftlv1.ListSecretsResponse], error) {
	return l.as.SecretsList(ctx, req)
}

func (l *localCmdClient) SecretGet(ctx context.Context, req *connect.Request[ftlv1.GetSecretRequest]) (*connect.Response[ftlv1.GetSecretResponse], error) {
	return l.as.SecretGet(ctx, req)
}

func (l *localCmdClient) SecretSet(ctx context.Context, req *connect.Request[ftlv1.SetSecretRequest]) (*connect.Response[ftlv1.SetSecretResponse], error) {
	return l.as.SecretSet(ctx, req)
}

func (l *localCmdClient) SecretUnset(ctx context.Context, req *connect.Request[ftlv1.UnsetSecretRequest]) (*connect.Response[ftlv1.UnsetSecretResponse], error) {
	return l.as.SecretUnset(ctx, req)
}
