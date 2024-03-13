package server

import (
	"context"
	"fmt"
	"net/url"
	"runtime/debug"

	"connectrpc.com/connect"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	cf "github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/go-runtime/encoding"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
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

		// Add config manager to context.
		cr := &cf.ProjectConfigResolver[cf.Configuration]{Config: uc.Config}
		cm, err := cf.NewConfigurationManager(ctx, cr)
		if err != nil {
			return nil, nil, err
		}
		ctx = cf.ContextWithConfig(ctx, cm)

		// Add secrets manager to context.
		sr := &cf.ProjectConfigResolver[cf.Secrets]{Config: uc.Config}
		sm, err := cf.NewSecretsManager(ctx, sr)
		if err != nil {
			return nil, nil, err
		}
		ctx = cf.ContextWithSecrets(ctx, sm)

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

// Handle creates a Handler from a Verb.
func Handle[Req, Resp any](verb func(ctx context.Context, req Req) (Resp, error)) Handler {
	ref := ftl.VerbToRef(verb)
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

func (m *moduleServer) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}
