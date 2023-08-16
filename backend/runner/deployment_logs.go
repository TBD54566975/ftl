package runner

import (
	"github.com/alecthomas/errors"

	"github.com/TBD54566975/ftl/backend/common/log"
)

var _ log.Sink = (*deploymentLogsSink)(nil)

func newDeploymentLogsSink(queue chan log.Entry) *deploymentLogsSink {
	return &deploymentLogsSink{
		logQueue: queue,
	}
}

type deploymentLogsSink struct {
	logQueue chan log.Entry
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
