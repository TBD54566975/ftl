package provisioner_test

import (
	"context"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	proto "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1/provisionerconnect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/provisioner"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/alecthomas/assert/v2"
	"github.com/google/uuid"
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

func (m *MockProvisioner) Plan(context.Context, *connect.Request[proto.PlanRequest]) (*connect.Response[proto.PlanResponse], error) {
	panic("unimplemented")
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

// Status implements provisionerconnect.ProvisionerPluginServiceClient.
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
		mock := &MockProvisioner{}

		registry := provisioner.ProvisionerRegistry{}
		registry.Register("mock", mock, provisioner.ResourceTypePostgres)
		registry.Register("mock", mock, provisioner.ResourceTypeMysql)

		graph := &provisioner.ResourceGraph{}
		graph.AddNode(&proto.Resource{ResourceId: "a", Resource: &proto.Resource_Mysql{}})
		graph.AddNode(&proto.Resource{ResourceId: "b", Resource: &proto.Resource_Postgres{}})

		dpl := registry.CreateDeployment(ctx, "test-module", graph, &provisioner.ResourceGraph{})

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

	t.Run("uses output of previous task in a follow up task", func(t *testing.T) {
		dbMock := &MockProvisioner{
			StatusFn: func(ctx context.Context, req *proto.StatusRequest) (*proto.StatusResponse, error) {
				if psql, ok := req.DesiredResources[0].Resource.(*proto.Resource_Postgres); ok {
					if psql.Postgres == nil {
						psql.Postgres = &proto.PostgresResource{}
					}
					if psql.Postgres.Output == nil {
						psql.Postgres.Output = &proto.PostgresResource_PostgresResourceOutput{}
					}
					psql.Postgres.Output.ReadDsn = "postgres://localhost:5432/foo"
				} else {
					return nil, fmt.Errorf("expected postgres resource, got %T", req.DesiredResources[0].Resource)
				}

				return &proto.StatusResponse{
					Status: &proto.StatusResponse_Success{
						Success: &proto.StatusResponse_ProvisioningSuccess{
							UpdatedResources: req.DesiredResources,
						},
					},
				}, nil
			},
		}

		moduleMock := &MockProvisioner{
			ProvisionFn: func(ctx context.Context, req *proto.ProvisionRequest) (*proto.ProvisionResponse, error) {
				for _, res := range req.DesiredResources {
					for _, dep := range res.Dependencies {
						if psql, ok := dep.Resource.(*proto.Resource_Postgres); ok && psql.Postgres != nil {
							if psql.Postgres.Output == nil || psql.Postgres.Output.ReadDsn == "" {
								return nil, fmt.Errorf("read dsn is empty")
							}
						}
					}
				}
				return &proto.ProvisionResponse{
					ProvisioningToken: uuid.New().String(),
				}, nil
			},
		}

		registry := provisioner.ProvisionerRegistry{}
		registry.Register("mockdb", dbMock, provisioner.ResourceTypePostgres)
		registry.Register("mockmod", moduleMock, provisioner.ResourceTypeModule)

		// Check that the deployment finishes without errors
		graph := &provisioner.ResourceGraph{}
		graph.AddNode(&proto.Resource{ResourceId: "db", Resource: &proto.Resource_Postgres{}})
		graph.AddNode(&proto.Resource{ResourceId: "mod", Resource: &proto.Resource_Module{}})

		dpl := registry.CreateDeployment(ctx, "test-module", graph, &provisioner.ResourceGraph{})

		running := true
		for running {
			r, err := dpl.Progress(ctx)
			assert.NoError(t, err)
			running = r
		}
	})
}
