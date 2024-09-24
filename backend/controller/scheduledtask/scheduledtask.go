// Package scheduledtask implements a task scheduler.
package scheduledtask

import (
	"context"
	"errors"
	"math/rand"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/alecthomas/types/optional"
	"github.com/benbjohnson/clock"
	"github.com/jpillora/backoff"

	dalmodel "github.com/TBD54566975/ftl/backend/controller/dal/model"
	"github.com/TBD54566975/ftl/backend/controller/leases"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/slices"
)

type descriptor struct {
	next        time.Time
	name        string
	retry       backoff.Backoff
	job         Job
	singlyHomed bool
	lease       leases.Lease
}

// A Job is a function that is scheduled to run periodically.
//
// The Job itself controls its schedule by returning the next time it should
// run.
type Job func(ctx context.Context) (time.Duration, error)

type DALFunc func(ctx context.Context, all bool) ([]dalmodel.Controller, error)

// Scheduler is a task scheduler for the controller.
//
// Each job runs in its own goroutine.
//
// The scheduler uses a consistent hash ring to attempt to ensure that jobs are
// only run on a single controller at a time. This is not guaranteed, however,
// as the hash ring is only updated periodically and controllers may have
// inconsistent views of the hash ring.
type Scheduler struct {
	key    model.ControllerKey
	jobs   chan *descriptor
	clock  clock.Clock
	leaser leases.Leaser
}

// New creates a new [Scheduler].
func New(ctx context.Context, id model.ControllerKey, leaser leases.Leaser) *Scheduler {
	return NewForTesting(ctx, id, clock.New(), leaser)
}

func NewForTesting(ctx context.Context, id model.ControllerKey, clock clock.Clock, leaser leases.Leaser) *Scheduler {
	s := &Scheduler{
		key:    id,
		jobs:   make(chan *descriptor),
		clock:  clock,
		leaser: leaser,
	}
	go s.run(ctx)
	return s
}

// Singleton schedules a job to attempt to run on only a single controller.
//
// This is not guaranteed, however, as controllers may have inconsistent views
// of the hash ring.
func (s *Scheduler) Singleton(retry backoff.Backoff, job Job) {
	s.schedule(retry, job, true)
}

// Parallel schedules a job to run on every controller.
func (s *Scheduler) Parallel(retry backoff.Backoff, job Job) {
	s.schedule(retry, job, false)
}

func (s *Scheduler) schedule(retry backoff.Backoff, job Job, singlyHomed bool) {
	name := runtime.FuncForPC(reflect.ValueOf(job).Pointer()).Name()
	name = name[strings.LastIndex(name, ".")+1:]
	name = strings.TrimSuffix(name, "-fm")
	s.jobs <- &descriptor{
		name:        name,
		retry:       retry,
		job:         job,
		singlyHomed: singlyHomed,
		next:        s.clock.Now().Add(time.Millisecond * time.Duration(rand.Int63n(2000))), //nolint:gosec
	}
}

func (s *Scheduler) run(ctx context.Context) {
	logger := log.FromContext(ctx).Scope("scheduler")
	// List of jobs to run.
	// For singleton jobs running on a different host, this can include jobs
	// scheduled in the past. These are skipped on each run.
	jobs := []*descriptor{}
	for {
		next := s.clock.Now().Add(time.Second)
		// Find the next job to run.
		if len(jobs) > 0 {
			sort.Slice(jobs, func(i, j int) bool { return jobs[i].next.Before(jobs[j].next) })
			for _, job := range jobs {
				if job.next.IsZero() {
					continue
				}
				next = job.next
				break
			}
		}

		now := s.clock.Now()
		select {
		case <-ctx.Done():
			return

		case <-s.clock.After(next.Sub(now)):
			// Jobs to reschedule on the next run.
			for i, job := range jobs {
				if job.next.After(s.clock.Now()) {
					continue
				}
				// If the job is singly homed, see if we can acquire the lease.
				if job.singlyHomed && job.lease == nil {
					key := leases.SystemKey("scheduledtask", job.name)
					lease, _, err := s.leaser.AcquireLease(ctx, key, time.Second*10, optional.None[any]())
					if err != nil {
						if errors.Is(err, leases.ErrConflict) {
							logger.Scope(job.name).Tracef("Scheduled task lease is held by another controller, will try again shortly: %v", key)
						} else {
							logger.Scope(job.name).Debugf("Failed to acquire scheduled task lease %v: %s", key, err)
						}
						job.next = s.clock.Now().Add(job.retry.Duration())
						continue
					}
					job.lease = lease
				}
				jobs[i] = nil // Zero out scheduled jobs.
				logger.Scope(job.name).Tracef("Running scheduled task")
				go func() {
					if delay, err := job.job(ctx); err != nil {
						logger.Scope(job.name).Warnf("%s", err)
						job.next = s.clock.Now().Add(job.retry.Duration())
					} else {
						// Reschedule the job.
						job.retry.Reset()
						job.next = s.clock.Now().Add(delay)
					}
					s.jobs <- job
				}()
			}
			jobs = slices.Filter(jobs, func(job *descriptor) bool { return job != nil })

		case job := <-s.jobs:
			jobs = append(jobs, job)
		}
	}
}
