package scaling

import (
	"context"

	"github.com/TBD54566975/ftl/backend/common/model"
)

type RunnerScaling interface {
	SetReplicas(ctx context.Context, replicas int, idleRunners []model.RunnerKey) error
}
