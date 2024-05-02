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
	"github.com/TBD54566975/ftl/go-runtime/modulecontext"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/maps"
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
func NewUserVerbServer(moduleName string, handlers ...Handler) plugin.Constructor[ftlv1connect.VerbServiceHandler, UserVerbConfig] {
	return func(ctx context.Context, uc UserVerbConfig) (context.Context, ftlv1connect.VerbServiceHandler, error) {
		verbServiceClient := rpc.Dial(ftlv1connect.NewVerbServiceClient, uc.FTLEndpoint.String(), log.Error)
		ctx = rpc.ContextWithClient(ctx, verbServiceClient)

		resp, err := verbServiceClient.GetModuleContext(ctx, connect.NewRequest(&ftlv1.ModuleContextRequest{
			Module: moduleName,
		}))
		if err != nil {
			return nil, nil, fmt.Errorf("could not get config: %w", err)
		}
		moduleCtx, err := modulecontext.FromProto(ctx, resp.Msg)
		if err != nil {
			return nil, nil, err
		}
		ctx = moduleCtx.ApplyToContext(ctx)

		err = observability.Init(ctx, moduleName, "HEAD", uc.ObservabilityConfig)
		if err != nil {
			return nil, nil, err
		}
		hmap := maps.FromSlice(handlers, func(h Handler) (ftl.Ref, Handler) { return h.ref, h })
		return ctx, &moduleServer{handlers: hmap}, nil
	}
}

// Handler for a Verb.
type Handler struct {
	ref ftl.Ref
	fn  func(ctx context.Context, req []byte) ([]byte, error)
}

func handler[Req, Resp any](ref ftl.Ref, verb func(ctx context.Context, req Req) (Resp, error)) Handler {
	return Handler{
		ref: ref,
		fn: func(ctx context.Context, reqdata []byte) ([]byte, error) {
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
	return handler(ftl.FuncRef(verb), verb)
}

// HandleSink creates a Handler from a Sink with no response.
func HandleSink[Req any](sink func(ctx context.Context, req Req) error) Handler {
	return handler(ftl.FuncRef(sink), func(ctx context.Context, req Req) (ftl.Unit, error) {
		err := sink(ctx, req)
		return ftl.Unit{}, err
	})
}

// HandleSource creates a Handler from a Source with no request.
func HandleSource[Resp any](source func(ctx context.Context) (Resp, error)) Handler {
	return handler(ftl.FuncRef(source), func(ctx context.Context, _ ftl.Unit) (Resp, error) {
		return source(ctx)
	})
}

// HandleEmpty creates a Handler from a Verb with no request or response.
func HandleEmpty(empty func(ctx context.Context) error) Handler {
	return handler(ftl.FuncRef(empty), func(ctx context.Context, _ ftl.Unit) (ftl.Unit, error) {
		err := empty(ctx)
		return ftl.Unit{}, err
	})
}

var _ ftlv1connect.VerbServiceHandler = (*moduleServer)(nil)

// This is the server that is compiled into the same binary as user-defined Verbs.
type moduleServer struct {
	handlers map[ftl.Ref]Handler
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
	handler, ok := m.handlers[ftl.RefFromProto(req.Msg.Verb)]
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("verb %q not found", req.Msg.Verb))
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

func (m *moduleServer) GetModuleContext(ctx context.Context, req *connect.Request[ftlv1.ModuleContextRequest]) (*connect.Response[ftlv1.ModuleContextResponse], error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *moduleServer) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}
