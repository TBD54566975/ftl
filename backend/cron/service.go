package cron

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"time"

	"connectrpc.com/connect"
	"github.com/jpillora/backoff"
	"golang.org/x/sync/errgroup"

	"github.com/TBD54566975/ftl/backend/cron/observability"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/internal/cron"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/slices"
)

type PullSchemaClient interface {
	PullSchema(ctx context.Context, req *connect.Request[ftlv1.PullSchemaRequest]) (*connect.ServerStreamForClient[ftlv1.PullSchemaResponse], error)
}

type CallClient interface {
	Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error)
}

type cronJob struct {
	module  string
	verb    *schema.Verb
	cronmd  *schema.MetadataCronJob
	pattern cron.Pattern
	next    time.Time
}

type Config struct {
	ControllerEndpoint *url.URL `name:"ftl-endpoint" help:"Controller endpoint." env:"FTL_ENDPOINT" default:"http://127.0.0.1:8892"`
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
func Start(ctx context.Context, pullSchemaClient PullSchemaClient, verbClient CallClient) error {
	wg, ctx := errgroup.WithContext(ctx)
	changes := make(chan *ftlv1.PullSchemaResponse, 8)
	// Start processing cron jobs and schema changes.
	wg.Go(func() error {
		return run(ctx, verbClient, changes)
	})
	// Start watching for schema changes.
	wg.Go(func() error {
		rpc.RetryStreamingServerStream(ctx, "pull-schema", backoff.Backoff{}, &ftlv1.PullSchemaRequest{}, pullSchemaClient.PullSchema, func(ctx context.Context, resp *ftlv1.PullSchemaResponse) error {
			changes <- resp
			return nil
		}, rpc.AlwaysRetry())
		return nil
	})
	err := wg.Wait()
	if err != nil {
		return fmt.Errorf("cron service stopped: %w", err)
	}
	return nil
}

func run(ctx context.Context, verbClient CallClient, changes chan *ftlv1.PullSchemaResponse) error {
	logger := log.FromContext(ctx).Scope("cron")
	// Map of cron jobs for each module.
	cronJobs := map[string][]cronJob{}
	// Cron jobs ordered by next execution.
	cronQueue := []cronJob{}

	logger.Debugf("Starting cron service")

	for {
		next, ok := scheduleNext(cronQueue)
		var nextCh <-chan time.Time
		if ok {
			logger.Tracef("Next cron job scheduled in %s", next)
			nextCh = time.After(next)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("cron service stopped: %w", ctx.Err())

		case resp := <-changes:
			if err := updateCronJobs(cronJobs, resp); err != nil {
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
			if err := callCronJob(ctx, verbClient, job); err != nil {
				observability.Cron.JobFailed(ctx, cronModel)
				logger.Errorf(err, "Failed to execute cron job")
			} else {
				observability.Cron.JobSuccess(ctx, cronModel)
			}
		}
	}
}

func callCronJob(ctx context.Context, verbClient CallClient, cronJob cronJob) error {
	logger := log.FromContext(ctx).Scope("cron")
	ref := schema.Ref{Module: cronJob.module, Name: cronJob.verb.Name}
	logger.Debugf("Calling cron job %s", cronJob)
	resp, err := verbClient.Call(ctx, connect.NewRequest(&ftlv1.CallRequest{
		Verb:     ref.ToProto().(*schemapb.Ref),
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

func scheduleNext(cronQueue []cronJob) (time.Duration, bool) {
	if len(cronQueue) == 0 {
		return 0, false
	}
	return time.Until(cronQueue[0].next), true
}

func updateCronJobs(cronJobs map[string][]cronJob, resp *ftlv1.PullSchemaResponse) error {
	switch resp.ChangeType {
	case ftlv1.DeploymentChangeType_DEPLOYMENT_REMOVED:
		delete(cronJobs, resp.ModuleName)

	case ftlv1.DeploymentChangeType_DEPLOYMENT_ADDED, ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGED:
		moduleSchema, err := schema.ModuleFromProto(resp.Schema)
		if err != nil {
			return fmt.Errorf("failed to extract module schema: %w", err)
		}
		moduleJobs, err := extractCronJobs(moduleSchema)
		if err != nil {
			return fmt.Errorf("failed to extract cron jobs: %w", err)
		}
		cronJobs[resp.ModuleName] = moduleJobs
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
