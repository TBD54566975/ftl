package deployment

import (
	"context"

	"connectrpc.com/connect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner/provisionerconnect"
)

// DevProvisioner is a provisioner for running FTL locally
type DevProvisioner struct{}

var _ provisionerconnect.ProvisionerPluginServiceClient = (*DevProvisioner)(nil)

func (d *DevProvisioner) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return &connect.Response[ftlv1.PingResponse]{}, nil
}

func (d *DevProvisioner) Plan(context.Context, *connect.Request[provisioner.PlanRequest]) (*connect.Response[provisioner.PlanResponse], error) {
	panic("unimplemented")
}

func (d *DevProvisioner) Provision(context.Context, *connect.Request[provisioner.ProvisionRequest]) (*connect.Response[provisioner.ProvisionResponse], error) {
	panic("unimplemented")
}

func (d *DevProvisioner) Status(context.Context, *connect.Request[provisioner.StatusRequest]) (*connect.Response[provisioner.StatusResponse], error) {
	panic("unimplemented")
}
