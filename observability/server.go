package observability

import (
	"context"
	"encoding/json"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"
	colmetricsv1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	metricsv1 "go.opentelemetry.io/proto/otlp/metrics/v1"

	"github.com/TBD54566975/ftl/internal/3rdparty/protos/opentelemetry/proto/collector/metrics/v1/v1connect"
	"github.com/TBD54566975/ftl/internal/log"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
)

type Observability struct{}

func NewService() *Observability {
	return &Observability{}
}

var _ v1connect.MetricsServiceHandler = (*Observability)(nil)

func (*Observability) Export(ctx context.Context, req *connect.Request[colmetricsv1.ExportMetricsServiceRequest]) (*connect.Response[colmetricsv1.ExportMetricsServiceResponse], error) {
	for i := range req.Msg.ResourceMetrics {
		for j := range req.Msg.ResourceMetrics[i].ScopeMetrics {
			scopeMetrics := req.Msg.ResourceMetrics[i].ScopeMetrics[j]
			if scopeMetrics.Scope.Name == instrumentationName {
				err := sendMetrics(ctx, scopeMetrics)
				if err != nil {
					return nil, errors.WithStack(err)
				}
			}
		}
	}

	return connect.NewResponse(&colmetricsv1.ExportMetricsServiceResponse{}), nil
}

func sendMetrics(ctx context.Context, metrics *metricsv1.ScopeMetrics) error {
	logger := log.FromContext(ctx)

	for i := range metrics.Metrics {
		metric := metrics.Metrics[i]

		var metricRequest *ftlv1.SendMetricRequest
		if sum := metric.GetSum(); sum != nil {
			jsonBytes, err := json.Marshal(sum.DataPoints)
			if err != nil {
				return errors.WithStack(err)
			}
			metricRequest = &ftlv1.SendMetricRequest{
				Name:       metric.Name,
				Type:       ftlv1.MetricType_COUNTER,
				Datapoints: jsonBytes,
			}
		}
		if histogram := metric.GetHistogram(); histogram != nil {
			jsonBytes, err := json.Marshal(histogram.DataPoints)
			if err != nil {
				return errors.WithStack(err)
			}
			metricRequest = &ftlv1.SendMetricRequest{
				Name:       metric.Name,
				Type:       ftlv1.MetricType_HISTOGRAM,
				Datapoints: jsonBytes,
			}
		}

		if metricRequest == nil {
			logger.Warnf("Unknown metric type for: %s", metric.Name)
			continue
		}

		logger.Tracef("Sending metric: %s", metricRequest.Name)
	}

	return nil
}
