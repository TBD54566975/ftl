package server

import (
	"context"
	"encoding/json"
	"net/url"
	"runtime/debug"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"

	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/maps"
	"github.com/TBD54566975/ftl/backend/common/observability"
	"github.com/TBD54566975/ftl/backend/common/plugin"
	"github.com/TBD54566975/ftl/backend/common/rpc"
	sdkgo "github.com/TBD54566975/ftl/go-runtime/sdk"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

type UserVerbConfig struct {
	FTLEndpoint         *url.URL             `help:"FTL endpoint." env:"FTL_ENDPOINT" required:""`
	ObservabilityConfig observability.Config `embed:"" prefix:"o11y-"`
}

// NewUserVerbServer starts a new code-generated drive for user Verbs.
//
// This function is intended to be used by the code generator.
func NewUserVerbServer(moduleName string, handlers ...Handler) plugin.Constructor[ftlv1connect.VerbServiceHandler, UserVerbConfig] {
	return func(ctx context.Context, uc UserVerbConfig) (context.Context, ftlv1connect.VerbServiceHandler, error) {
		verbServiceClient := rpc.Dial(ftlv1connect.NewVerbServiceClient, uc.FTLEndpoint.String(), log.Error)
		ctx = rpc.ContextWithClient(ctx, verbServiceClient)

		err := observability.Init(ctx, moduleName, "HEAD", uc.ObservabilityConfig)
		if err != nil {
			return nil, nil, errors.WithStack(err)
		}
		hmap := maps.FromSlice(handlers, func(h Handler) (sdkgo.VerbRef, Handler) { return h.ref, h })
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

func (m *moduleServer) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (response *connect.Response[ftlv1.CallResponse], err error) {
	logger := log.FromContext(ctx)
	// Recover from panics and return an error ftlv1.CallResponse.
	defer func() {
		if r := recover(); r != nil {
			var err error
			if rerr, ok := r.(error); ok {
				err = errors.WithStack(rerr)
			} else {
				err = errors.Errorf("%v", r)
			}
			stack := string(debug.Stack())
			logger.Errorf(err, "panic in verb %s.%s", req.Msg.Verb.Module, req.Msg.Verb.Name)
			response = connect.NewResponse(&ftlv1.CallResponse{Response: &ftlv1.CallResponse_Error_{Error: &ftlv1.CallResponse_Error{
				Message: err.Error(),
				Stack:   &stack,
			}}})
		}
	}()
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

func (m *moduleServer) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}
