// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package sql

import (
	"context"
	"time"

	"github.com/TBD54566975/ftl/internal/model"
)

type Querier interface {
	AsyncCallQueueDepth(ctx context.Context) (int64, error)
	CreateAsyncCall(ctx context.Context, arg CreateAsyncCallParams) (int64, error)
	CreateCronJob(ctx context.Context, arg CreateCronJobParams) error
	GetCronJobByKey(ctx context.Context, key model.CronJobKey) (GetCronJobByKeyRow, error)
	GetUnscheduledCronJobs(ctx context.Context, startTime time.Time) ([]GetUnscheduledCronJobsRow, error)
	IsCronJobPending(ctx context.Context, key model.CronJobKey, startTime time.Time) (bool, error)
	UpdateCronJobExecution(ctx context.Context, arg UpdateCronJobExecutionParams) error
}

var _ Querier = (*Queries)(nil)