package server

import (
	"context"
	"fmt"
	"net/url"
	"runtime/debug"

	"connectrpc.com/connect"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/go-runtime/encoding"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/TBD54566975/ftl/go-runtime/internal"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/maps"
	"github.com/TBD54566975/ftl/internal/modulecontext"
	"github.com/TBD54566975/ftl/internal/observability"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type UserVerbConfig struct {
	FTLEndpoint         *url.URL             `help:"FTL endpoint." env:"FTL_ENDPOINT" required:""`
	ObservabilityConfig observability.Config `embed:"" prefix:"o11y-"`
	Config              []string             `name:"config" short:"C" help:"Paths to FTL project configuration files." env:"FTL_CONFIG" placeholder:"FILE[,FILE,...]" type:"existingfile"`
}

// NewUserVerbServer starts a new code-generated drive for user Verbs.
//
// This function is intended to be used by the code generator.
func NewUserVerbServer(projectName string, moduleName string, handlers ...Handler) plugin.Constructor[ftlv1connect.VerbServiceHandler, UserVerbConfig] {
	return func(ctx context.Context, uc UserVerbConfig) (context.Context, ftlv1connect.VerbServiceHandler, error) {
		verbServiceClient := rpc.Dial(ftlv1connect.NewVerbServiceClient, uc.FTLEndpoint.String(), log.Error)
		ctx = rpc.ContextWithClient(ctx, verbServiceClient)

		moduleContextSupplier := modulecontext.NewModuleContextSupplier(verbServiceClient)
		dynamicCtx, err := modulecontext.NewDynamicContext(ctx, moduleContextSupplier, moduleName)
		if err != nil {
			return nil, nil, fmt.Errorf("could not get config: %w", err)
		}

		ctx = dynamicCtx.ApplyToContext(ctx)
		ctx = internal.WithContext(ctx, internal.New(dynamicCtx))

		err = observability.Init(ctx, true, projectName, moduleName, "HEAD", uc.ObservabilityConfig)
		if err != nil {
			return nil, nil, fmt.Errorf("could not initialize metrics: %w", err)
		}
		hmap := maps.FromSlice(handlers, func(h Handler) (reflection.Ref, Handler) { return h.ref, h })
		return ctx, &moduleServer{handlers: hmap}, nil
	}
}

// Handler for a Verb.
type Handler struct {
	ref reflection.Ref
	fn  func(ctx context.Context, req []byte, metadata map[internal.MetadataKey]string) ([]byte, error)
}

func handler[Req, Resp any](ref reflection.Ref, verb func(ctx context.Context, req Req) (Resp, error)) Handler {
	return Handler{
		ref: ref,
		fn: func(ctx context.Context, reqdata []byte, metadata map[internal.MetadataKey]string) ([]byte, error) {
			ctx = internal.ContextWithCallMetadata(ctx, metadata)

			// Decode request.
			var req Req
			err := encoding.Unmarshal(reqdata, &req)
			if err != nil {
				return nil, fmt.Errorf("invalid request to verb %s: %w", ref, err)
			}

			// Call Verb.
			resp, err := verb(ctx, req)
			if err != nil {
				return nil, fmt.Errorf("call to verb %s failed: %w", ref, err)
			}

			respdata, err := encoding.Marshal(resp)
			if err != nil {
				return nil, err
			}

			return respdata, nil
		},
	}
}

// HandleCall creates a Handler from a Verb.
func HandleCall[Req, Resp any](verb func(ctx context.Context, req Req) (Resp, error)) Handler {
	return handler(reflection.FuncRef(verb), verb)
}

// HandleSink creates a Handler from a Sink with no response.
func HandleSink[Req any](sink func(ctx context.Context, req Req) error) Handler {
	return handler(reflection.FuncRef(sink), func(ctx context.Context, req Req) (ftl.Unit, error) {
		err := sink(ctx, req)
		return ftl.Unit{}, err
	})
}

// HandleSource creates a Handler from a Source with no request.
func HandleSource[Resp any](source func(ctx context.Context) (Resp, error)) Handler {
	return handler(reflection.FuncRef(source), func(ctx context.Context, _ ftl.Unit) (Resp, error) {
		return source(ctx)
	})
}

// HandleEmpty creates a Handler from a Verb with no request or response.
func HandleEmpty(empty func(ctx context.Context) error) Handler {
	return handler(reflection.FuncRef(empty), func(ctx context.Context, _ ftl.Unit) (ftl.Unit, error) {
		err := empty(ctx)
		return ftl.Unit{}, err
	})
}

var _ ftlv1connect.VerbServiceHandler = (*moduleServer)(nil)

// This is the server that is compiled into the same binary as user-defined Verbs.
type moduleServer struct {
	handlers map[reflection.Ref]Handler
}

func (m *moduleServer) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (response *connect.Response[ftlv1.CallResponse], err error) {
	logger := log.FromContext(ctx)
	// Recover from panics and return an error ftlv1.CallResponse.
	defer func() {
		if r := recover(); r != nil {
			var err error
			if rerr, ok := r.(error); ok {
				err = rerr
			} else {
				err = fmt.Errorf("%v", r)
			}
			stack := string(debug.Stack())
			logger.Errorf(err, "panic in verb %s.%s", req.Msg.Verb.Module, req.Msg.Verb.Name)
			response = connect.NewResponse(&ftlv1.CallResponse{Response: &ftlv1.CallResponse_Error_{Error: &ftlv1.CallResponse_Error{
				Message: err.Error(),
				Stack:   &stack,
			}}})
		}
	}()

	handler, ok := m.handlers[reflection.RefFromProto(req.Msg.Verb)]
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("verb %s.%s not found", req.Msg.Verb.Module, req.Msg.Verb.Name))
	}

	metadata := map[internal.MetadataKey]string{}
	if req.Msg.Metadata != nil {
		for _, pair := range req.Msg.Metadata.Values {
			metadata[internal.MetadataKey(pair.Key)] = pair.Value
		}
	}

	respdata, err := handler.fn(ctx, req.Msg.Body, metadata)
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

func (m *moduleServer) GetModuleContext(_ context.Context, _ *connect.Request[ftlv1.ModuleContextRequest], _ *connect.ServerStream[ftlv1.ModuleContextResponse]) error {
	return connect.NewError(connect.CodeUnimplemented, fmt.Errorf("GetModuleContext not implemented"))
}

func (m *moduleServer) AcquireLease(context.Context, *connect.BidiStream[ftlv1.AcquireLeaseRequest, ftlv1.AcquireLeaseResponse]) error {
	return connect.NewError(connect.CodeUnimplemented, fmt.Errorf("AcquireLease not implemented"))
}

func (m *moduleServer) Ping(_ context.Context, _ *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (m *moduleServer) SendFSMEvent(context.Context, *connect.Request[ftlv1.SendFSMEventRequest]) (*connect.Response[ftlv1.SendFSMEventResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("SendFSMEvent not implemented"))
}

func (m *moduleServer) SetNextFSMEvent(ctx context.Context, req *connect.Request[ftlv1.SendFSMEventRequest]) (*connect.Response[ftlv1.SendFSMEventResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("SetNextFSMEvent not implemented"))
}

func (m *moduleServer) PublishEvent(context.Context, *connect.Request[ftlv1.PublishEventRequest]) (*connect.Response[ftlv1.PublishEventResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("PublishEvent not implemented"))
}
