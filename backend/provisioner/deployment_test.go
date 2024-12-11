package provisioner_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"
	"github.com/google/uuid"

	proto "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1"
	provisionerconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1/provisionerpbconnect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/provisioner"
	"github.com/TBD54566975/ftl/common/schema"
	"github.com/TBD54566975/ftl/internal/log"
)

// MockProvisioner is a mock implementation of the Provisioner interface
type MockProvisioner struct {
	StatusFn    func(ctx context.Context, req *proto.StatusRequest) (*proto.StatusResponse, error)
	ProvisionFn func(ctx context.Context, req *proto.ProvisionRequest) (*proto.ProvisionResponse, error)

	stateCalls int
}

var _ provisionerconnect.ProvisionerPluginServiceClient = (*MockProvisioner)(nil)

func (m *MockProvisioner) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return &connect.Response[ftlv1.PingResponse]{}, nil
}

func (m *MockProvisioner) Provision(ctx context.Context, req *connect.Request[proto.ProvisionRequest]) (*connect.Response[proto.ProvisionResponse], error) {
	if m.ProvisionFn != nil {
		resp, err := m.ProvisionFn(ctx, req.Msg)
		if err != nil {
			return nil, err
		}
		return connect.NewResponse(resp), nil
	}

	return connect.NewResponse(&proto.ProvisionResponse{
		ProvisioningToken: uuid.New().String(),
	}), nil
}

func (m *MockProvisioner) Status(ctx context.Context, req *connect.Request[proto.StatusRequest]) (*connect.Response[proto.StatusResponse], error) {
	m.stateCalls++
	if m.stateCalls <= 1 {
		return connect.NewResponse(&proto.StatusResponse{
			Status: &proto.StatusResponse_Running{},
		}), nil
	}

	if m.StatusFn != nil {
		rep, err := m.StatusFn(ctx, req.Msg)
		if err != nil {
			return nil, err
		}
		return connect.NewResponse(rep), nil
	}

	return connect.NewResponse(&proto.StatusResponse{
		Status: &proto.StatusResponse_Success{
			Success: &proto.StatusResponse_ProvisioningSuccess{
				Events: []*proto.ProvisioningEvent{},
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
		mock := &MockProvisioner{}

		registry := provisioner.ProvisionerRegistry{}
		registry.Register("mock", mock, schema.ResourceTypePostgres)
		registry.Register("mock", mock, schema.ResourceTypeMysql)

		dpl := registry.CreateDeployment(ctx, &schema.Module{
			Name: "test-module",
			Decls: []schema.Decl{
				&schema.Database{Name: "a", Type: "mysql"},
				&schema.Database{Name: "b", Type: "postgres"},
			},
		}, nil)

		assert.Equal(t, 2, len(dpl.State().Pending))

		_, err := dpl.Progress(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(dpl.State().Pending))
		assert.NotEqual(t, 0, len(dpl.State().Done))

		_, err = dpl.Progress(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(dpl.State().Done))

		running, err := dpl.Progress(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(dpl.State().Done))
		assert.False(t, running)
	})
}
