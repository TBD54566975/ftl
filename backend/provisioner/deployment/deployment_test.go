package deployment_test

import (
	"context"
	"testing"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
	"github.com/TBD54566975/ftl/backend/provisioner/deployment"
	"github.com/alecthomas/assert/v2"
)

// MockProvisioner is a mock implementation of the Provisioner interface
type MockProvisioner struct {
	Token      string
	stateCalls int
}

var _ deployment.Provisioner = (*MockProvisioner)(nil)

func (m *MockProvisioner) Provision(
	ctx context.Context,
	module string,
	desired []*provisioner.ResourceContext,
	existing []*provisioner.Resource,
) (string, error) {
	return m.Token, nil
}

func (m *MockProvisioner) State(
	ctx context.Context,
	token string,
	desired []*provisioner.Resource,
) (deployment.TaskState, []*provisioner.Resource, error) {
	m.stateCalls++
	if m.stateCalls <= 1 {
		return deployment.TaskStateRunning, nil, nil
	}
	return deployment.TaskStateDone, desired, nil
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
