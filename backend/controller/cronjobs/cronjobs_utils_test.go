package cronjobs

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	db "github.com/TBD54566975/ftl/backend/controller/dal"
	"github.com/TBD54566975/ftl/backend/controller/scheduledtask"
	"github.com/TBD54566975/ftl/internal/cron"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/alecthomas/assert/v2"
	"github.com/benbjohnson/clock"
	"github.com/jpillora/backoff"
)

type mockDAL struct {
	lock            sync.Mutex
	clock           clock.Clock
	jobs            []model.CronJob
	attemptCountMap map[string]int
}

func (d *mockDAL) GetCronJobs(ctx context.Context) ([]model.CronJob, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.jobs, nil
}

func (d *mockDAL) createCronJob(deploymentKey model.DeploymentKey, module string, verb string, schedule string, startTime time.Time, nextExecution time.Time) {
	d.lock.Lock()
	defer d.lock.Unlock()

	job := model.CronJob{
		Key:           model.NewCronJobKey(module, verb),
		DeploymentKey: deploymentKey,
		Verb:          model.VerbRef{Module: module, Name: verb},
		Schedule:      schedule,
		StartTime:     startTime,
		NextExecution: nextExecution,
		State:         model.CronJobStateIdle,
	}
	d.jobs = append(d.jobs, job)
}

func (d *mockDAL) indexForJob(job model.CronJob) (int, error) {
	for i, j := range d.jobs {
		if j.Key.String() == job.Key.String() {
			return i, nil
		}
	}
	return -1, fmt.Errorf("job not found")
}

func (d *mockDAL) StartCronJobs(ctx context.Context, jobs []model.CronJob) (attemptedJobs []db.AttemptedCronJob, err error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	attemptedJobs = []db.AttemptedCronJob{}
	now := d.clock.Now()

	for _, inputJob := range jobs {
		i, err := d.indexForJob(inputJob)
		if err != nil {
			return nil, err
		}
		job := d.jobs[i]
		if !job.NextExecution.After(now) && job.State == model.CronJobStateIdle {
			job.State = model.CronJobStateExecuting
			job.StartTime = d.clock.Now()
			d.jobs[i] = job
			attemptedJobs = append(attemptedJobs, db.AttemptedCronJob{
				CronJob:           job,
				DidStartExecution: true,
				HasMinReplicas:    true,
			})
		} else {
			attemptedJobs = append(attemptedJobs, db.AttemptedCronJob{
				CronJob:           job,
				DidStartExecution: false,
				HasMinReplicas:    true,
			})
		}
		d.attemptCountMap[job.Key.String()]++
	}
	return attemptedJobs, nil
}

func (d *mockDAL) EndCronJob(ctx context.Context, job model.CronJob, next time.Time) (model.CronJob, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	i, err := d.indexForJob(job)
	if err != nil {
		return model.CronJob{}, err
	}
	internalJob := d.jobs[i]
	if internalJob.State != model.CronJobStateExecuting {
		return model.CronJob{}, fmt.Errorf("job can not be stopped, it isnt running")
	}
	if internalJob.StartTime != job.StartTime {
		return model.CronJob{}, fmt.Errorf("job can not be stopped, start time does not match")
	}
	internalJob.State = model.CronJobStateIdle
	internalJob.NextExecution = next
	d.jobs[i] = internalJob
	return internalJob, nil
}

func (d *mockDAL) GetStaleCronJobs(ctx context.Context, duration time.Duration) ([]model.CronJob, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	return slices.Filter(d.jobs, func(job model.CronJob) bool {
		return d.clock.Now().After(job.StartTime.Add(duration))
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
	key       model.ControllerKey
	dal       DAL
	clock     clock.Clock
	mockClock *clock.Mock // only set when clock is a mock
	cronJobs  *Service
}

func newJobs(t *testing.T, moduleName string, cronPattern string, clock clock.Clock, count int) []model.CronJob {
	t.Helper()
	newJobs := []model.CronJob{}
	for i := range count {
		now := clock.Now()
		pattern, err := cron.Parse(cronPattern)
		assert.NoError(t, err)
		next, err := cron.NextAfter(pattern, now, false)
		assert.NoError(t, err)
		newJobs = append(newJobs, model.CronJob{
			Key:           model.NewCronJobKey(moduleName, fmt.Sprintf("verb%d", i)),
			Verb:          model.VerbRef{Module: moduleName, Name: fmt.Sprintf("verb%d", i)},
			Schedule:      pattern.String(),
			StartTime:     now,
			NextExecution: next,
			State:         model.CronJobStateIdle,
		})
	}
	return newJobs
}

func newControllers(ctx context.Context, count int, dal DAL, clockFactory func() clock.Clock, call ExecuteCallFunc) []*controller {
	controllers := []*controller{}
	for i := range count {
		key := model.NewControllerKey("localhost", strconv.Itoa(8080+i))
		clk := clockFactory()
		controller := &controller{
			key:   key,
			dal:   dal,
			clock: clk,
			cronJobs: NewForTesting(ctx,
				key, "test.com",
				Config{Timeout: time.Minute * 5},
				dal,
				&mockScheduler{},
				call,
				clk),
		}
		if mockClock, ok := clk.(*clock.Mock); ok {
			controller.mockClock = mockClock
		}
		controllers = append(controllers, controller)
	}

	time.Sleep(time.Millisecond * 100)

	for _, c := range controllers {
		s := c.cronJobs
		go func() {
			s.UpdatedControllerList(ctx, slices.Map(controllers, func(ctrl *controller) db.Controller {
				return db.Controller{
					Key: ctrl.key,
				}
			}))
			_, _ = s.syncJobs(ctx)
		}()
	}

	time.Sleep(time.Millisecond * 100)

	return controllers
}
