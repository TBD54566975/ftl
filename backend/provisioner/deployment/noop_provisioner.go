package deployment

import (
	"context"

	"connectrpc.com/connect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner/provisionerconnect"
)

type NoopProvisioner struct{}

var _ provisionerconnect.ProvisionerPluginServiceClient = (*NoopProvisioner)(nil)

func (d *NoopProvisioner) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return &connect.Response[ftlv1.PingResponse]{}, nil
}

func (d *NoopProvisioner) Plan(context.Context, *connect.Request[provisioner.PlanRequest]) (*connect.Response[provisioner.PlanResponse], error) {
	panic("unimplemented")
}

func (d *NoopProvisioner) Provision(context.Context, *connect.Request[provisioner.ProvisionRequest]) (*connect.Response[provisioner.ProvisionResponse], error) {
	return connect.NewResponse(&provisioner.ProvisionResponse{
		Status:            provisioner.ProvisionResponse_NO_CHANGES,
		ProvisioningToken: "",
	}), nil
}

func (d *NoopProvisioner) Status(context.Context, *connect.Request[provisioner.StatusRequest]) (*connect.Response[provisioner.StatusResponse], error) {
	panic("should not be called")
}
