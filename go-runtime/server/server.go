package server

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"runtime/debug"
	"strings"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"

	deploymentconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/deployment/v1/ftlv1connect"
	ftlv1connect2 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/lease/v1/ftlv1connect"
	pubconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/publish/v1/publishpbconnect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/common/encoding"
	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/common/reflection"
	"github.com/TBD54566975/ftl/common/schema"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/internal"
	"github.com/TBD54566975/ftl/internal/deploymentcontext"
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
func NewUserVerbServer(projectName string, moduleName string, handlers ...Handler) plugin.Constructor[ftlv1connect.VerbServiceHandler, UserVerbConfig] {
	return func(ctx context.Context, uc UserVerbConfig) (context.Context, ftlv1connect.VerbServiceHandler, error) {
		moduleServiceClient := rpc.Dial(deploymentconnect.NewDeploymentServiceClient, uc.FTLEndpoint.String(), log.Error)
		ctx = rpc.ContextWithClient(ctx, moduleServiceClient)
		verbServiceClient := rpc.Dial(ftlv1connect.NewVerbServiceClient, uc.FTLEndpoint.String(), log.Error)
		ctx = rpc.ContextWithClient(ctx, verbServiceClient)
		pubClient := rpc.Dial(pubconnect.NewPublishServiceClient, uc.FTLEndpoint.String(), log.Error)
		ctx = rpc.ContextWithClient(ctx, pubClient)
		leaseClient := rpc.Dial(ftlv1connect2.NewLeaseServiceClient, uc.FTLEndpoint.String(), log.Error)
		ctx = rpc.ContextWithClient(ctx, leaseClient)

		moduleContextSupplier := deploymentcontext.NewDeploymentContextSupplier(moduleServiceClient)
		// FTL_DEPLOYMENT is set by the FTL runtime.
		dynamicCtx, err := deploymentcontext.NewDynamicContext(ctx, moduleContextSupplier, os.Getenv("FTL_DEPLOYMENT"))
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

func HandleCall[Req, Resp any](verb any) Handler {
	ref := reflection.FuncRef(verb)
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
			ctx = observability.AddSpanContextToLogger(ctx)

			// InvokeVerb Verb.
			resp, err := InvokeVerb[Req, Resp](ref)(ctx, req)
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

func HandleSink[Req any](verb any) Handler {
	return HandleCall[Req, ftl.Unit](verb)
}

func HandleSource[Resp any](verb any) Handler {
	return HandleCall[ftl.Unit, Resp](verb)
}

func HandleEmpty(verb any) Handler {
	return HandleCall[ftl.Unit, ftl.Unit](verb)
}

func InvokeVerb[Req, Resp any](ref reflection.Ref) func(ctx context.Context, req Req) (resp Resp, err error) {
	return func(ctx context.Context, req Req) (resp Resp, err error) {
		request := optional.Some[any](req)
		if reflect.TypeFor[Req]() == reflect.TypeFor[ftl.Unit]() {
			request = optional.None[any]()
		}

		out, err := reflection.CallVerb(reflection.Ref{Module: ref.Module, Name: ref.Name})(ctx, request)
		if err != nil {
			return resp, err
		}

		var respValue any
		if r, ok := out.Get(); ok {
			respValue = r
		} else {
			respValue = ftl.Unit{}
		}
		resp, ok := respValue.(Resp)
		if !ok {
			return resp, fmt.Errorf("unexpected response type from verb %s: %T", ref, resp)
		}
		return resp, err
	}
}

func VerbClient[Verb, Req, Resp any]() reflection.VerbResource {
	fnCall := call[Verb, Req, Resp]()
	return func() reflect.Value {
		return reflect.ValueOf(fnCall)
	}
}

func SinkClient[Verb, Req any]() reflection.VerbResource {
	fnCall := call[Verb, Req, ftl.Unit]()
	sink := func(ctx context.Context, req Req) error {
		_, err := fnCall(ctx, req)
		return err
	}
	return func() reflect.Value {
		return reflect.ValueOf(sink)
	}
}

func SourceClient[Verb, Resp any]() reflection.VerbResource {
	fnCall := call[Verb, ftl.Unit, Resp]()
	source := func(ctx context.Context) (Resp, error) {
		return fnCall(ctx, ftl.Unit{})
	}
	return func() reflect.Value {
		return reflect.ValueOf(source)
	}
}

func EmptyClient[Verb any]() reflection.VerbResource {
	fnCall := call[Verb, ftl.Unit, ftl.Unit]()
	source := func(ctx context.Context) error {
		_, err := fnCall(ctx, ftl.Unit{})
		return err
	}
	return func() reflect.Value {
		return reflect.ValueOf(source)
	}
}

func call[Verb, Req, Resp any]() func(ctx context.Context, req Req) (resp Resp, err error) {
	typ := reflect.TypeFor[Verb]()
	if typ.Kind() != reflect.Func {
		panic(fmt.Sprintf("Cannot register %s: expected function, got %s", typ, typ.Kind()))
	}
	callee := reflection.TypeRef[Verb]()
	callee.Name = strings.TrimSuffix(callee.Name, "Client")
	return func(ctx context.Context, req Req) (resp Resp, err error) {
		ref := reflection.Ref{Module: callee.Module, Name: callee.Name}
		moduleCtx := deploymentcontext.FromContext(ctx).CurrentContext()
		override, err := moduleCtx.BehaviorForVerb(schema.Ref{Module: ref.Module, Name: ref.Name})
		if err != nil {
			return resp, fmt.Errorf("%s: %w", ref, err)
		}
		if behavior, ok := override.Get(); ok {
			uncheckedResp, err := behavior.Call(ctx, deploymentcontext.Verb(widenVerb(InvokeVerb[Req, Resp](ref))), req)
			if err != nil {
				return resp, fmt.Errorf("%s: %w", ref, err)
			}
			if r, ok := uncheckedResp.(Resp); ok {
				return r, nil
			}
			return resp, fmt.Errorf("%s: overridden verb had invalid response type %T, expected %v", ref,
				uncheckedResp, reflect.TypeFor[Resp]())
		}

		reqData, err := encoding.Marshal(req)
		if err != nil {
			return resp, fmt.Errorf("%s: failed to marshal request: %w", callee, err)
		}

		client := rpc.ClientFromContext[ftlv1connect.VerbServiceClient](ctx)
		cresp, err := client.Call(ctx, connect.NewRequest(&ftlv1.CallRequest{Verb: callee.ToProto(), Body: reqData}))
		if err != nil {
			return resp, fmt.Errorf("%s: failed to call Verb: %w", callee, err)
		}
		switch cresp := cresp.Msg.Response.(type) {
		case *ftlv1.CallResponse_Error_:
			return resp, fmt.Errorf("%s: %s", callee, cresp.Error.Message)

		case *ftlv1.CallResponse_Body:
			err = encoding.Unmarshal(cresp.Body, &resp)
			if err != nil {
				return resp, fmt.Errorf("%s: failed to decode response: %w", callee, err)
			}
			return resp, nil

		default:
			panic(fmt.Sprintf("%s: invalid response type %T", callee, cresp))
		}
	}
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

func (m *moduleServer) Ping(_ context.Context, _ *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func widenVerb[Req, Resp any](verb ftl.Verb[Req, Resp]) ftl.Verb[any, any] {
	return func(ctx context.Context, uncheckedReq any) (any, error) {
		req, ok := uncheckedReq.(Req)
		if !ok {
			return nil, fmt.Errorf("invalid request type %T for %v, expected %v", uncheckedReq, reflection.FuncRef(verb), reflect.TypeFor[Req]())
		}
		return verb(ctx, req)
	}
}
