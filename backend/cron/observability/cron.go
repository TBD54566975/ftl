package observability

import (
	"context"
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"

	"github.com/block/ftl/internal/model"
	"github.com/block/ftl/internal/observability"
)

const (
	cronMeterName            = "ftl.cron"
	cronJobFullNameAttribute = "ftl.cron.job.full_name"

	deploymentMeterName = "ftl.deployments.cron"
)

type CronMetrics struct {
	jobsActive    metric.Int64UpDownCounter
	jobsCompleted metric.Int64Counter
	jobLatency    metric.Int64Histogram
}

var Cron *CronMetrics

func init() {
	var err error
	Cron, err = initCronMetrics()
	if err != nil {
		panic(fmt.Errorf("could not initialize cron metrics: %w", err))
	}
}

func initCronMetrics() (*CronMetrics, error) {
	result := &CronMetrics{
		jobsActive:    noop.Int64UpDownCounter{},
		jobsCompleted: noop.Int64Counter{},
		jobLatency:    noop.Int64Histogram{},
	}

	var err error
	meter := otel.Meter(deploymentMeterName)

	signalName := fmt.Sprintf("%s.jobs.completed", cronMeterName)
	if result.jobsCompleted, err = meter.Int64Counter(
		signalName,
		metric.WithDescription("the number of cron jobs completed; successful or otherwise")); err != nil {
		return nil, wrapErr(signalName, err)
	}

	signalName = fmt.Sprintf("%s.jobs.active", cronMeterName)
	if result.jobsActive, err = meter.Int64UpDownCounter(
		signalName,
		metric.WithDescription("the number of actively executing cron jobs")); err != nil {
		return nil, wrapErr(signalName, err)
	}

	signalName = fmt.Sprintf("%s.job.latency", cronMeterName)
	if result.jobLatency, err = meter.Int64Histogram(
		signalName,
		metric.WithDescription("the latency between the scheduled execution time of a cron job"),
		metric.WithUnit("ms")); err != nil {
		return nil, wrapErr(signalName, err)
	}

	return result, nil
}

func (m *CronMetrics) JobStarted(ctx context.Context, job model.CronJob) {
	m.jobsActive.Add(ctx, 1, cronAttributes(job, optional.None[string]()))
}

func (m *CronMetrics) JobSuccess(ctx context.Context, job model.CronJob) {
	m.jobCompleted(ctx, job, observability.SuccessStatus)
}

func (m *CronMetrics) JobFailed(ctx context.Context, job model.CronJob) {
	m.jobCompleted(ctx, job, observability.FailureStatus)
}

func (m *CronMetrics) jobCompleted(ctx context.Context, job model.CronJob, status string) {
	elapsed := timeSinceMS(job.StartTime)

	m.jobsActive.Add(ctx, -1, cronAttributes(job, optional.None[string]()))

	completionAttributes := cronAttributes(job, optional.Some(status))
	m.jobLatency.Record(ctx, elapsed, completionAttributes)
	m.jobsCompleted.Add(ctx, 1, completionAttributes)
}

func cronAttributes(job model.CronJob, maybeStatus optional.Option[string]) metric.MeasurementOption {
	attributes := []attribute.KeyValue{
		attribute.String(observability.ModuleNameAttribute, job.Key.Payload.Module),
		attribute.String(cronJobFullNameAttribute, job.Key.String()),
		attribute.String(observability.RunnerDeploymentKeyAttribute, job.DeploymentKey.String()),
	}
	if status, ok := maybeStatus.Get(); ok {
		attributes = append(attributes, attribute.String(observability.OutcomeStatusNameAttribute, status))
	}
	return metric.WithAttributes(attributes...)
}

func wrapErr(signalName string, err error) error {
	return fmt.Errorf("failed to create %q signal: %w", signalName, err)
}

func timeSinceMS(start time.Time) int64 {
	return time.Since(start).Milliseconds()
}
