package model

import (
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/schema"
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
