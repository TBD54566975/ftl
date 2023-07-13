package controlplane

import (
	"context"
	"time"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	metricsv1 "go.opentelemetry.io/proto/otlp/metrics/v1"

	"github.com/TBD54566975/ftl/common/model"
	"github.com/TBD54566975/ftl/controlplane/internal/dal"
	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/TBD54566975/ftl/observability"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

type ObservabilityService struct {
	dal *dal.DAL
}

var _ ftlv1connect.ObservabilityServiceHandler = (*ObservabilityService)(nil)

func NewObservabilityService(dal *dal.DAL) *ObservabilityService {
	return &ObservabilityService{
		dal: dal,
	}
}

func (*ObservabilityService) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (o *ObservabilityService) SendMetric(ctx context.Context, req *connect.Request[ftlv1.SendMetricRequest]) (*connect.Response[ftlv1.SendMetricResponse], error) {
	runnerKey, err := model.ParseRunnerKey(req.Msg.RunnerKey)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid runner key"))
	}

	for _, scopeMetric := range req.Msg.ScopeMetrics {
		for _, metric := range scopeMetric.Metrics {
			var dalMetrics []dal.Metric
			switch data := metric.Data.(type) {
			case *metricsv1.Metric_Sum:
				dalMetrics = o.extractSumMetrics(runnerKey, metric.Name, data)
			case *metricsv1.Metric_Histogram:
				dalMetrics = o.extractHistogramMetrics(runnerKey, metric.Name, data)
			}
			for _, metric := range dalMetrics {
				err := o.dal.InsertMetricEntry(ctx, metric)
				if err != nil {
					return nil, connect.NewError(connect.CodeInternal, errors.Wrap(err, "failed to insert metrics"))
				}
			}
		}
	}

	return connect.NewResponse(&ftlv1.SendMetricResponse{}), nil
}

func (o *ObservabilityService) extractSumMetrics(runnerKey model.RunnerKey, name string, metric *metricsv1.Metric_Sum) []dal.Metric {
	return slices.Map(metric.Sum.DataPoints, func(dataPoint *metricsv1.NumberDataPoint) dal.Metric {
		return newMetric(runnerKey, name, dataPoint, dal.MetricCounter{
			Value: dataPoint.GetAsInt(),
		})
	})
}

func (o *ObservabilityService) extractHistogramMetrics(runnerKey model.RunnerKey, name string, metric *metricsv1.Metric_Histogram) []dal.Metric {
	return slices.Map(metric.Histogram.DataPoints, func(dataPoint *metricsv1.HistogramDataPoint) dal.Metric {
		return newMetric(runnerKey, name, dataPoint, dal.MetricHistogram{
			Count:  int64(dataPoint.Count),
			Sum:    int64(*dataPoint.Sum),
			Bucket: dataPoint.BucketCounts,
		})
	})
}

func fillMetricFromAttributes(metric *dal.Metric, attributes []*commonv1.KeyValue) {
	for _, attr := range attributes {
		switch attr.Key {
		case observability.SourceVerbKey:
			metric.SourceVerb.Name = attr.Value.GetStringValue()
		case observability.SourceModuleKey:
			metric.SourceVerb.Module = attr.Value.GetStringValue()
		case observability.DestVerbKey:
			metric.DestVerb.Name = attr.Value.GetStringValue()
		case observability.DestModuleKey:
			metric.DestVerb.Module = attr.Value.GetStringValue()
		case observability.RequestIDKey:
			metric.RequestID = attr.Value.GetIntValue()
		}
	}
}

// Interface common to Sum and Histogram
type metricInterface interface {
	GetStartTimeUnixNano() uint64
	GetTimeUnixNano() uint64
	GetAttributes() []*commonv1.KeyValue
}

func newMetric(runnerKey model.RunnerKey, name string, data metricInterface, dataPoint dal.DataPoint) dal.Metric {
	metric := dal.Metric{
		RunnerKey: runnerKey,
		Name:      name,
		StartTime: time.Unix(0, int64(data.GetStartTimeUnixNano())),
		EndTime:   time.Unix(0, int64(data.GetTimeUnixNano())),
		DataPoint: dataPoint,
	}
	fillMetricFromAttributes(&metric, data.GetAttributes())
	return metric
}
