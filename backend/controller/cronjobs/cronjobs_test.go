package cronjobs

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/either"
	"github.com/alecthomas/types/optional"
	"github.com/benbjohnson/clock"

	"github.com/TBD54566975/ftl/backend/controller/async"
	"github.com/TBD54566975/ftl/backend/controller/cronjobs/internal/dal"
	parentdal "github.com/TBD54566975/ftl/backend/controller/dal"
	dalmodel "github.com/TBD54566975/ftl/backend/controller/dal/model"
	"github.com/TBD54566975/ftl/backend/controller/encryption"
	"github.com/TBD54566975/ftl/backend/controller/leases"
	"github.com/TBD54566975/ftl/backend/controller/pubsub"
	"github.com/TBD54566975/ftl/backend/controller/scheduledtask"
	"github.com/TBD54566975/ftl/backend/controller/sql/sqltest"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/cron"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
)

func TestNewCronJobsForModule(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	clk := clock.NewMock()
	clk.Add(time.Second) // half way between cron job executions

	key := model.NewControllerKey("localhost", strconv.Itoa(8080+1))
	conn := sqltest.OpenForTesting(ctx, t)
	dal := dal.New(conn)

	uri := "fake-kms://CK6YwYkBElQKSAowdHlwZS5nb29nbGVhcGlzLmNvbS9nb29nbGUuY3J5cHRvLnRpbmsuQWVzR2NtS2V5EhIaEJy4TIQgfCuwxA3ZZgChp_wYARABGK6YwYkBIAE"
	encryption, err := encryption.New(ctx, conn, encryption.NewBuilder().WithKMSURI(optional.Some(uri)))
	assert.NoError(t, err)

	scheduler := scheduledtask.New(ctx, key, leases.NewFakeLeaser())
	pubSub := pubsub.New(conn, encryption, scheduler, optional.None[pubsub.AsyncCallListener]())
	parentDAL := parentdal.New(ctx, conn, encryption, pubSub)
	moduleName := "initial"
	jobsToCreate := newCronJobs(t, moduleName, "* * * * * *", clk, 2) // every minute

	deploymentKey, err := parentDAL.CreateDeployment(ctx, "go", &schema.Module{
		Name: moduleName,
	}, []dalmodel.DeploymentArtefact{}, []parentdal.IngressRoutingEntry{}, jobsToCreate)
	assert.NoError(t, err)
	err = parentDAL.ReplaceDeployment(ctx, deploymentKey, 1)
	assert.NoError(t, err)

	// Progress so that start_time is valid
	clk.Add(time.Second)
	cjs := NewForTesting(ctx, key, "test.com", encryption, *dal, clk)
	// All jobs need to be scheduled
	expectUnscheduledJobs(t, dal, clk, 2)
	unscheduledJobs, err := dal.GetUnscheduledCronJobs(ctx, clk.Now())
	assert.NoError(t, err)
	assert.Equal(t, len(unscheduledJobs), 2)

	// No async calls yet
	_, _, err = parentDAL.AcquireAsyncCall(ctx)
	assert.IsError(t, err, libdal.ErrNotFound)
	assert.EqualError(t, err, "no pending async calls: not found")

	err = cjs.scheduleCronJobs(ctx)
	assert.NoError(t, err)
	expectUnscheduledJobs(t, dal, clk, 0)
	for _, job := range jobsToCreate {
		j, err := dal.GetCronJobByKey(ctx, job.Key)
		assert.NoError(t, err)
		assert.Equal(t, j.StartTime, job.StartTime)
		assert.Equal(t, clk.Now().Add(time.Second), j.NextExecution)

		p, err := dal.IsCronJobPending(ctx, job.Key, job.StartTime)
		assert.NoError(t, err)
		assert.True(t, p)
	}
	// Now there should be async calls
	calls := []*parentdal.AsyncCall{}
	for i, job := range jobsToCreate {
		call, _, err := parentDAL.AcquireAsyncCall(ctx)
		assert.NoError(t, err)
		assert.Equal(t, job.Verb.ToRefKey(), call.Verb)
		assert.Equal(t, fmt.Sprintf("cron:%s", job.Key), call.Origin.String())
		assert.Equal(t, []byte("{}"), call.Request)
		assert.Equal(t, int64(len(jobsToCreate)-i), call.QueueDepth) // widdling down queue

		p, err := dal.IsCronJobPending(ctx, job.Key, job.StartTime)
		assert.NoError(t, err)
		assert.False(t, p)

		calls = append(calls, call)
	}
	clk.Add(time.Second)
	expectUnscheduledJobs(t, dal, clk, 0)
	// Complete all calls
	for _, call := range calls {
		callResult := either.LeftOf[string]([]byte("{}"))
		_, err = parentDAL.CompleteAsyncCall(ctx, call, callResult, func(tx *parentdal.DAL, isFinalResult bool) error {
			return nil
		})
		assert.NoError(t, err)
	}
	clk.Add(time.Second)
	expectUnscheduledJobs(t, dal, clk, 2)
	// Use the completion handler to schedule the next execution
	for _, call := range calls {
		origin, ok := call.Origin.(async.AsyncOriginCron)
		assert.True(t, ok)
		err = cjs.OnJobCompletion(ctx, origin.CronJobKey, false)
		assert.NoError(t, err)
	}
	expectUnscheduledJobs(t, dal, clk, 0)
	for i, job := range jobsToCreate {
		call, _, err := parentDAL.AcquireAsyncCall(ctx)
		assert.NoError(t, err)
		assert.Equal(t, job.Verb.ToRefKey(), call.Verb)
		assert.Equal(t, fmt.Sprintf("cron:%s", job.Key), call.Origin.String())
		assert.Equal(t, []byte("{}"), call.Request)
		assert.Equal(t, int64(len(jobsToCreate)-i), call.QueueDepth) // widdling down queue

		assert.Equal(t, call.ScheduledAt, clk.Now())

		p, err := dal.IsCronJobPending(ctx, job.Key, job.StartTime)
		assert.NoError(t, err)
		assert.False(t, p)
	}
}

func expectUnscheduledJobs(t *testing.T, dal *dal.DAL, clk *clock.Mock, count int) {
	t.Helper()
	unscheduledJobs, err := dal.GetUnscheduledCronJobs(context.Background(), clk.Now())
	assert.NoError(t, err)
	assert.Equal(t, len(unscheduledJobs), count)
}

func newCronJobs(t *testing.T, moduleName string, cronPattern string, clock clock.Clock, count int) []model.CronJob {
	t.Helper()
	newJobs := []model.CronJob{}
	for i := range count {
		now := clock.Now()
		pattern, err := cron.Parse(cronPattern)
		assert.NoError(t, err)
		next, err := cron.NextAfter(pattern, now, false)
		assert.NoError(t, err)
		newJobs = append(newJobs, model.CronJob{
			Key:           model.NewCronJobKey(moduleName, fmt.Sprintf("verb%dCron", i)),
			Verb:          schema.Ref{Module: moduleName, Name: fmt.Sprintf("verb%dCron", i)},
			Schedule:      pattern.String(),
			StartTime:     now,
			NextExecution: next,
		})
	}
	return newJobs
}
