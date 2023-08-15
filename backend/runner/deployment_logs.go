package runner

import (
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types"

	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/model"
)

var _ log.Sink = (*deploymentLogsSink)(nil)

func newDeploymentLogsSink(queue chan logEntry) *deploymentLogsSink {
	return &deploymentLogsSink{
		logQueue: queue,
	}
}

type deploymentLogsSink struct {
	logQueue chan logEntry
}

// Log implements Sink
func (d *deploymentLogsSink) Log(entry log.Entry) error {
	var request types.Option[model.IngressRequestKey]
	if reqStr, ok := entry.Attributes["request"]; ok {
		req, err := model.ParseIngressRequestKey(reqStr)
		if err == nil {
			request = types.Some(req)
		}
	}
	select {
	case d.logQueue <- logEntry{request: request, Entry: entry}:
	default:
		// Drop log entry if queue is full
		return errors.Errorf("log queue is full")
	}
	return nil
}
