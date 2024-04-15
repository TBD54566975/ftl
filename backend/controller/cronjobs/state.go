package cronjobs

import (
	"time"

	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/alecthomas/types/optional"
)

// state models the state of the cron job service's private state for scheduling jobs and reacting to events
type state struct {
	jobs []model.CronJob

	// Used to determine if this controller is currently executing a job
	executing map[string]bool

	// Newly created jobs should be attempted by the controller that created them until other controllers
	// have a chance to resync their job lists and share responsibilities through the hash ring
	newJobs map[string]time.Time

	// We delay any job attempts in case of db errors to avoid hammering the db in a tight loop
	blockedUntil time.Time
}

func (s *state) isExecutingInCurrentController(job model.CronJob) bool {
	return s.executing[job.Key.String()]
}

func (s *state) startedExecutingJob(job model.CronJob) {
	s.executing[job.Key.String()] = true
}

func (s *state) isJobTooNewForHashRing(job model.CronJob) bool {
	if t, ok := s.newJobs[job.Key.String()]; ok {
		if time.Since(t) < newJobHashRingOverrideInterval {
			return true
		}
		delete(s.newJobs, job.Key.String())
	}
	return false
}

func (s *state) sync(jobs []model.CronJob, newDeploymentKey optional.Option[model.DeploymentKey]) {
	s.jobs = make([]model.CronJob, len(jobs))
	copy(s.jobs, jobs)
	for _, job := range s.jobs {
		if job.State != model.CronJobStateExecuting {
			delete(s.executing, job.Key.String())
		}
		if newKey, ok := newDeploymentKey.Get(); ok && job.DeploymentKey.String() == newKey.String() {
			// This job is new and should be attempted by the current controller
			s.newJobs[job.Key.String()] = time.Now()
		}
	}
}

func (s *state) updateJobs(jobs []model.CronJob) {
	updatedJobMap := jobMap(jobs)
	for idx, old := range s.jobs {
		if updated, exists := updatedJobMap[old.Key.String()]; exists {
			s.jobs[idx] = updated
			if updated.State != model.CronJobStateExecuting {
				delete(s.executing, updated.Key.String())
			}
		}
	}
}

func (s *state) removeDeploymentKey(key model.DeploymentKey) {
	s.jobs = slices.Filter(s.jobs, func(j model.CronJob) bool {
		return j.DeploymentKey.String() != key.String()
	})
}

func jobMap(jobs []model.CronJob) map[string]model.CronJob {
	m := map[string]model.CronJob{}
	for _, job := range jobs {
		m[job.Key.String()] = job
	}
	return m
}
