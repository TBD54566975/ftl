package runner

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/model"
)

var _ log.Sink = (*deploymentLogsSink)(nil)

func newDeploymentLogsSink(ctx context.Context, deploymentKey model.DeploymentKey, runnerKey model.RunnerKey) *deploymentLogsSink {
	return &deploymentLogsSink{
		deploymentKey: deploymentKey,
		runnerKey:     runnerKey,
		context:       ctx,
	}
}

type deploymentLogsSink struct {
	deploymentKey model.DeploymentKey
	runnerKey     model.RunnerKey
	context       context.Context
}

// Log implements Sink
func (d *deploymentLogsSink) Log(entry log.Entry) error {
	fmt.Println("entry is %m", entry.Attributes)
	log.FromContext(d.context).Log(entry)
	return nil
}
