package model

import (
	"time"
)

type CronJobState string

const (
	CronJobStateIdle      = "idle"
	CronJobStateExecuting = "executing"
)

type CronJob struct {
	Key           CronJobKey
	DeploymentKey DeploymentKey
	Verb          VerbRef
	Schedule      string
	StartTime     time.Time
	NextExecution time.Time
	State         CronJobState
}
