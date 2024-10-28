package provisioner_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	proto "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner/provisionerconnect"
	"github.com/TBD54566975/ftl/backend/provisioner"
	"github.com/TBD54566975/ftl/internal/log"
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
func (m *MockProvisioner) Plan(context.Context, *connect.Request[proto.PlanRequest]) (*connect.Response[proto.PlanResponse], error) {
	panic("unimplemented")
}

// Provision implements provisionerconnect.ProvisionerPluginServiceClient.
func (m *MockProvisioner) Provision(context.Context, *connect.Request[proto.ProvisionRequest]) (*connect.Response[proto.ProvisionResponse], error) {
	return connect.NewResponse(&proto.ProvisionResponse{
		ProvisioningToken: m.Token,
	}), nil
}

// Status implements provisionerconnect.ProvisionerPluginServiceClient.
func (m *MockProvisioner) Status(ctx context.Context, req *connect.Request[proto.StatusRequest]) (*connect.Response[proto.StatusResponse], error) {
	m.stateCalls++
	if m.stateCalls <= 1 {
		return connect.NewResponse(&proto.StatusResponse{
			Status: &proto.StatusResponse_Running{},
		}), nil
	}
	return connect.NewResponse(&proto.StatusResponse{
		Status: &proto.StatusResponse_Success{
			Success: &proto.StatusResponse_ProvisioningSuccess{
				UpdatedResources: req.Msg.DesiredResources,
			},
		},
	}), nil
}

func TestDeployment_Progress(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	t.Run("no tasks", func(t *testing.T) {
		deployment := &provisioner.Deployment{}
		progress, err := deployment.Progress(ctx)
		assert.NoError(t, err)
		assert.False(t, progress)
	})

	t.Run("progresses each provisioner in order", func(t *testing.T) {
		registry := provisioner.ProvisionerRegistry{}
		registry.Register(&MockProvisioner{Token: "foo"}, provisioner.ResourceTypePostgres)
		registry.Register(&MockProvisioner{Token: "bar"}, provisioner.ResourceTypeMysql)

		graph := &provisioner.ResourceGraph{}
		graph.AddNode(&proto.Resource{ResourceId: "a", Resource: &proto.Resource_Mysql{}})
		graph.AddNode(&proto.Resource{ResourceId: "b", Resource: &proto.Resource_Postgres{}})

		dpl := registry.CreateDeployment(ctx, "test-module", graph, &provisioner.ResourceGraph{})

		assert.Equal(t, 2, len(dpl.State().Pending))

		_, err := dpl.Progress(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(dpl.State().Pending))
		assert.NotZero(t, dpl.State().Done)

		_, err = dpl.Progress(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(dpl.State().Done))

		running, err := dpl.Progress(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(dpl.State().Done))
		assert.False(t, running)
	})
}
