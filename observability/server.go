package observability

import (
	"context"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"
	colmetricsv1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	metricsv1 "go.opentelemetry.io/proto/otlp/metrics/v1"

	"github.com/TBD54566975/ftl/common/model"
	"github.com/TBD54566975/ftl/internal/3rdparty/protos/opentelemetry/proto/collector/metrics/v1/v1connect"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

type Observability struct {
	runnerKey model.RunnerKey
	client    ftlv1connect.ObservabilityServiceClient
}

func NewService(runnerKey model.RunnerKey, client ftlv1connect.ObservabilityServiceClient) *Observability {
	return &Observability{
		runnerKey: runnerKey,
		client:    client,
	}
}

var _ v1connect.MetricsServiceHandler = (*Observability)(nil)

func (o *Observability) Export(ctx context.Context, req *connect.Request[colmetricsv1.ExportMetricsServiceRequest]) (*connect.Response[colmetricsv1.ExportMetricsServiceResponse], error) {
	ftlMetrics := []*metricsv1.ScopeMetrics{}

	for i := range req.Msg.ResourceMetrics {
		for j := range req.Msg.ResourceMetrics[i].ScopeMetrics {
			scopeMetrics := req.Msg.ResourceMetrics[i].ScopeMetrics[j]
			if scopeMetrics.Scope.Name == instrumentationName {
				ftlMetrics = append(ftlMetrics, scopeMetrics)
			}
		}
	}

	if len(ftlMetrics) > 0 {
		_, err := o.client.SendMetric(ctx, connect.NewRequest(&ftlv1.SendMetricRequest{
			RunnerKey:    o.runnerKey.String(),
			ScopeMetrics: ftlMetrics,
		}))
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return connect.NewResponse(&colmetricsv1.ExportMetricsServiceResponse{}), nil
}
