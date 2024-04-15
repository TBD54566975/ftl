package cronjobs

import (
	"context"
	"fmt"
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
	xslices "golang.org/x/exp/slices"
)

type mockDAL struct {
	lock            sync.Mutex
	clock           *clock.Mock
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

func (d *mockDAL) StartCronJobs(ctx context.Context, jobs []model.CronJob) (attemptedJobs []dal.AttemptedCronJob, err error) {
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
		if !job.NextExecution.After(now) && job.State == model.CronJobStateIdle {
			job.State = model.CronJobStateExecuting
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
	clock    *clock.Mock
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
			key:   key,
			DAL:   mockDal,
			clock: clock,
			cronJobs: NewForTesting(ctx, key, "test.com", config, mockDal, scheduler, func(ctx context.Context, r *connect.Request[ftlv1.CallRequest], o optional.Option[model.RequestKey], s string) (*connect.Response[ftlv1.CallResponse], error) {
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
			_, _ = c.cronJobs.syncJobs(ctx)
		}()
	}

	clock.Add(time.Second * 5)
	time.Sleep(time.Millisecond * 100)
	for range 3 {
		clock.Add(time.Second * 10)
		time.Sleep(time.Millisecond * 100)
	}

	for _, j := range mockDal.jobs {
		count := verbCallCount[j.Verb.Name]
		assert.Equal(t, count, 3, "expected verb %s to be called 3 times", j.Verb.Name)
	}
}

func TestHashRing(t *testing.T) {
	// This test uses multiple mock clocks to progress time for each controller individually
	// This allows us to compare attempts for each cron job and know which controller attempted it
	t.Parallel()
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	config := Config{Timeout: time.Minute * 5}
	mockDal := &mockDAL{
		clock:           clock.NewMock(),
		lock:            sync.Mutex{},
		attemptCountMap: map[string]int{},
	}
	scheduler := &mockScheduler{}

	verbCallCount := map[string]int{}
	verbCallCountLock := sync.Mutex{}

	// initial jobs
	for i := range 100 {
		deploymentKey := model.NewDeploymentKey("initial")
		now := mockDal.clock.Now()
		cronStr := "*/10 * * * * * *"
		pattern, err := cron.Parse(cronStr)
		assert.NoError(t, err)
		next, err := cron.NextAfter(pattern, now, false)
		assert.NoError(t, err)
		mockDal.createCronJob(deploymentKey, "initial", fmt.Sprintf("verb%d", i), cronStr, now, next)
	}

	controllers := []*controller{}
	for i := range 20 {
		key := model.NewControllerKey("localhost", strconv.Itoa(8080+i))
		clock := clock.NewMock()
		controller := &controller{
			key:   key,
			DAL:   mockDal,
			clock: clock,
			cronJobs: NewForTesting(ctx, key, "test.com", config, mockDal, scheduler, func(ctx context.Context, r *connect.Request[ftlv1.CallRequest], o optional.Option[model.RequestKey], s string) (*connect.Response[ftlv1.CallResponse], error) {
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
			_, _ = c.cronJobs.syncJobs(ctx)
		}()
	}
	time.Sleep(time.Millisecond * 100)

	// progress time for each controller one at a time, noting which verbs got attempted each time
	// to build a map of verb to controller keys
	controllersForVerbs := map[string][]model.ControllerKey{}
	for _, c := range controllers {
		beforeAttemptCount := map[string]int{}
		for k, v := range mockDal.attemptCountMap {
			beforeAttemptCount[k] = v
		}

		c.clock.Add(time.Second * 15)
		time.Sleep(time.Millisecond * 100)

		for k, v := range mockDal.attemptCountMap {
			if beforeAttemptCount[k] == v {
				continue
			}
			controllersForVerbs[k] = append(controllersForVerbs[k], c.key)
		}
	}

	// Check if each job has the same key list
	// Theoretically this is is possible for all jobs to have the same assigned controllers, but with 100 jobs and 20 controllers, this is unlikely
	keys := []string{}
	hasFoundNonMatchingKeys := false
	for v, k := range controllersForVerbs {
		assert.Equal(t, len(k), 2, "expected verb %s to be attempted by 2 controllers", v)

		kStrs := slices.Map(k, func(k model.ControllerKey) string { return k.String() })
		xslices.Sort(kStrs)
		if len(keys) == 0 {
			keys = kStrs
			continue
		}

		if hasFoundNonMatchingKeys == false {
			for keyIdx, keyStr := range kStrs {
				if keys[keyIdx] != keyStr {
					hasFoundNonMatchingKeys = true
				}
			}
		}
	}
	assert.True(t, hasFoundNonMatchingKeys, "expected at least one verb to have different controllers assigned")
}
