package observability

import (
	"context"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/golang/protobuf/jsonpb"
	metricsv1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"

	"github.com/TBD54566975/ftl/internal/3rdparty/protos/opentelemetry/proto/collector/metrics/v1/v1connect"
	"github.com/TBD54566975/ftl/internal/log"
)

type Observability struct{}

func NewService() *Observability {
	return &Observability{}
}

var _ v1connect.MetricsServiceHandler = (*Observability)(nil)

func (*Observability) Export(ctx context.Context, req *connect.Request[metricsv1.ExportMetricsServiceRequest]) (*connect.Response[metricsv1.ExportMetricsServiceResponse], error) {
	logger := log.FromContext(ctx)
	for i := range req.Msg.ResourceMetrics {
		for j := range req.Msg.ResourceMetrics[i].ScopeMetrics {
			scope := req.Msg.ResourceMetrics[i].ScopeMetrics[j].Scope
			if scope.Name == instrumentationName {
				marshaler := jsonpb.Marshaler{}
				sb := &strings.Builder{}
				err := marshaler.Marshal(sb, req.Msg.ResourceMetrics[i].ScopeMetrics[j])
				if err != nil {
					logger.Errorf(err, "Failed to marshal metrics")
					continue
				}
				logger.Tracef("FTL Metrics: %s", sb)
			}
		}
	}

	return connect.NewResponse(&metricsv1.ExportMetricsServiceResponse{}), nil
}
