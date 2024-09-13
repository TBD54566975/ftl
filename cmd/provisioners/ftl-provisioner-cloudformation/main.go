package main

import (
	"context"

	"connectrpc.com/connect"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/provisioner"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/provisioner/provisionerconnect"
	"github.com/TBD54566975/ftl/common/plugin"
)

type CloudformationProvisioner struct {
}

func (c *CloudformationProvisioner) Ping(context.Context, *connect.Request[provisioner.PingRequest]) (*connect.Response[provisioner.PingResponse], error) {
	return &connect.Response[provisioner.PingResponse]{}, nil
}

var _ provisionerconnect.ProvisionerServiceHandler = (*CloudformationProvisioner)(nil)

func NewCloudformationProvisioner(ctx context.Context, config struct{}) (context.Context, *CloudformationProvisioner, error) {
	return ctx, &CloudformationProvisioner{}, nil
}

func main() {
	plugin.Start(
		context.Background(),
		"ftl-provisioner-cloudformation",
		NewCloudformationProvisioner,
		"",
		provisionerconnect.NewProvisionerServiceHandler,
	)
}
