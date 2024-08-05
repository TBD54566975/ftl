package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/observability"
)

const (
	cronMeterName       = "ftl.cron"
	cronJobRefAttribute = "ftl.cron.job.ref"
)

type CronMetrics struct {
	jobFailures metric.Int64Counter
	jobsKilled  metric.Int64Counter
	jobsActive  metric.Int64UpDownCounter
	jobLatency  metric.Int64Histogram
}

func initCronMetrics() (*CronMetrics, error) {
	result := &CronMetrics{}

	var errs error
	var err error

	meter := otel.Meter(deploymentMeterName)

	counter := fmt.Sprintf("%s.job.failures", cronMeterName)
	if result.jobFailures, err = meter.Int64Counter(
		counter,
		metric.WithDescription("the number of failures encountered while performing activities associated with starting or ending a cron job")); err != nil {
		result.jobFailures, errs = handleInt64CounterError(counter, err, errs)
	}

	counter = fmt.Sprintf("%s.jobs.kills", cronMeterName)
	if result.jobsKilled, err = meter.Int64Counter(
		counter,
		metric.WithDescription("the number cron jobs killed by the controller")); err != nil {
		result.jobsKilled, errs = handleInt64CounterError(counter, err, errs)
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
		metric.WithDescription("the latency between the scheduled execution time and completion of a cron job"),
		metric.WithUnit("ms")); err != nil {
		result.jobLatency, errs = handleInt64HistogramCounterError(counter, err, errs)
	}

	return result, errs
}

func (m *CronMetrics) JobExecutionStarted(ctx context.Context, job model.CronJobKey, deployment model.DeploymentKey) {
	m.jobsActive.Add(ctx, 1, metric.WithAttributes(
		attribute.String(observability.ModuleNameAttribute, job.Payload.Module),
		attribute.String(cronJobRefAttribute, job.String()),
		attribute.String(observability.RunnerDeploymentKeyAttribute, deployment.String()),
	))
}

func (m *CronMetrics) JobExecutionCompleted(ctx context.Context, job model.CronJobKey, deployment model.DeploymentKey, scheduled time.Time) {
	elapsed := time.Since(scheduled)

	m.jobsActive.Add(ctx, -1, metric.WithAttributes(
		attribute.String(observability.ModuleNameAttribute, job.Payload.Module),
		attribute.String(cronJobRefAttribute, job.String()),
		attribute.String(observability.RunnerDeploymentKeyAttribute, deployment.String()),
	))

	m.jobLatency.Record(ctx, elapsed.Milliseconds(), metric.WithAttributes(
		attribute.String(observability.ModuleNameAttribute, job.Payload.Module),
		attribute.String(cronJobRefAttribute, job.String()),
		attribute.String(observability.RunnerDeploymentKeyAttribute, deployment.String()),
	))
}

func (m *CronMetrics) JobKilled(ctx context.Context, job model.CronJobKey, deployment model.DeploymentKey) {
	m.jobsActive.Add(ctx, -1, metric.WithAttributes(
		attribute.String(observability.ModuleNameAttribute, job.Payload.Module),
		attribute.String(cronJobRefAttribute, job.String()),
		attribute.String(observability.RunnerDeploymentKeyAttribute, deployment.String()),
	))
}

func (m *CronMetrics) JobFailed(ctx context.Context, job model.CronJobKey, deployment model.DeploymentKey) {
	m.jobFailures.Add(ctx, 1, metric.WithAttributes(
		attribute.String(observability.ModuleNameAttribute, job.Payload.Module),
		attribute.String(cronJobRefAttribute, job.String()),
		attribute.String(observability.RunnerDeploymentKeyAttribute, deployment.String()),
	))
}
