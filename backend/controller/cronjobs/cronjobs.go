package cronjobs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"
	"github.com/benbjohnson/clock"

	"github.com/TBD54566975/ftl/backend/controller/async"
	"github.com/TBD54566975/ftl/backend/controller/cronjobs/internal/dal"
	encryptionsvc "github.com/TBD54566975/ftl/backend/controller/encryption"
	"github.com/TBD54566975/ftl/backend/controller/encryption/api"
	"github.com/TBD54566975/ftl/backend/controller/timeline"
	"github.com/TBD54566975/ftl/internal/cron"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
)

type Service struct {
	key             model.ControllerKey
	requestSource   string
	dal             dal.DAL
	encryption      *encryptionsvc.Service
	timelineService *timeline.Service
	clock           clock.Clock
}

func New(ctx context.Context, key model.ControllerKey, requestSource string, encryption *encryptionsvc.Service, timeline *timeline.Service, conn *sql.DB) *Service {
	return NewForTesting(ctx, key, requestSource, encryption, timeline, *dal.New(conn), clock.New())
}

func NewForTesting(ctx context.Context, key model.ControllerKey, requestSource string, encryption *encryptionsvc.Service, timeline *timeline.Service, dal dal.DAL, clock clock.Clock) *Service {
	svc := &Service{
		key:             key,
		requestSource:   requestSource,
		dal:             dal,
		encryption:      encryption,
		timelineService: timeline,
		clock:           clock,
	}
	return svc
}

func (s *Service) NewCronJobsForModule(ctx context.Context, module *schema.Module) ([]model.CronJob, error) {
	logger := log.FromContext(ctx).Scope("cron").Module(module.Name)
	start := s.clock.Now().UTC()
	newJobs := []model.CronJob{}
	merr := []error{}
	for _, decl := range module.Decls {
		verb, ok := decl.(*schema.Verb)
		if !ok {
			continue
		}
		for _, metadata := range verb.Metadata {
			cronMetadata, ok := metadata.(*schema.MetadataCronJob)
			if !ok {
				continue
			}
			cronStr := cronMetadata.Cron
			schedule, err := cron.Parse(cronStr)
			if err != nil {
				merr = append(merr, fmt.Errorf("failed to parse cron schedule %q: %w", cronStr, err))
				continue
			}
			next, err := cron.NextAfter(schedule, start, false)
			if err != nil {
				merr = append(merr, fmt.Errorf("failed to calculate next execution for cron job %v:%v with schedule %q: %w", module.Name, verb.Name, schedule, err))
				continue
			}
			newJobs = append(newJobs, model.CronJob{
				Key:           model.NewCronJobKey(module.Name, verb.Name),
				Verb:          schema.Ref{Module: module.Name, Name: verb.Name},
				Schedule:      cronStr,
				StartTime:     start,
				NextExecution: next,
				// DeploymentKey: Filled in by DAL
			})
		}
	}
	logger.Tracef("Found %d cron jobs", len(newJobs))
	if len(merr) > 0 {
		return nil, errors.Join(merr...)
	}
	return newJobs, nil
}

// CreatedOrReplacedDeloyment is called by the responsible controller to its cron service, we can
// schedule all cron jobs here since the cron_jobs rows are locked within the transaction and the
// controllers won't step on each other.
func (s *Service) CreatedOrReplacedDeloyment(ctx context.Context) error {
	logger := log.FromContext(ctx).Scope("cron")
	logger.Tracef("New deployment; scheduling cron jobs")
	err := s.ScheduleCronJobs(ctx)
	if err != nil {
		return fmt.Errorf("failed to schedule cron jobs: %w", err)
	}
	return nil
}

// scheduleCronJobs schedules all cron jobs that are not already scheduled.
func (s *Service) ScheduleCronJobs(ctx context.Context) (err error) {
	logger := log.FromContext(ctx).Scope("cron")
	now := s.clock.Now().UTC()

	tx, err := s.dal.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.CommitOrRollback(ctx, &err)

	jobs, err := tx.GetUnscheduledCronJobs(ctx, now)
	if err != nil {
		return fmt.Errorf("failed to get unscheduled cron jobs: %w", err)
	}
	logger.Tracef("Scheduling %d cron jobs", len(jobs))
	for _, job := range jobs {
		err = s.scheduleCronJob(ctx, tx, job, now)
		if err != nil {
			s.timelineService.EnqueueEvent(ctx, &timeline.CronScheduled{
				DeploymentKey: job.DeploymentKey,
				Verb:          job.Verb,
				Time:          now,
				ScheduledAt:   job.NextExecution,
				Schedule:      job.Schedule,
				Error:         optional.Some(err.Error()),
			})
			return fmt.Errorf("failed to schedule cron job %q: %w", job.Key, err)
		}
	}

	return nil
}

// OnJobCompletion is called by the controller when a cron job async call completes. We schedule
// the next execution of the cron job here.
func (s *Service) OnJobCompletion(ctx context.Context, key model.CronJobKey, failed bool) (err error) {
	logger := log.FromContext(ctx).Scope("cron")
	logger.Tracef("Cron job %q completed with failed=%v", key, failed)
	now := s.clock.Now().UTC()

	tx, err := s.dal.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.CommitOrRollback(ctx, &err)

	job, err := tx.GetCronJobByKey(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to get cron job %q: %w", key, err)
	}
	err = s.scheduleCronJob(ctx, tx, job, now)
	if err != nil {
		s.timelineService.EnqueueEvent(ctx, &timeline.CronScheduled{
			DeploymentKey: job.DeploymentKey,
			Verb:          job.Verb,
			Time:          now,
			ScheduledAt:   job.NextExecution,
			Schedule:      job.Schedule,
			Error:         optional.Some(err.Error()),
		})
		return fmt.Errorf("failed to schedule cron job %q: %w", key, err)
	}
	return nil
}

// scheduleCronJob schedules the next execution of a single cron job.
func (s *Service) scheduleCronJob(ctx context.Context, tx *dal.DAL, job model.CronJob, startTime time.Time) error {
	logger := log.FromContext(ctx).Scope("cron").Module(job.Verb.Module)
	now := s.clock.Now().UTC()
	pending, err := tx.IsCronJobPending(ctx, job.Key, now)
	if err != nil {
		return fmt.Errorf("failed to check if cron job %q is pending: %w", job.Key, err)
	}
	if pending {
		logger.Tracef("Attempt to schedule cron job %q which is already pending", job.Key)
		return nil
	}

	pattern, err := cron.Parse(job.Schedule)
	if err != nil {
		return fmt.Errorf("failed to parse cron schedule %q: %w", job.Schedule, err)
	}
	originTime := job.StartTime
	if t, ok := job.LastExecution.Get(); ok {
		originTime = t
	}
	nextAttemptForJob, err := cron.NextAfter(pattern, originTime, false)
	if err != nil {
		return fmt.Errorf("failed to calculate next execution for cron job %q with schedule %q: %w", job.Key, job.Schedule, err)
	}
	if nextAttemptForJob.Before(now) {
		nextAttemptForJob = now
	}

	logger.Tracef("Scheduling cron job %q async_call execution at %s", job.Key, nextAttemptForJob)
	origin := &async.AsyncOriginCron{CronJobKey: job.Key}
	var request api.EncryptedColumn[api.AsyncSubKey]
	err = s.encryption.Encrypt([]byte(`{}`), &request)
	if err != nil {
		return fmt.Errorf("failed to encrypt request for job %q: %w", job.Key, err)
	}
	id, err := tx.CreateAsyncCall(ctx, dal.CreateAsyncCallParams{
		ScheduledAt: nextAttemptForJob,
		Verb:        schema.RefKey{Module: job.Verb.Module, Name: job.Verb.Name},
		Origin:      origin.String(),
		Request:     request,
	})
	if err != nil {
		return fmt.Errorf("failed to create async call for job %q: %w", job.Key, err)
	}
	futureAttemptForJob, err := cron.NextAfter(pattern, nextAttemptForJob, false)
	if err != nil {
		return fmt.Errorf("failed to calculate future execution for cron job %q with schedule %q: %w", job.Key, job.Schedule, err)
	}
	logger.Tracef("Updating cron job %q with last attempt at %s and next attempt at %s", job.Key, nextAttemptForJob, futureAttemptForJob)
	err = tx.UpdateCronJobExecution(ctx, dal.UpdateCronJobExecutionParams{
		LastAsyncCallID: id,
		LastExecution:   nextAttemptForJob,
		NextExecution:   futureAttemptForJob,
		Key:             job.Key,
	})
	if err != nil {
		return fmt.Errorf("failed to update cron job %q: %w", job.Key, err)
	}
	s.timelineService.EnqueueEvent(ctx, &timeline.CronScheduled{
		DeploymentKey: job.DeploymentKey,
		Verb:          job.Verb,
		Time:          startTime,
		ScheduledAt:   nextAttemptForJob,
		Schedule:      job.Schedule,
	})
	return nil
}

func (s *Service) DeleteCronJobsForDeployment(ctx context.Context, key model.DeploymentKey) error {
	logger := log.FromContext(ctx).Scope("cron")
	logger.Tracef("Deleting cron jobs for deployment %q", key)
	err := s.dal.DeleteCronJobsForDeployment(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to remove cron jobs for deployment %q: %w", key, err)
	}
	return nil
}
