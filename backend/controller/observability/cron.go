package observability

import (
	"context"
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/observability"
)

const (
	cronMeterName            = "ftl.cron"
	cronJobFullNameAttribute = "ftl.cron.job.full_name"

	cronJobKilledStatus      = "killed"
	cronJobFailedStartStatus = "failed_start"
)

type CronMetrics struct {
	jobsActive    metric.Int64UpDownCounter
	jobsCompleted metric.Int64Counter
	jobLatency    metric.Int64Histogram
}

func initCronMetrics() (*CronMetrics, error) {
	result := &CronMetrics{}

	var errs error
	var err error

	meter := otel.Meter(deploymentMeterName)

	counter := fmt.Sprintf("%s.job.failures", cronMeterName)
	if result.jobsCompleted, err = meter.Int64Counter(
		counter,
		metric.WithDescription("the number of cron jobs completed; successful or otherwise")); err != nil {
		result.jobsCompleted, errs = handleInt64CounterError(counter, err, errs)
	}

	counter = fmt.Sprintf("%s.jobs.active", cronMeterName)
	if result.jobsActive, err = meter.Int64UpDownCounter(
		counter,
		metric.WithDescription("the number of actively executing cron jobs")); err != nil {
		result.jobsActive, errs = handleInt64UpDownCounterError(counter, err, errs)
	}

	counter = fmt.Sprintf("%s.job.latency", cronMeterName)
	if result.jobLatency, err = meter.Int64Histogram(
		counter,
		metric.WithDescription("the latency between the scheduled execution time of a cron job"),
		metric.WithUnit("ms")); err != nil {
		result.jobLatency, errs = handleInt64HistogramCounterError(counter, err, errs)
	}

	return result, errs
}

func (m *CronMetrics) JobStarted(ctx context.Context, job model.CronJob) {
	m.jobsActive.Add(ctx, 1, cronAttributes(job, optional.None[string]()))
}

func (m *CronMetrics) JobSuccess(ctx context.Context, job model.CronJob) {
	m.jobCompleted(ctx, job, observability.SuccessStatus)
}

func (m *CronMetrics) JobKilled(ctx context.Context, job model.CronJob) {
	m.jobCompleted(ctx, job, cronJobKilledStatus)
}

func (m *CronMetrics) JobFailedStart(ctx context.Context, job model.CronJob) {
	completionAttributes := cronAttributes(job, optional.Some(cronJobFailedStartStatus))

	elapsed := time.Since(job.NextExecution)
	m.jobLatency.Record(ctx, elapsed.Milliseconds(), completionAttributes)
	m.jobsCompleted.Add(ctx, 1, completionAttributes)
}

func (m *CronMetrics) JobFailed(ctx context.Context, job model.CronJob) {
	m.jobCompleted(ctx, job, observability.FailureStatus)
}

func (m *CronMetrics) jobCompleted(ctx context.Context, job model.CronJob, status string) {
	elapsed := time.Since(job.NextExecution)

	m.jobsActive.Add(ctx, -1, cronAttributes(job, optional.None[string]()))

	completionAttributes := cronAttributes(job, optional.Some(status))
	m.jobLatency.Record(ctx, elapsed.Milliseconds(), completionAttributes)
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
