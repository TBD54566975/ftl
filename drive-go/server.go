package drivego

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"github.com/TBD54566975/ftl/common/log"
	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/common/rpc"
	"github.com/TBD54566975/ftl/observability"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	sdkgo "github.com/TBD54566975/ftl/sdk-go"
)

type UserVerbConfig struct {
	FTLEndpoint              *url.URL `help:"FTL endpoint." env:"FTL_ENDPOINT" required:""`
	FTLObservabilityEndpoint *url.URL `help:"FTL observability endpoint." env:"FTL_OBSERVABILITY_ENDPOINT" required:""`
}

// NewUserVerbServer starts a new code-generated drive for user Verbs.
//
// This function is intended to be used by the code generator.
func NewUserVerbServer(handlers ...Handler) plugin.Constructor[ftlv1connect.VerbServiceHandler, UserVerbConfig] {
	return func(ctx context.Context, uc UserVerbConfig) (context.Context, ftlv1connect.VerbServiceHandler, error) {
		logger := log.FromContext(ctx)

		//TODO(wb): Are the 2 endpoints below supposed to match? Because they do :)
		logger.Infof("FTLEndpoint: %s", uc.FTLEndpoint)
		logger.Infof("FTLObservabilityEndpoint: %s", uc.FTLObservabilityEndpoint)

		verbServiceClient := rpc.Dial(ftlv1connect.NewVerbServiceClient, uc.FTLEndpoint.String(), log.Error)
		ctx = rpc.ContextWithClient(ctx, verbServiceClient)
		observabilityServiceClient := rpc.Dial(ftlv1connect.NewObservabilityServiceClient, uc.FTLObservabilityEndpoint.String(), log.Error)
		ctx = rpc.ContextWithClient(ctx, observabilityServiceClient)

		// TODO(wb): Do we want to put the meter provider in the context? If so we probably want the Trace and Log stuff there as well
		tracer, meter := initOpenTelemetry(ctx, observabilityServiceClient)
		ctx = observability.ContextWithTracerProvider(ctx, tracer)
		ctx = observability.ContextWithMeterProvider(ctx, meter)
		hmap := map[sdkgo.VerbRef]Handler{}
		for _, handler := range handlers {
			hmap[handler.ref] = handler
		}
		return ctx, &moduleServer{handlers: hmap}, nil
	}
}

// Handler for a Verb.
type Handler struct {
	ref sdkgo.VerbRef
	fn  func(ctx context.Context, req []byte) ([]byte, error)
}

// Handle creates a Handler from a Verb.
func Handle[Req, Resp any](verb func(ctx context.Context, req Req) (Resp, error)) Handler {
	ref := sdkgo.ToVerbRef(verb)
	return Handler{
		ref: ref,
		fn: func(ctx context.Context, reqdata []byte) ([]byte, error) {
			// Decode request.
			var req Req
			err := json.Unmarshal(reqdata, &req)
			if err != nil {
				return nil, errors.Wrapf(err, "invalid request to verb %s", ref)
			}

			// Call Verb.
			resp, err := verb(ctx, req)
			if err != nil {
				return nil, errors.Wrapf(err, "call to verb %s failed", ref)
			}

			respdata, err := json.Marshal(resp)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			return respdata, nil
		},
	}
}

var _ ftlv1connect.VerbServiceHandler = (*moduleServer)(nil)

// This is the server that is compiled into the same binary as user-defined Verbs.
type moduleServer struct {
	handlers map[sdkgo.VerbRef]Handler
}

func (m *moduleServer) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	handler, ok := m.handlers[sdkgo.VerbRefFromProto(req.Msg.Verb)]
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, errors.Errorf("verb %q not found", req.Msg.Verb))
	}

	respdata, err := handler.fn(ctx, req.Msg.Body)
	if err != nil {
		// This makes me slightly ill.
		return connect.NewResponse(&ftlv1.CallResponse{
			Response: &ftlv1.CallResponse_Error_{Error: &ftlv1.CallResponse_Error{Message: err.Error()}},
		}), nil
	}

	return connect.NewResponse(&ftlv1.CallResponse{
		Response: &ftlv1.CallResponse_Body{Body: respdata},
	}), nil
}

func (m *moduleServer) List(ctx context.Context, req *connect.Request[ftlv1.ListRequest]) (*connect.Response[ftlv1.ListResponse], error) {
	out := &ftlv1.ListResponse{}
	for handler := range m.handlers {
		out.Verbs = append(out.Verbs, handler.ToProto())
	}
	return connect.NewResponse(out), nil
}

func (m *moduleServer) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func initOpenTelemetry(ctx context.Context, observabilityServiceClient ftlv1connect.ObservabilityServiceClient) (*trace.TracerProvider, *metric.MeterProvider) {
	logger := log.FromContext(ctx)
	logger.Infof("initOpenTelemetry...")

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("verb"),
	)

	logger.Infof("Initializing TracesExporter...")
	spanExporter := observability.NewSpanExporter(ctx, observabilityServiceClient, observability.SpanExporterConfig{TracesBuffer: 1048576})

	tp := trace.NewTracerProvider(
		trace.WithBatcher(spanExporter),
		trace.WithResource(res))

	logger.Infof("Initializing MetricsExporter...")

	// TODO(wb): How to get MetricsExporterConfig values here? Hardcoded for now.
	metricsExporter := observability.NewMetricsExporter(ctx, observabilityServiceClient, observability.MetricsExporterConfig{MetricsBuffer: 1048576})

	provider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricsExporter)),
	)

	// TODO(wb): Alternatively we could add this to the global otel
	// otel.SetMeterProvider(provider)

	return tp, provider
}
