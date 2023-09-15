package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/types"

	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/model"
	"github.com/TBD54566975/ftl/backend/controller/internal/dal"
)

var _ log.Sink = (*deploymentLogsSink)(nil)

func newDeploymentLogsSink(ctx context.Context, dal *dal.DAL) *deploymentLogsSink {
	sink := &deploymentLogsSink{
		logQueue: make(chan log.Entry, 10000),
		dal:      dal,
	}

	// Process logs in background
	go sink.processLogs(ctx)

	return sink
}

type deploymentLogsSink struct {
	logQueue chan log.Entry
	dal      *dal.DAL
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

func (d *deploymentLogsSink) processLogs(ctx context.Context) {
	for {
		select {
		case entry := <-d.logQueue:
			var deployment model.DeploymentName
			depStr, ok := entry.Attributes["deployment"]
			if !ok {
				continue
			}

			dep, err := model.ParseDeploymentName(depStr)
			if err != nil {
				continue
			}
			deployment = dep

			var errorStr types.Option[string]
			if entry.Error != nil {
				errorStr = types.Some(entry.Error.Error())
			}

			var request types.Option[model.RequestName]
			if reqStr, ok := entry.Attributes["request"]; ok {
				_, req, err := model.ParseRequestName(reqStr)
				if err == nil {
					request = types.Some(req)
				}
			}

			err = d.dal.InsertLogEvent(ctx, &dal.LogEvent{
				RequestName:    request,
				DeploymentName: deployment,
				Time:           entry.Time,
				Level:          int32(entry.Level.Severity()),
				Attributes:     entry.Attributes,
				Message:        entry.Message,
				Error:          errorStr,
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
