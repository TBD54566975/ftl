package server

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"connectrpc.com/connect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/go-runtime/encoding"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/TBD54566975/ftl/go-runtime/internal"
	"github.com/TBD54566975/ftl/internal/modulecontext"
	"github.com/TBD54566975/ftl/internal/observability"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/alecthomas/types/optional"
)

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
				return nil, fmt.Errorf("failed to marshal response from verb %s: %w", ref, err)
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
		moduleCtx := modulecontext.FromContext(ctx).CurrentContext()
		override, err := moduleCtx.BehaviorForVerb(schema.Ref{Module: ref.Module, Name: ref.Name})
		if err != nil {
			return resp, fmt.Errorf("%s: %w", ref, err)
		}
		if behavior, ok := override.Get(); ok {
			uncheckedResp, err := behavior.Call(ctx, modulecontext.Verb(widenVerb(InvokeVerb[Req, Resp](ref))), req)
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
