package controller

import (
	"context"
	"fmt"
	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/model"
	"github.com/TBD54566975/ftl/backend/controller/internal/dal"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types"
	"time"
)

type logEntry struct {
	request types.Option[model.IngressRequestKey]
	log.Entry
}

var _ log.Sink = (*deploymentLogsSink)(nil)

func newDeploymentLogsSink(ctx context.Context, key model.DeploymentKey, dal *dal.DAL) *deploymentLogsSink {
	sink := &deploymentLogsSink{
		deploymentKey: key,
		logQueue:      make(chan logEntry, 10000),
		dal:           dal,
	}

	// Process logs in background
	go sink.processLogs(ctx)

	return sink
}

type deploymentLogsSink struct {
	deploymentKey model.DeploymentKey
	logQueue      chan logEntry
	dal           *dal.DAL
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

func (d *deploymentLogsSink) processLogs(ctx context.Context) {
	for {
		select {
		case entry := <-d.logQueue:
			var errorStr types.Option[string]
			if entry.Error != nil {
				errorStr = types.Some(entry.Error.Error())
			}

			err := d.dal.InsertLogEvent(ctx, &dal.LogEvent{
				RequestKey:    entry.request,
				DeploymentKey: d.deploymentKey,
				Time:          entry.Time,
				Level:         int32(entry.Level.Severity()),
				Attributes:    entry.Attributes,
				Message:       entry.Message,
				Error:         errorStr,
			})
			if err != nil {
				fmt.Printf("failed to insert log event: %v\n", err)
			}
		case <-ctx.Done():
			return
		case <-time.After(1 * time.Second):
		}
	}
}
