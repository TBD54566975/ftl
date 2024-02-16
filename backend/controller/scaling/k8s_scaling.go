package scaling

import (
	"context"

	"github.com/TBD54566975/ftl/internal/model"
)

var _ RunnerScaling = (*K8sScaling)(nil)

type K8sScaling struct {
}

func NewK8sScaling() *K8sScaling {
	return &K8sScaling{}
}

func (k *K8sScaling) SetReplicas(ctx context.Context, replicas int, idleRunners []model.RunnerKey) error {
	return nil
}
