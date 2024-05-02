package cronjobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/atomic"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/pubsub"
	"github.com/benbjohnson/clock"
	"github.com/jpillora/backoff"
	"github.com/serialx/hashring"

	"github.com/TBD54566975/ftl/backend/controller/dal"
	"github.com/TBD54566975/ftl/backend/controller/scheduledtask"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/cron"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/slices"
)

const (
	controllersPerJob              = 2
	jobResetInterval               = time.Minute
	newJobHashRingOverrideInterval = time.Minute + time.Second*20
)

type Config struct {
	Timeout time.Duration
}

//sumtype:decl
type event interface {
	// cronJobEvent is a marker to ensure that all events implement the interface.
	cronJobEvent()
}

type syncEvent struct {
	jobs               []model.CronJob
	addedDeploymentKey optional.Option[model.DeploymentKey]
}

func (syncEvent) cronJobEvent() {}

type endedJobsEvent struct {
	jobs []model.CronJob
}

func (endedJobsEvent) cronJobEvent() {}

type updatedHashRingEvent struct{}

func (updatedHashRingEvent) cronJobEvent() {}

type hashRingState struct {
	hashRing    *hashring.HashRing
	controllers []dal.Controller
	idx         int
}

type DAL interface {
	GetCronJobs(ctx context.Context) ([]model.CronJob, error)
	StartCronJobs(ctx context.Context, jobs []model.CronJob) (attemptedJobs []dal.AttemptedCronJob, err error)
	EndCronJob(ctx context.Context, job model.CronJob, next time.Time) (model.CronJob, error)
	GetStaleCronJobs(ctx context.Context, duration time.Duration) ([]model.CronJob, error)
}

type Scheduler interface {
	Singleton(retry backoff.Backoff, job scheduledtask.Job)
	Parallel(retry backoff.Backoff, job scheduledtask.Job)
}

type ExecuteCallFunc func(context.Context, *connect.Request[ftlv1.CallRequest], optional.Option[model.RequestKey], string) (*connect.Response[ftlv1.CallResponse], error)

type Service struct {
	config        Config
	key           model.ControllerKey
	requestSource string

	dal       DAL
	scheduler Scheduler
	call      ExecuteCallFunc

	clock  clock.Clock
	events *pubsub.Topic[event]

	hashRingState atomic.Value[*hashRingState]
}

func New(ctx context.Context, key model.ControllerKey, requestSource string, config Config, dal DAL, scheduler Scheduler, call ExecuteCallFunc) *Service {
	return NewForTesting(ctx, key, requestSource, config, dal, scheduler, call, clock.New())
}

func NewForTesting(ctx context.Context, key model.ControllerKey, requestSource string, config Config, dal DAL, scheduler Scheduler, call ExecuteCallFunc, clock clock.Clock) *Service {
	svc := &Service{
		config:        config,
		key:           key,
		requestSource: requestSource,
		dal:           dal,
		scheduler:     scheduler,
		call:          call,
		clock:         clock,
		events:        pubsub.New[event](),
	}
	svc.UpdatedControllerList(ctx, nil)

	svc.scheduler.Parallel(backoff.Backoff{Min: time.Second, Max: jobResetInterval}, svc.syncJobs)
	svc.scheduler.Singleton(backoff.Backoff{Min: time.Second, Max: time.Minute}, svc.killOldJobs)

	go svc.watchForUpdates(ctx)

	return svc
}

func (s *Service) NewCronJobsForModule(ctx context.Context, module *schemapb.Module) ([]model.CronJob, error) {
	start := s.clock.Now().UTC()
	newJobs := []model.CronJob{}
	merr := []error{}
	for _, decl := range module.Decls {
		verb, ok := decl.Value.(*schemapb.Decl_Verb)
		if !ok {
			continue
		}
		for _, metadata := range verb.Verb.Metadata {
			cronMetadata, ok := metadata.Value.(*schemapb.Metadata_CronJob)
			if !ok {
				continue
			}
			cronStr := cronMetadata.CronJob.Cron
			schedule, err := cron.Parse(cronStr)
			if err != nil {
				merr = append(merr, fmt.Errorf("failed to parse cron schedule %q: %w", cronStr, err))
				continue
			}
			next, err := cron.NextAfter(schedule, start, false)
			if err != nil {
				merr = append(merr, fmt.Errorf("failed to calculate next execution for cron job %v:%v with schedule %q: %w", module.Name, verb.Verb.Name, schedule, err))
				continue
			}
			newJobs = append(newJobs, model.CronJob{
				Key:           model.NewCronJobKey(module.Name, verb.Verb.Name),
				Verb:          schema.Ref{Module: module.Name, Name: verb.Verb.Name},
				Schedule:      cronStr,
				StartTime:     start,
				NextExecution: next,
				State:         model.CronJobStateIdle,
				// DeploymentKey: Filled in by DAL
			})
		}
	}
	if len(merr) > 0 {
		return nil, errors.Join(merr...)
	}
	return newJobs, nil
}

// CreatedOrReplacedDeloyment is only called by the responsible controller to its cron service, and will not be received by the other cron services.
// When a controller creates/replaces a deployment, its cron job service is responsible for
// the newly created cron jobs until other controllers have a chance to resync their list of jobs and start sharing responsibility of the new cron jobs.
func (s *Service) CreatedOrReplacedDeloyment(ctx context.Context, newDeploymentKey model.DeploymentKey) {
	// Rather than finding old/new cron jobs and updating our state, we can just resync the list of jobs
	_ = s.syncJobsWithNewDeploymentKey(ctx, optional.Some(newDeploymentKey))
}

// SyncJobs is run periodically via a scheduled task
func (s *Service) syncJobs(ctx context.Context) (time.Duration, error) {
	err := s.syncJobsWithNewDeploymentKey(ctx, optional.None[model.DeploymentKey]())
	if err != nil {
		return 0, err
	}
	return jobResetInterval, nil
}

// syncJobsWithNewDeploymentKey resyncs the list of jobs and marks the deployment key as added so that it can overrule the hash ring for a short time.
func (s *Service) syncJobsWithNewDeploymentKey(ctx context.Context, deploymentKey optional.Option[model.DeploymentKey]) error {
	logger := log.FromContext(ctx)

	jobs, err := s.dal.GetCronJobs(ctx)
	if err != nil {
		logger.Errorf(err, "failed to get cron jobs")
		return fmt.Errorf("failed to get cron jobs: %w", err)
	}
	s.events.Publish(syncEvent{
		jobs:               jobs,
		addedDeploymentKey: deploymentKey,
	})
	return nil
}

func (s *Service) executeJob(ctx context.Context, job model.CronJob) {
	logger := log.FromContext(ctx)
	requestBody := map[string]any{}
	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		logger.Errorf(err, "could not build body for cron job: %v", job.Key)
		return
	}

	req := connect.NewRequest(&ftlv1.CallRequest{
		Verb: &schemapb.Ref{Module: job.Verb.Module, Name: job.Verb.Name},
		Body: requestJSON,
	})

	requestKey := model.NewRequestKey(model.OriginCron, fmt.Sprintf("%s-%s", job.Verb.Module, job.Verb.Name))

	callCtx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()
	_, err = s.call(callCtx, req, optional.Some(requestKey), s.requestSource)
	if err != nil {
		logger.Errorf(err, "failed to execute cron job %v", job.Key)
		// Do not return, continue to end the job and schedule the next execution
	}

	schedule, err := cron.Parse(job.Schedule)
	if err != nil {
		logger.Errorf(err, "failed to parse cron schedule %q", job.Schedule)
		return
	}
	next, err := cron.NextAfter(schedule, s.clock.Now().UTC(), false)
	if err != nil {
		logger.Errorf(err, "failed to calculate next execution for cron job %v with schedule %q", job.Key, job.Schedule)
		return
	}

	updatedJob, err := s.dal.EndCronJob(ctx, job, next)
	if err != nil {
		logger.Errorf(err, "failed to end cron job %v", job.Key)
	} else {
		s.events.Publish(endedJobsEvent{
			jobs: []model.CronJob{updatedJob},
		})
	}
}

// killOldJobs looks for jobs that have been executing for too long.
// A soft timeout should normally occur from the job's context timing out, but there are cases where this does not happen (eg: unresponsive or dead controller)
// In these cases we need a hard timout after an additional grace period.
// To do this, this function resets these job's state to idle and updates the next execution time in the db so the job can be picked up again next time.
func (s *Service) killOldJobs(ctx context.Context) (time.Duration, error) {
	logger := log.FromContext(ctx)
	staleJobs, err := s.dal.GetStaleCronJobs(ctx, s.config.Timeout+time.Minute)
	if err != nil {
		return 0, err
	} else if len(staleJobs) == 0 {
		return time.Minute, nil
	}

	updatedJobs := []model.CronJob{}
	for _, stale := range staleJobs {
		start := s.clock.Now().UTC()
		pattern, err := cron.Parse(stale.Schedule)
		if err != nil {
			logger.Errorf(err, "Could not kill stale cron job %q because schedule could not be parsed: %q", stale.Key, stale.Schedule)
			continue
		}
		next, err := cron.NextAfter(pattern, start, false)
		if err != nil {
			logger.Errorf(err, "Could not kill stale cron job %q because next date could not be calculated: %q", stale.Key, stale.Schedule)
			continue
		}

		updated, err := s.dal.EndCronJob(ctx, stale, next)
		if err != nil {
			logger.Errorf(err, "Could not kill stale cron job %s because: %v", stale.Key, err)
			continue
		}
		logger.Warnf("Killed stale cron job %s", stale.Key)
		updatedJobs = append(updatedJobs, updated)
	}

	s.events.Publish(endedJobsEvent{
		jobs: updatedJobs,
	})

	return time.Minute, nil
}

// watchForUpdates is the centralized place that handles:
// - the list of known jobs and their state
// - executing jobs when they are due
// - reacting to events that change the list of jobs, deployments or hash ring
//
// State is private to this function to ensure thread safety.
func (s *Service) watchForUpdates(ctx context.Context) {
	logger := log.FromContext(ctx).Scope("cron")

	events := make(chan event, 128)
	s.events.Subscribe(events)
	defer s.events.Unsubscribe(events)

	state := &state{
		executing:    map[string]bool{},
		newJobs:      map[string]time.Time{},
		blockedUntil: s.clock.Now(),
	}

	for {
		now := s.clock.Now()
		next := now.Add(time.Hour) // should never be reached, expect a different signal long beforehand
		for _, j := range state.jobs {
			if possibleNext, err := s.nextAttemptForJob(j, state, false); err == nil && possibleNext.Before(next) {
				next = possibleNext
			}
		}

		if next.Before(state.blockedUntil) {
			next = state.blockedUntil
			logger.Tracef("loop blocked for %v", next.Sub(now))
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
			jobsToAttempt := slices.Filter(state.jobs, func(j model.CronJob) bool {
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
			updatedJobs := []model.CronJob{}
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
						logger.Errorf(err, "failed to end cron job %s", job.Key.String())
					}
					continue
				}
				logger.Infof("executing job %v", job.Key)
				state.startedExecutingJob(job.CronJob)
				go s.executeJob(ctx, job.CronJob)
			}

			// Update job list
			state.updateJobs(updatedJobs)
			for _, key := range removedDeploymentKeys {
				state.removeDeploymentKey(key)
			}
		case e := <-events:
			switch event := e.(type) {
			case syncEvent:
				logger.Tracef("syncing job list: %d jobs", len(event.jobs))
				state.sync(event.jobs, event.addedDeploymentKey)
			case endedJobsEvent:
				logger.Tracef("updating %d jobs", len(event.jobs))
				state.updateJobs(event.jobs)
			case updatedHashRingEvent:
				// do another cycle through the loop to see if new jobs need to be scheduled
			}
		}
	}
}

func (s *Service) nextAttemptForJob(job model.CronJob, state *state, allowsNow bool) (time.Time, error) {
	if !s.isResponsibleForJob(job, state) {
		return s.clock.Now(), fmt.Errorf("controller is not responsible for job")
	}
	if job.State == model.CronJobStateExecuting {
		if state.isExecutingInCurrentController(job) {
			// no need to schedule this job until it finishes
			return s.clock.Now(), fmt.Errorf("controller is already waiting for job to finish")
		}
		// We don't know when the other controller that is executing this job will finish it
		// So we should optimistically attempt it when the next execution date is due assuming the job finishes
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

	s.events.Publish(updatedHashRingEvent{})
}

// isResponsibleForJob indicates whether a this service should be responsible for attempting jobs,
// or if enough other controllers will handle it. This allows us to spread the job load across controllers.
func (s *Service) isResponsibleForJob(job model.CronJob, state *state) bool {
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
