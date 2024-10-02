package deployment_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner/provisionerconnect"
	"github.com/TBD54566975/ftl/backend/provisioner/deployment"
	"github.com/alecthomas/assert/v2"
)

// MockProvisioner is a mock implementation of the Provisioner interface
type MockProvisioner struct {
	Token      string
	stateCalls int
}

var _ provisionerconnect.ProvisionerPluginServiceClient = (*MockProvisioner)(nil)

// Ping implements provisionerconnect.ProvisionerPluginServiceClient.
func (m *MockProvisioner) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return &connect.Response[ftlv1.PingResponse]{}, nil
}

// Plan implements provisionerconnect.ProvisionerPluginServiceClient.
func (m *MockProvisioner) Plan(context.Context, *connect.Request[provisioner.PlanRequest]) (*connect.Response[provisioner.PlanResponse], error) {
	panic("unimplemented")
}

// Provision implements provisionerconnect.ProvisionerPluginServiceClient.
func (m *MockProvisioner) Provision(context.Context, *connect.Request[provisioner.ProvisionRequest]) (*connect.Response[provisioner.ProvisionResponse], error) {
	return connect.NewResponse(&provisioner.ProvisionResponse{
		ProvisioningToken: m.Token,
	}), nil
}

// Status implements provisionerconnect.ProvisionerPluginServiceClient.
func (m *MockProvisioner) Status(ctx context.Context, req *connect.Request[provisioner.StatusRequest]) (*connect.Response[provisioner.StatusResponse], error) {
	m.stateCalls++
	if m.stateCalls <= 1 {
		return connect.NewResponse(&provisioner.StatusResponse{
			Status: &provisioner.StatusResponse_Running{},
		}), nil
	}
	return connect.NewResponse(&provisioner.StatusResponse{
		Status: &provisioner.StatusResponse_Success{
			Success: &provisioner.StatusResponse_ProvisioningSuccess{
				UpdatedResources: req.Msg.DesiredResources,
			},
		},
	}), nil
}

func TestDeployment_Progress(t *testing.T) {
	ctx := context.Background()

	t.Run("no tasks", func(t *testing.T) {
		deployment := &deployment.Deployment{}
		progress, err := deployment.Progress(ctx)
		assert.NoError(t, err)
		assert.False(t, progress)
	})

	t.Run("progresses each provisioner in order", func(t *testing.T) {
		registry := deployment.ProvisionerRegistry{}
		registry.Register(&MockProvisioner{Token: "foo"}, deployment.ResourceTypePostgres)
		registry.Register(&MockProvisioner{Token: "bar"}, deployment.ResourceTypeMysql)

		dpl := registry.CreateDeployment(
			"test-module",
			[]*provisioner.Resource{{
				ResourceId: "a",
				Resource:   &provisioner.Resource_Mysql{},
			}, {
				ResourceId: "b",
				Resource:   &provisioner.Resource_Postgres{},
			}},
			[]*provisioner.Resource{},
		)

		assert.Equal(t, 2, len(dpl.State().Pending))

		_, err := dpl.Progress(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(dpl.State().Pending))
		assert.NotZero(t, dpl.State().Running)

		_, err = dpl.Progress(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(dpl.State().Pending))
		assert.Zero(t, dpl.State().Running)
		assert.Equal(t, 1, len(dpl.State().Done))

		_, err = dpl.Progress(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(dpl.State().Pending))
		assert.NotZero(t, dpl.State().Running)
		assert.Equal(t, 1, len(dpl.State().Done))

		running, err := dpl.Progress(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(dpl.State().Done))
		assert.False(t, running)
	})
}
