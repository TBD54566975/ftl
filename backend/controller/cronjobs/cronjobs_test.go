package cronjobs

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/TBD54566975/ftl/backend/controller/dal"
	"github.com/TBD54566975/ftl/backend/controller/scheduledtask"
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

type mockDAL struct {
	lock            sync.Mutex
	clock           *clock.Mock
	jobs            []dal.CronJob
	attemptCountMap map[string]int
}

func (d *mockDAL) GetCronJobs(ctx context.Context) ([]dal.CronJob, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.jobs, nil
}

func (d *mockDAL) createCronJob(deploymentKey model.DeploymentKey, module string, verb string, schedule string, startTime time.Time, nextExecution time.Time) {
	d.lock.Lock()
	defer d.lock.Unlock()

	job := dal.CronJob{
		Key:           model.NewCronJobKey(module, verb),
		DeploymentKey: deploymentKey,
		Ref:           schema.Ref{Module: module, Name: verb},
		Schedule:      schedule,
		StartTime:     startTime,
		NextExecution: nextExecution,
		State:         dal.JobStateIdle,
	}
	d.jobs = append(d.jobs, job)
}

func (d *mockDAL) indexForJob(job dal.CronJob) (int, error) {
	for i, j := range d.jobs {
		if j.DeploymentKey.String() == job.DeploymentKey.String() && j.Ref.Name == job.Ref.Name {
			return i, nil
		}
	}
	return -1, fmt.Errorf("job not found")
}

func (d *mockDAL) StartCronJobs(ctx context.Context, jobs []dal.CronJob) (attemptedJobs []dal.AttemptedCronJob, err error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	attemptedJobs = []dal.AttemptedCronJob{}
	now := (*d.clock).Now()

	for _, inputJob := range jobs {
		i, err := d.indexForJob(inputJob)
		if err != nil {
			return nil, err
		}
		job := d.jobs[i]
		if !job.NextExecution.After(now) && job.State == dal.JobStateIdle {
			job.State = dal.JobStateExecuting
			job.StartTime = (*d.clock).Now()
			d.jobs[i] = job
			attemptedJobs = append(attemptedJobs, dal.AttemptedCronJob{
				CronJob:           job,
				DidStartExecution: true,
				HasMinReplicas:    true,
			})
		} else {
			attemptedJobs = append(attemptedJobs, dal.AttemptedCronJob{
				CronJob:           job,
				DidStartExecution: false,
				HasMinReplicas:    true,
			})
		}
		d.attemptCountMap[job.Ref.String()]++
	}
	return attemptedJobs, nil
}

func (d *mockDAL) EndCronJob(ctx context.Context, job dal.CronJob, next time.Time) (dal.CronJob, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	i, err := d.indexForJob(job)
	if err != nil {
		return dal.CronJob{}, err
	}
	internalJob := d.jobs[i]
	if internalJob.State != dal.JobStateExecuting {
		return dal.CronJob{}, fmt.Errorf("job can not be stopped, it isnt running")
	}
	if internalJob.StartTime != job.StartTime {
		return dal.CronJob{}, fmt.Errorf("job can not be stopped, start time does not match")
	}
	internalJob.State = dal.JobStateIdle
	internalJob.NextExecution = next
	d.jobs[i] = internalJob
	return internalJob, nil
}

func (d *mockDAL) GetStaleCronJobs(ctx context.Context, duration time.Duration) ([]dal.CronJob, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	return slices.Filter(d.jobs, func(job dal.CronJob) bool {
		return (*d.clock).Now().After(job.StartTime.Add(duration))
	}), nil
}

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
	DAL      DAL
	cronJobs *Service
}

func TestService(t *testing.T) {
	t.Parallel()
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	config := Config{Timeout: time.Minute * 5}
	clock := clock.NewMock()
	mockDal := &mockDAL{
		clock:           clock,
		lock:            sync.Mutex{},
		attemptCountMap: map[string]int{},
	}
	scheduler := &mockScheduler{}

	verbCallCount := map[string]int{}
	verbCallCountLock := sync.Mutex{}

	// initial jobs
	for i := range 20 {
		deploymentKey := model.NewDeploymentKey("initial")
		now := clock.Now()
		cronStr := "*/10 * * * * * *"
		pattern, err := cron.Parse(cronStr)
		assert.NoError(t, err)
		next, err := cron.NextAfter(pattern, now, false)
		assert.NoError(t, err)
		mockDal.createCronJob(deploymentKey, "initial", fmt.Sprintf("verb%d", i), cronStr, now, next)
	}

	controllers := []*controller{}
	for i := range 5 {
		key := model.NewControllerKey("localhost", strconv.Itoa(8080+i))
		controller := &controller{
			key: key,
			DAL: mockDal,
			cronJobs: NewForTesting(ctx, key, &url.URL{Host: "test.com"}, config, mockDal, scheduler, func(ctx context.Context, r *connect.Request[ftlv1.CallRequest], o optional.Option[model.RequestKey], s string) (*connect.Response[ftlv1.CallResponse], error) {
				verbRef := schema.RefFromProto(r.Msg.Verb)

				verbCallCountLock.Lock()
				verbCallCount[verbRef.Name]++
				verbCallCountLock.Unlock()

				return &connect.Response[ftlv1.CallResponse]{}, nil
			}, clock),
		}
		controllers = append(controllers, controller)
	}

	time.Sleep(time.Millisecond * 100)

	for _, c := range controllers {
		go func() {
			c.cronJobs.UpdatedControllerList(ctx, slices.Map(controllers, func(ctrl *controller) dal.Controller {
				return dal.Controller{
					Key: ctrl.key,
				}
			}))
			_, _ = c.cronJobs.resetJobs(ctx)
		}()
	}

	clock.Add(time.Second * 5)
	time.Sleep(time.Millisecond * 100)
	for range 3 {
		clock.Add(time.Second * 10)
		time.Sleep(time.Millisecond * 100)
	}

	for _, j := range mockDal.jobs {
		count := verbCallCount[j.Ref.Name]
		assert.Equal(t, count, 3, "expected verb %s to be called 3 times", j.Ref.Name)
	}
}
