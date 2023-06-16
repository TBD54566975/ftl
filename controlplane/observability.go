package controlplane

import (
	"context"

	"github.com/bufbuild/connect-go"

	"github.com/TBD54566975/ftl/controlplane/internal/dal"
	"github.com/TBD54566975/ftl/internal/log"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

type ObservabilityService struct {
	dal dal.DAL
}

var _ ftlv1connect.ObservabilityServiceHandler = (*ObservabilityService)(nil)

func NewObservabilityService(dal dal.DAL) *ObservabilityService {
	return &ObservabilityService{
		dal: dal,
	}
}

func (*ObservabilityService) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (o *ObservabilityService) SendMetric(ctx context.Context, req *connect.Request[ftlv1.SendMetricRequest]) (*connect.Response[ftlv1.SendMetricResponse], error) {
	logger := log.FromContext(ctx)
	logger.Tracef("Received metric for %s", req.Msg.RunnerKey)

	return connect.NewResponse(&ftlv1.SendMetricResponse{}), nil
}
