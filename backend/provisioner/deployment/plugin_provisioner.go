package deployment

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner/provisionerconnect"
	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/internal/log"
)

// PluginProvisioner delegates provisioning to an external plugin
type PluginProvisioner struct {
	cmdCtx context.Context
	client *plugin.Plugin[provisionerconnect.ProvisionerPluginServiceClient]
}

var _ Provisioner = (*PluginProvisioner)(nil)

func NewPluginProvisioner(ctx context.Context, name string) (*PluginProvisioner, error) {
	client, cmdCtx, err := plugin.Spawn(
		ctx,
		log.Debug,
		"ftl-provisioner-"+name,
		".",
		"ftl-provisioner-"+name,
		provisionerconnect.NewProvisionerPluginServiceClient,
	)
	if err != nil {
		return nil, fmt.Errorf("error spawning plugin: %w", err)
	}

	return &PluginProvisioner{
		cmdCtx: cmdCtx,
		client: client,
	}, nil
}

func (p *PluginProvisioner) Provision(ctx context.Context, module string, desired []*provisioner.ResourceContext, existing []*provisioner.Resource) (string, error) {
	resp, err := p.client.Client.Provision(ctx, connect.NewRequest(&provisioner.ProvisionRequest{
		DesiredResources:  desired,
		ExistingResources: existing,
		FtlClusterId:      "ftl",
		Module:            module,
	}))
	if err != nil {
		return "", fmt.Errorf("error calling plugin: %w", err)
	}
	if resp.Msg.Status != provisioner.ProvisionResponse_SUBMITTED {
		return resp.Msg.ProvisioningToken, nil
	}
	return "", nil
}

func (p *PluginProvisioner) State(ctx context.Context, token string, desired []*provisioner.Resource) (TaskState, []*provisioner.Resource, error) {
	resp, err := p.client.Client.Status(ctx, connect.NewRequest(&provisioner.StatusRequest{
		ProvisioningToken: token,
	}))
	if err != nil {
		return "", nil, fmt.Errorf("error getting status from plugin: %w", err)
	}
	if failed, ok := resp.Msg.Status.(*provisioner.StatusResponse_Failed); ok {
		return TaskStateFailed, nil, fmt.Errorf("provisioning failed: %s", failed.Failed.ErrorMessage)
	} else if success, ok := resp.Msg.Status.(*provisioner.StatusResponse_Success); ok {
		return TaskStateDone, success.Success.UpdatedResources, nil
	}
	return TaskStateRunning, nil, nil
}
