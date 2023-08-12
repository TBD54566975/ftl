package runner

import (
	"github.com/alecthomas/errors"

	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/model"
)

var _ log.Sink = (*deploymentLogsSink)(nil)

func newDeploymentLogsSink(deploymentKey model.DeploymentKey, runnerKey model.RunnerKey, queue chan log.Entry) *deploymentLogsSink {
	return &deploymentLogsSink{
		deploymentKey: deploymentKey,
		runnerKey:     runnerKey,
		logQueue:      queue,
	}
}

type deploymentLogsSink struct {
	deploymentKey model.DeploymentKey
	runnerKey     model.RunnerKey
	logQueue      chan log.Entry
}

// Log implements Sink
func (d *deploymentLogsSink) Log(entry log.Entry) error {
	select {
	case d.logQueue <- entry:
	default:
		// Drop log entry if queue is full
		return errors.Errorf("log queue is full")
	}
	return nil
}
