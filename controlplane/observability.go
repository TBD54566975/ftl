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
	"github.com/TBD54566975/ftl/observability"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/schema"
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

			switch data := metric.Data.(type) {
			case *metricsv1.Metric_Sum:
				err := o.insertSumMetrics(ctx, runnerKey, metric.Name, data)
				if err != nil {
					return nil, errors.WithStack(err)
				}
			case *metricsv1.Metric_Histogram:
				err := o.insertHistogramMetrics(ctx, runnerKey, metric.Name, data)
				if err != nil {
					return nil, errors.WithStack(err)
				}
			}
		}
	}

	return connect.NewResponse(&ftlv1.SendMetricResponse{}), nil
}

func (o *ObservabilityService) insertSumMetrics(ctx context.Context, runnerKey model.RunnerKey, name string, metric *metricsv1.Metric_Sum) error {
	for _, dataPoint := range metric.Sum.DataPoints {
		err := o.dal.InsertMetricEntry(ctx, dal.Metric{
			RunnerKey:  runnerKey,
			StartTime:  time.Unix(0, int64(dataPoint.StartTimeUnixNano)),
			EndTime:    time.Unix(0, int64(dataPoint.TimeUnixNano)),
			SourceVerb: getVerbRef(dataPoint.Attributes, true),
			DestVerb:   getVerbRef(dataPoint.Attributes, false),
			Name:       name,
			DataPoint:  dal.MetricCounter{Value: dataPoint.GetAsInt()},
		})
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (o *ObservabilityService) insertHistogramMetrics(ctx context.Context, runnerKey model.RunnerKey, name string, metric *metricsv1.Metric_Histogram) error {
	for _, dataPoint := range metric.Histogram.DataPoints {
		err := o.dal.InsertMetricEntry(ctx, dal.Metric{
			RunnerKey:  runnerKey,
			StartTime:  time.Unix(0, int64(dataPoint.StartTimeUnixNano)),
			EndTime:    time.Unix(0, int64(dataPoint.TimeUnixNano)),
			SourceVerb: getVerbRef(dataPoint.Attributes, true),
			DestVerb:   getVerbRef(dataPoint.Attributes, false),
			Name:       name,
			DataPoint: dal.MetricHistogram{
				Count:  int64(dataPoint.Count),
				Sum:    int64(*dataPoint.Sum),
				Bucket: dataPoint.BucketCounts,
			},
		})
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func getVerbRef(attributes []*commonv1.KeyValue, isSource bool) schema.VerbRef {
	var name string
	var module string

	for _, attr := range attributes {
		if attr.Key == observability.SourceVerbKey && isSource {
			name = attr.Value.GetStringValue()
		}
		if attr.Key == observability.SourceModuleKey && isSource {
			module = attr.Value.GetStringValue()
		}
		if attr.Key == observability.DestVerbKey && !isSource {
			name = attr.Value.GetStringValue()
		}
		if attr.Key == observability.DestModuleKey && !isSource {
			module = attr.Value.GetStringValue()
		}
	}

	return schema.VerbRef{
		Module: module,
		Name:   name,
	}
}
