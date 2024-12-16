package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/backend/timeline"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/model"
)

var _ log.Sink = (*deploymentLogsSink)(nil)

func newDeploymentLogsSink(ctx context.Context, timelineClient *timeline.Client) *deploymentLogsSink {
	sink := &deploymentLogsSink{
		logQueue: make(chan log.Entry, 10000),
	}

	// Process logs in background
	go sink.processLogs(ctx, timelineClient)

	return sink
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
		return fmt.Errorf("log queue is full")
	}
	return nil
}

func (d *deploymentLogsSink) processLogs(ctx context.Context, timelineClient *timeline.Client) {
	for {
		select {
		case entry := <-d.logQueue:
			var deployment model.DeploymentKey
			depStr, ok := entry.Attributes["deployment"]
			if !ok {
				continue
			}

			dep, err := model.ParseDeploymentKey(depStr)
			if err != nil {
				continue
			}
			deployment = dep

			var request optional.Option[model.RequestKey]
			if reqStr, ok := entry.Attributes["request"]; ok {
				req, err := model.ParseRequestKey(reqStr)
				if err == nil {
					request = optional.Some(req)
				}
			}

			var errorStr optional.Option[string]
			if entry.Error != nil {
				errorStr = optional.Some(entry.Error.Error())
			}

			timelineClient.Publish(ctx, &timeline.Log{
				RequestKey:    request,
				DeploymentKey: deployment,
				Time:          entry.Time,
				Level:         int32(entry.Level.Severity()),
				Attributes:    entry.Attributes,
				Message:       entry.Message,
				Error:         errorStr,
			})
		case <-ctx.Done():
			return
		case <-time.After(1 * time.Second):
		}
	}
}
