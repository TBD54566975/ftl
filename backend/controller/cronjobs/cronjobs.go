package cronjobs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"connectrpc.com/connect"
	"github.com/TBD54566975/ftl/backend/controller/dal"
	"github.com/TBD54566975/ftl/backend/controller/scheduledtask"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/cron"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/alecthomas/atomic"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/pubsub"
	"github.com/benbjohnson/clock"
	"github.com/jpillora/backoff"
	"github.com/serialx/hashring"
	sl "golang.org/x/exp/slices"
)

const (
	controllersPerJob              = 2
	jobResetInterval               = time.Minute
	newJobHashRingOverrideInterval = time.Minute + time.Second*20
)

type Config struct {
	Timeout time.Duration
}

type jobChangeType int

const (
	resetJobs jobChangeType = iota
	finishedJobs
	updatedHashring
)

type jobChange struct {
	changeType         jobChangeType
	jobs               []dal.CronJob
	addedDeploymentKey optional.Option[model.DeploymentKey]
}

type hashRingState struct {
	hashRing    *hashring.HashRing
	controllers []dal.Controller
	idx         int
}

type DAL interface {
	GetCronJobs(ctx context.Context) ([]dal.CronJob, error)
	StartCronJobs(ctx context.Context, jobs []dal.CronJob) (attemptedJobs []dal.AttemptedCronJob, err error)
	EndCronJob(ctx context.Context, job dal.CronJob, next time.Time) (dal.CronJob, error)
	GetStaleCronJobs(ctx context.Context, duration time.Duration) ([]dal.CronJob, error)
}

type Scheduler interface {
	Singleton(retry backoff.Backoff, job scheduledtask.Job)
	Parallel(retry backoff.Backoff, job scheduledtask.Job)
}

type ExecuteCallFunc func(context.Context, *connect.Request[ftlv1.CallRequest], optional.Option[model.RequestKey], string) (*connect.Response[ftlv1.CallResponse], error)

type Service struct {
	config    Config
	key       model.ControllerKey
	originURL *url.URL

	dal       DAL
	scheduler Scheduler
	call      ExecuteCallFunc

	clock      clock.Clock
	jobChanges *pubsub.Topic[jobChange]

	hashRingState atomic.Value[*hashRingState]
}

func New(ctx context.Context, key model.ControllerKey, originURL *url.URL, config Config, dal DAL, scheduler Scheduler, call ExecuteCallFunc) *Service {
	return NewForTesting(ctx, key, originURL, config, dal, scheduler, call, clock.New())
}

func NewForTesting(ctx context.Context, key model.ControllerKey, originURL *url.URL, config Config, dal DAL, scheduler Scheduler, call ExecuteCallFunc, clock clock.Clock) *Service {
	svc := &Service{
		config:     config,
		key:        key,
		originURL:  originURL,
		dal:        dal,
		scheduler:  scheduler,
		call:       call,
		clock:      clock,
		jobChanges: pubsub.New[jobChange](),
	}
	svc.UpdatedControllerList(ctx, nil)

	svc.scheduler.Parallel(backoff.Backoff{Min: time.Second, Max: jobResetInterval}, svc.resetJobs)
	svc.scheduler.Singleton(backoff.Backoff{Min: time.Second, Max: time.Minute}, svc.killOldJobs)

	go svc.watchForUpdates(ctx)

	return svc
}

func (s *Service) NewCronJobsForModule(ctx context.Context, module *schemapb.Module) (jobs []dal.CronJob, err error) {
	logger := log.FromContext(ctx)

	start := s.clock.Now().UTC()
	newJobs := []dal.CronJob{}
	for _, decl := range module.Decls {
		if verb, ok := decl.Value.(*schemapb.Decl_Verb); ok {
			for _, metadata := range verb.Verb.Metadata {
				if cronMetadata, ok := metadata.Value.(*schemapb.Metadata_CronJob); ok {
					cronStr := cronMetadata.CronJob.Cron
					schedule, err := cron.Parse(cronStr)
					if err != nil {
						logger.Errorf(err, "failed to parse cron schedule %q", cronStr)
						continue
					}
					next, err := cron.NextAfter(schedule, start, false)
					if err != nil {
						logger.Errorf(err, "failed to calculate next execution for cron job %v:%v with schedule %q", module.Name, verb.Verb.Name, schedule)
						continue
					}
					newJobs = append(newJobs, dal.CronJob{
						Key:           model.NewCronJobKey(module.Name, verb.Verb.Name),
						Ref:           schema.Ref{Module: module.Name, Name: verb.Verb.Name},
						Schedule:      cronStr,
						StartTime:     start,
						NextExecution: next,
						State:         dal.JobStateIdle,
						// DeploymentKey: Filled in by DAL
					})
				}
			}
		}
	}
	return newJobs, nil
}

func (s *Service) CreatedOrReplacedDeloyment(ctx context.Context, newDeploymentKey model.DeploymentKey) {
	// Rather than finding old/new cronjobs and updating our state, we can just reset the list of jobs
	_ = s.resetJobsWithNewDeploymentKey(ctx, optional.Some(newDeploymentKey))
}

// resetJobs is run periodically via a scheduled task
func (s *Service) resetJobs(ctx context.Context) (time.Duration, error) {
	err := s.resetJobsWithNewDeploymentKey(ctx, optional.None[model.DeploymentKey]())
	if err != nil {
		return 0, err
	}
	return jobResetInterval, nil
}

// resetJobsWithNewDeploymentKey resets the list of jobs and marks the deployment key as added so that it can overrule the hash ring for a short time.
func (s *Service) resetJobsWithNewDeploymentKey(ctx context.Context, deploymentKey optional.Option[model.DeploymentKey]) error {
	logger := log.FromContext(ctx)

	jobs, err := s.dal.GetCronJobs(ctx)
	if err != nil {
		logger.Errorf(err, "failed to get cron jobs")
		return fmt.Errorf("%s: %w", "failed to get cron jobs", err)
	}
	s.jobChanges.Publish(jobChange{
		changeType:         resetJobs,
		jobs:               jobs,
		addedDeploymentKey: deploymentKey,
	})
	return nil
}

func (s *Service) executeJob(ctx context.Context, job dal.CronJob) {
	logger := log.FromContext(ctx)
	requestBody := map[string]any{}
	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		logger.Errorf(err, "could not build cron job body %v:%v", job.DeploymentKey, job.Ref.String())
		return
	}

	req := connect.NewRequest(&ftlv1.CallRequest{
		Verb: &schemapb.Ref{Module: job.Ref.Module, Name: job.Ref.Name},
		Body: requestJSON,
	})

	requestKey := model.NewRequestKey(model.OriginCron, fmt.Sprintf("%s-%s", job.Ref.Module, job.Ref.Name))

	callCtx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()
	_, err = s.call(callCtx, req, optional.Some(requestKey), s.originURL.Host)
	if err != nil {
		logger.Errorf(err, "failed to execute cron job %s", job.Ref.String())
	}

	schedule, err := cron.Parse(job.Schedule)
	if err != nil {
		logger.Errorf(err, "failed to parse cron schedule %q", job.Schedule)
		return
	}
	next, err := cron.NextAfter(schedule, s.clock.Now().UTC(), false)
	if err != nil {
		logger.Errorf(err, "failed to calculate next execution for cron job %s with schedule %q", job.Ref.String(), job.Schedule)
	}

	updatedJob, err := s.dal.EndCronJob(ctx, job, next)
	if err != nil {
		logger.Errorf(err, "failed to end cronjob %s", job.Ref.String())
	} else {
		s.jobChanges.Publish(jobChange{
			changeType: finishedJobs,
			jobs:       []dal.CronJob{updatedJob},
		})
	}
}

// killOldJobs looks for jobs that have been executing for too long
// This is the hard timout which happens after the usual timeout plus a grace period for the soft timeout to occur (context's timeout which cancel the call)
func (s *Service) killOldJobs(ctx context.Context) (time.Duration, error) {
	logger := log.FromContext(ctx)
	staleJobs, err := s.dal.GetStaleCronJobs(ctx, s.config.Timeout+time.Minute)
	if err != nil {
		return 0, err
	}

	updatedJobs := []dal.CronJob{}
	for _, stale := range staleJobs {
		start := s.clock.Now().UTC()
		pattern, err := cron.Parse(stale.Schedule)
		if err != nil {
			logger.Errorf(err, "Could not kill stale cron job %s because schedule could not be parsed: %q", stale.Ref.String(), stale.Schedule)
			continue
		}
		next, err := cron.NextAfter(pattern, start, false)
		if err != nil {
			logger.Errorf(err, "Could not kill stale cron job %s because next date could not be calculated: %q", stale.Ref.String(), stale.Schedule)
			continue
		}

		updated, err := s.dal.EndCronJob(ctx, stale, next)
		if err != nil {
			logger.Errorf(err, "Could not kill stale cron job %s because: %v", stale.Ref.String(), err)
			continue
		}
		logger.Warnf("Killed stale cron job %s", stale.Ref.String())
		updatedJobs = append(updatedJobs, updated)
	}

	s.jobChanges.Publish(jobChange{
		changeType: finishedJobs,
		jobs:       updatedJobs,
	})

	return time.Minute, nil
}

// watchForUpdates is the centralized place that handles:
// - the list of known jobs and their state
// - executing jobs when they are due
// - reacting to events that change the list of jobs, deployments or hash ring
func (s *Service) watchForUpdates(ctx context.Context) {
	logger := log.FromContext(ctx)

	jobChanges := make(chan jobChange, 128)
	s.jobChanges.Subscribe(jobChanges)
	defer s.jobChanges.Unsubscribe(jobChanges)

	state := &State{
		executing:    map[string]bool{},
		newJobs:      map[string]time.Time{},
		blockedUntil: s.clock.Now(),
	}

	for {
		sl.SortFunc(state.jobs, func(i, j dal.CronJob) int {
			return s.sortJobs(state, i, j)
		})

		now := s.clock.Now()
		next := now.Add(time.Hour) // should never be reached, expect a different signal long beforehand
		for _, j := range state.jobs {
			if possibleNext, err := s.nextAttemptForJob(j, state, false); err == nil {
				next = possibleNext
				break
			}
		}

		if next.Before(state.blockedUntil) {
			next = state.blockedUntil
			logger.Tracef("loop blocked for %vs", next.Sub(now))
		} else if next.Sub(now) < time.Second {
			next = now.Add(time.Second)
			logger.Tracef("loop while gated for 1s")
		} else if next.Sub(now) > time.Minute*59 {
			logger.Tracef("loop while idling")
		} else {
			logger.Tracef("loop with next %v, %d jobs", next.Sub(now), len(state.jobs))
		}

		select {
		case <-ctx.Done():
			return
		case <-s.clock.After(next.Sub(now)):
			// Try starting jobs in db
			jobsToAttempt := slices.Filter(state.jobs, func(j dal.CronJob) bool {
				if n, err := s.nextAttemptForJob(j, state, true); err == nil {
					return !n.After(s.clock.Now().UTC())
				}
				return false
			})
			jobResults, err := s.dal.StartCronJobs(ctx, jobsToAttempt)
			if err != nil {
				logger.Errorf(err, "failed to start cron jobs in db")
				state.blockedUntil = s.clock.Now().Add(time.Second * 5)
				continue
			}

			// Start jobs that were successfully updated
			updatedJobs := []dal.CronJob{}
			removedDeploymentKeys := map[string]model.DeploymentKey{}

			for _, job := range jobResults {
				updatedJobs = append(updatedJobs, job.CronJob)
				if !job.DidStartExecution {
					continue
				}
				if !job.HasMinReplicas {
					// We successfully updated the db to start this job but the deployment has min replicas set to 0
					// We need to update the db to end this job
					removedDeploymentKeys[job.DeploymentKey.String()] = job.DeploymentKey
					_, err := s.dal.EndCronJob(ctx, job.CronJob, next)
					if err != nil {
						logger.Errorf(err, "failed to end cronjob %s", job.Ref.String())
					}
					continue
				}
				logger.Infof("executing job %s", job.Ref.String())
				state.startedExecutingJob(job.CronJob)
				go s.executeJob(ctx, job.CronJob)
			}

			// Update job list
			state.updateJobs(updatedJobs)
			for _, key := range removedDeploymentKeys {
				state.removeDeploymentKey(key)
			}
		case event := <-jobChanges:
			switch event.changeType {
			case resetJobs:
				logger.Tracef("resetting job list: %d jobs", len(event.jobs))
				state.reset(event.jobs, event.addedDeploymentKey)
			case finishedJobs:
				logger.Tracef("updating %d jobs", len(event.jobs))
				state.updateJobs(event.jobs)
			case updatedHashring:
				// do another cycle through the loop to see if new jobs need to be scheduled
			}
		}
	}
}

func (s *Service) sortJobs(state *State, i, j dal.CronJob) int {
	iNext, err := s.nextAttemptForJob(i, state, false)
	if err != nil {
		return 1
	}
	jNext, err := s.nextAttemptForJob(j, state, false)
	if err != nil {
		return -1
	}
	return iNext.Compare(jNext)
}

func (s *Service) nextAttemptForJob(job dal.CronJob, state *State, allowsNow bool) (time.Time, error) {
	if !s.isResponsibleForJob(job, state) {
		return s.clock.Now(), fmt.Errorf("controller is not responsible for job")
	}
	if job.State == dal.JobStateExecuting {
		if state.isExecutingInCurrentController(job) {
			// return a time in the future, meaning don't schedule at this time
			return s.clock.Now(), fmt.Errorf("controller is already waiting for job to finish")
		}
		// We don't know when the other controller will finish this job
		// We should check again when the next execution date is assuming the job finishes
		pattern, err := cron.Parse(job.Schedule)
		if err != nil {
			return s.clock.Now(), fmt.Errorf("failed to parse cron schedule %q", job.Schedule)
		}
		next, err := cron.NextAfter(pattern, s.clock.Now().UTC(), allowsNow)
		if err == nil {
			return next, nil
		}
	}
	return job.NextExecution, nil
}

// UpdatedControllerList synchronises the hash ring with the active controllers.
func (s *Service) UpdatedControllerList(ctx context.Context, controllers []dal.Controller) {
	logger := log.FromContext(ctx).Scope("cron")
	controllerIdx := -1
	for idx, controller := range controllers {
		if controller.Key.String() == s.key.String() {
			controllerIdx = idx
			break
		}
	}
	if controllerIdx == -1 {
		logger.Tracef("controller %q not found in list of controllers", s.key)
	}

	oldState := s.hashRingState.Load()
	if oldState != nil && len(oldState.controllers) == len(controllers) {
		hasChanged := false
		for idx, new := range controllers {
			old := oldState.controllers[idx]
			if new.Key.String() != old.Key.String() {
				hasChanged = true
				break
			}
		}
		if !hasChanged {
			return
		}
	}

	hashRing := hashring.New(slices.Map(controllers, func(c dal.Controller) string { return c.Key.String() }))
	s.hashRingState.Store(&hashRingState{
		hashRing:    hashRing,
		controllers: controllers,
		idx:         controllerIdx,
	})

	s.jobChanges.Publish(jobChange{
		changeType: updatedHashring,
	})
}

// isResponsibleForJob indicates whether a this service should be responsible for attempting jobs,
// or if enough other controllers will handle it. This allows us to spread the job load across controllers.
func (s *Service) isResponsibleForJob(job dal.CronJob, state *State) bool {
	if state.isJobTooNewForHashRing(job) {
		return true
	}
	hashringState := s.hashRingState.Load()
	if hashringState == nil {
		return true
	}

	initialKey, ok := hashringState.hashRing.GetNode(job.Key.String())
	if !ok {
		return true
	}

	initialIdx := -1
	for idx, controller := range hashringState.controllers {
		if controller.Key.String() == initialKey {
			initialIdx = idx
			break
		}
	}
	if initialIdx == -1 {
		return true
	}

	if initialIdx+controllersPerJob > len(hashringState.controllers) {
		// wraps around
		return hashringState.idx >= initialIdx || hashringState.idx < (initialIdx+controllersPerJob)-len(hashringState.controllers)
	}
	return hashringState.idx >= initialIdx && hashringState.idx < initialIdx+controllersPerJob
}
