package model

import (
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/common/schema"
)

type CronJob struct {
	Key           CronJobKey
	DeploymentKey DeploymentKey
	Verb          schema.Ref
	Schedule      string
	StartTime     time.Time
	NextExecution time.Time
	LastExecution optional.Option[time.Time]
}
