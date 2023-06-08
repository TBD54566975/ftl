package observability

import (
	"context"
	"net/url"
	"time"

	"github.com/bufbuild/connect-go"
	metricsv1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"

	"github.com/TBD54566975/ftl/internal/3rdparty/protos/opentelemetry/proto/collector/metrics/v1/v1connect"
	"github.com/TBD54566975/ftl/internal/log"
)

type Config struct {
	ObservabilityEndpoint *url.URL      `help:"FTL observability endpoint." env:"FTL_OBSERVABILITY_ENDPOINT" required:""`
	Interval              time.Duration `default:"30s" help:"Interval to export metrics." env:"FTL_METRICS_INTERVAL"`
}

type Observability struct{}

func NewService() *Observability {
	return &Observability{}
}

var _ v1connect.MetricsServiceHandler = (*Observability)(nil)

// Export implements v1connect.MetricsServiceHandler
func (*Observability) Export(ctx context.Context, req *connect.Request[metricsv1.ExportMetricsServiceRequest]) (*connect.Response[metricsv1.ExportMetricsServiceResponse], error) {
	logger := log.FromContext(ctx)
	logger.Infof("Metric Req %s:", req.Msg)
	return connect.NewResponse(&metricsv1.ExportMetricsServiceResponse{}), nil
}
