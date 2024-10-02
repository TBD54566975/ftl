package deployment

import (
	"context"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
)

type NoopProvisioner struct{}

func (n *NoopProvisioner) Provision(ctx context.Context, module string, desired []*provisioner.ResourceContext, existing []*provisioner.Resource) (string, error) {
	return "", nil
}

func (n *NoopProvisioner) State(ctx context.Context, token string, desired []*provisioner.Resource) (TaskState, []*provisioner.Resource, error) {
	return TaskStateDone, desired, nil
}

var _ Provisioner = (*NoopProvisioner)(nil)
