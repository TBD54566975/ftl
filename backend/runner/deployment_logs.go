package runner

import (
	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/model"
)

var _ log.Sink = (*deploymentLogsSink)(nil)

func newDeploymentLogsSink(deploymentKey model.DeploymentKey, runnerKey model.RunnerKey) *deploymentLogsSink {
	return &deploymentLogsSink{
		deploymentKey: deploymentKey,
		runnerKey:     runnerKey,
	}
}

type deploymentLogsSink struct {
	deploymentKey model.DeploymentKey
	runnerKey     model.RunnerKey
}

// Log implements Sink
func (d *deploymentLogsSink) Log(entry log.Entry) error {
	return nil
}
