//go:build integration

package integration

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/TBD54566975/ftl/backend/controller/cronjobs"
	db "github.com/TBD54566975/ftl/backend/controller/dal"
	"github.com/TBD54566975/ftl/backend/controller/scheduledtask"
	"github.com/TBD54566975/ftl/backend/controller/sql/sqltest"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/cron"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
	"github.com/benbjohnson/clock"
	"github.com/jpillora/backoff"
)

type mockScheduler struct {
}

func (s *mockScheduler) Singleton(retry backoff.Backoff, job scheduledtask.Job) {
	// do nothing
}

func (s *mockScheduler) Parallel(retry backoff.Backoff, job scheduledtask.Job) {
	// do nothing
}

type controller struct {
	key      model.ControllerKey
	dal      *db.DAL
	clock    *clock.Mock
	cronJobs *cronjobs.Service
}

func TestServiceWithDB(t *testing.T) {
	t.Parallel()
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	conn := sqltest.OpenForTesting(ctx, t)
	dal, err := db.New(ctx, conn)
	assert.NoError(t, err)

	config := cronjobs.Config{Timeout: time.Minute * 5}
	clock := clock.NewMock()
	scheduler := &mockScheduler{}

	verbCallCount := map[string]int{}
	verbCallCountLock := sync.Mutex{}

	// initial jobs
	jobsToCreate := []model.CronJob{}
	for i := range 20 {
		now := clock.Now()
		cronStr := "*/10 * * * * * *"
		pattern, err := cron.Parse(cronStr)
		assert.NoError(t, err)
		next, err := cron.NextAfter(pattern, now, false)
		assert.NoError(t, err)
		jobsToCreate = append(jobsToCreate, model.CronJob{
			Key:           model.NewCronJobKey("initial", fmt.Sprintf("verb%d", i)),
			Verb:          model.VerbRef{Module: "initial", Name: fmt.Sprintf("verb%d", i)},
			Schedule:      pattern.String(),
			StartTime:     now,
			NextExecution: next,
			State:         model.CronJobStateIdle,
		})
	}

	_, err = dal.CreateDeployment(ctx, "go", &schema.Module{
		Name: "initial",
	}, []db.DeploymentArtefact{}, []db.IngressRoutingEntry{}, jobsToCreate)
	assert.NoError(t, err)

	controllers := []*controller{}
	for i := range 5 {
		key := model.NewControllerKey("localhost", strconv.Itoa(8080+i))
		controller := &controller{
			key:   key,
			dal:   dal,
			clock: clock,
			cronJobs: cronjobs.NewForTesting(ctx, key, "test.com", config, dal, scheduler, func(ctx context.Context, r *connect.Request[ftlv1.CallRequest], o optional.Option[model.RequestKey], s string) (*connect.Response[ftlv1.CallResponse], error) {
				verbRef := schema.RefFromProto(r.Msg.Verb)

				verbCallCountLock.Lock()
				verbCallCount[verbRef.Name]++
				verbCallCountLock.Unlock()

				return &connect.Response[ftlv1.CallResponse]{}, nil
			}, clock),
		}
		controllers = append(controllers, controller)
	}

	time.Sleep(time.Millisecond * 1000)

	for _, c := range controllers {
		go func() {
			c.cronJobs.UpdatedControllerList(ctx, slices.Map(controllers, func(ctrl *controller) db.Controller {
				return db.Controller{
					Key: ctrl.key,
				}
			}))
			_, _ = c.cronJobs.SyncJobs(ctx)
		}()
	}

	clock.Add(time.Second * 5)
	time.Sleep(time.Millisecond * 100)
	for range 3 {
		clock.Add(time.Second * 10)
		time.Sleep(time.Millisecond * 100)
	}

	for _, j := range jobsToCreate {
		count := verbCallCount[j.Verb.Name]
		assert.Equal(t, count, 3, "expected verb %s to be called 3 times", j.Verb.Name)
	}
}
