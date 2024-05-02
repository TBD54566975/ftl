package model

import (
	"time"

	"github.com/TBD54566975/ftl/backend/schema"
)

type CronJobState string

const (
	CronJobStateIdle      = "idle"
	CronJobStateExecuting = "executing"
)

type CronJob struct {
	Key           CronJobKey
	DeploymentKey DeploymentKey
	Verb          schema.Ref
	Schedule      string
	StartTime     time.Time
	NextExecution time.Time
	State         CronJobState
}
