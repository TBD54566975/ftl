package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/timeline"
	timelinedal "github.com/TBD54566975/ftl/backend/controller/timeline/dal"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
)

var _ log.Sink = (*deploymentLogsSink)(nil)

func newDeploymentLogsSink(ctx context.Context, timeline *timeline.Service) *deploymentLogsSink {
	sink := &deploymentLogsSink{
		logQueue: make(chan log.Entry, 10000),
		timeline: timeline,
	}

	// Process logs in background
	go sink.processLogs(ctx)

	return sink
}

type deploymentLogsSink struct {
	logQueue chan log.Entry
	timeline *timeline.Service
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

func (d *deploymentLogsSink) processLogs(ctx context.Context) {
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

			var errorStr optional.Option[string]
			if entry.Error != nil {
				errorStr = optional.Some(entry.Error.Error())
			}

			var request optional.Option[model.RequestKey]
			if reqStr, ok := entry.Attributes["request"]; ok {
				req, err := model.ParseRequestKey(reqStr)
				if err == nil {
					request = optional.Some(req)
				}
			}

			err = d.timeline.InsertLogEvent(ctx, &timelinedal.LogEvent{
				RequestKey:    request,
				DeploymentKey: deployment,
				Time:          entry.Time,
				Level:         int32(entry.Level.Severity()),
				Attributes:    entry.Attributes,
				Message:       entry.Message,
				Error:         errorStr,
			})
			if err != nil {
				fmt.Printf("failed to insert log entry: %v :: error: %v\n", entry, err)
			}
		case <-ctx.Done():
			return
		case <-time.After(1 * time.Second):
		}
	}
}
