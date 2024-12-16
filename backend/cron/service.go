package cron

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"time"

	"connectrpc.com/connect"

	"github.com/block/ftl/backend/cron/observability"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/timeline"
	"github.com/block/ftl/common/cron"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/model"
	"github.com/block/ftl/internal/routing"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

type cronJob struct {
	module  string
	verb    *schema.Verb
	cronmd  *schema.MetadataCronJob
	pattern cron.Pattern
	next    time.Time
}

type Config struct {
	SchemaServiceEndpoint *url.URL `name:"ftl-endpoint" help:"Schema Service endpoint." env:"FTL_ENDPOINT" default:"http://127.0.0.1:8892"`
	TimelineEndpoint      *url.URL `help:"Timeline endpoint." env:"FTL_TIMELINE_ENDPOINT" default:"http://127.0.0.1:8894"`
}

func (c cronJob) String() string {
	desc := fmt.Sprintf("%s.%s (%s)", c.module, c.verb.Name, c.pattern)
	var next string
	if time.Until(c.next) > 0 {
		next = fmt.Sprintf(" (next run in %s)", time.Until(c.next))
	}
	return desc + next
}

// Start the cron service. Blocks until the context is cancelled.
func Start(ctx context.Context, eventSource schemaeventsource.EventSource, client routing.CallClient, timelineClient *timeline.Client) error {
	logger := log.FromContext(ctx).Scope("cron")
	ctx = log.ContextWithLogger(ctx, logger)
	// Map of cron jobs for each module.
	cronJobs := map[string][]cronJob{}
	// Cron jobs ordered by next execution.
	cronQueue := []cronJob{}

	logger.Debugf("Starting cron service")

	for {
		next, ok := scheduleNext(ctx, cronQueue, timelineClient)
		var nextCh <-chan time.Time
		if ok {
			logger.Debugf("Next cron job scheduled in %s", next)
			nextCh = time.After(next)
		} else {
			logger.Debugf("No cron jobs scheduled")
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("cron service stopped: %w", ctx.Err())

		case change := <-eventSource.Events():
			if err := updateCronJobs(ctx, cronJobs, change); err != nil {
				logger.Errorf(err, "Failed to update cron jobs")
				continue
			}
			cronQueue = rebuildQueue(cronJobs)

		// Execute scheduled cron job
		case <-nextCh:
			job := cronQueue[0]
			logger.Debugf("Executing cron job %s", job)

			nextRun, err := cron.Next(job.pattern, false)
			if err != nil {
				logger.Errorf(err, "Failed to calculate next run time")
				continue
			}
			job.next = nextRun
			cronQueue[0] = job
			orderQueue(cronQueue)

			cronModel := model.CronJob{
				// TODO: We don't have the runner key available here.
				Key:           model.NewCronJobKey(job.module, job.verb.Name),
				Verb:          schema.Ref{Module: job.module, Name: job.verb.Name},
				Schedule:      job.pattern.String(),
				StartTime:     time.Now(),
				NextExecution: job.next,
			}
			observability.Cron.JobStarted(ctx, cronModel)
			if err := callCronJob(ctx, client, job); err != nil {
				observability.Cron.JobFailed(ctx, cronModel)
				logger.Errorf(err, "Failed to execute cron job")
			} else {
				observability.Cron.JobSuccess(ctx, cronModel)
			}
		}
	}
}

func callCronJob(ctx context.Context, verbClient routing.CallClient, cronJob cronJob) error {
	logger := log.FromContext(ctx).Scope("cron")
	ref := schema.Ref{Module: cronJob.module, Name: cronJob.verb.Name}
	logger.Debugf("Calling cron job %s", cronJob)
	resp, err := verbClient.Call(ctx, connect.NewRequest(&ftlv1.CallRequest{
		Verb:     ref.ToProto(),
		Body:     []byte(`{}`),
		Metadata: &ftlv1.Metadata{},
	}))
	if err != nil {
		return fmt.Errorf("%s: call to cron job failed: %w", ref, err)
	}
	switch resp := resp.Msg.Response.(type) {
	default:
		return nil

	case *ftlv1.CallResponse_Error_:
		return fmt.Errorf("%s: cron job failed: %s", ref, resp.Error.Message)
	}
}

func scheduleNext(ctx context.Context, cronQueue []cronJob, timelineClient *timeline.Client) (time.Duration, bool) {
	if len(cronQueue) == 0 {
		return 0, false
	}
	timelineClient.Publish(ctx, timeline.CronScheduled{
		DeploymentKey: model.NewDeploymentKey(cronQueue[0].module),
		Verb:          schema.Ref{Module: cronQueue[0].module, Name: cronQueue[0].verb.Name},
		ScheduledAt:   cronQueue[0].next,
		Schedule:      cronQueue[0].pattern.String(),
	})
	return time.Until(cronQueue[0].next), true
}

func updateCronJobs(ctx context.Context, cronJobs map[string][]cronJob, change schemaeventsource.Event) error {
	logger := log.FromContext(ctx).Scope("cron")
	switch change := change.(type) {
	case schemaeventsource.EventRemove:
		// We see the new state of the module before we see the removed deployment.
		// We only want to actually remove if it was not replaced by a new deployment.
		if !change.Deleted {
			logger.Debugf("Not removing cron jobs for %s as module is still present", change.Deployment)
			return nil
		}
		logger.Debugf("Removing cron jobs for module %s", change.Module.Name)
		delete(cronJobs, change.Module.Name)

	case schemaeventsource.EventUpsert:
		logger.Debugf("Updated cron jobs for module %s", change.Module.Name)
		moduleJobs, err := extractCronJobs(change.Module)
		if err != nil {
			return fmt.Errorf("failed to extract cron jobs: %w", err)
		}
		logger.Debugf("Adding %d cron jobs for module %s", len(moduleJobs), change.Module.Name)
		cronJobs[change.Module.Name] = moduleJobs
	}
	return nil
}

func orderQueue(queue []cronJob) {
	sort.SliceStable(queue, func(i, j int) bool {
		return queue[i].next.Before(queue[j].next)
	})
}

func rebuildQueue(cronJobs map[string][]cronJob) []cronJob {
	queue := make([]cronJob, 0, len(cronJobs)*2) // Assume 2 cron jobs per module.
	for _, jobs := range cronJobs {
		queue = append(queue, jobs...)
	}
	orderQueue(queue)
	return queue
}

func extractCronJobs(module *schema.Module) ([]cronJob, error) {
	cronJobs := []cronJob{}
	for verb := range slices.FilterVariants[*schema.Verb](module.Decls) {
		cronmd, ok := slices.FindVariant[*schema.MetadataCronJob](verb.Metadata)
		if !ok {
			continue
		}
		pattern, err := cron.Parse(cronmd.Cron)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cronmd.Pos, err)
		}
		next, err := cron.Next(pattern, false)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cronmd.Pos, err)
		}
		cronJobs = append(cronJobs, cronJob{
			module:  module.Name,
			verb:    verb,
			cronmd:  cronmd,
			pattern: pattern,
			next:    next,
		})
	}
	return cronJobs, nil
}
