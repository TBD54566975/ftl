package ftl

import (
	"context"
	"fmt"
	"reflect"

	"connectrpc.com/connect"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/encoding"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/TBD54566975/ftl/internal/modulecontext"
	"github.com/TBD54566975/ftl/internal/rpc"
)

func call[Req, Resp any](ctx context.Context, callee reflection.Ref, req Req, inline Verb[Req, Resp]) (resp Resp, err error) {
	moduleCtx := modulecontext.FromContext(ctx).CurrentContext()
	override, err := moduleCtx.BehaviorForVerb(schema.Ref{Module: callee.Module, Name: callee.Name})
	if err != nil {
		return resp, fmt.Errorf("%s: %w", callee, err)
	}
	if behavior, ok := override.Get(); ok {
		uncheckedResp, err := behavior.Call(ctx, modulecontext.Verb(widenVerb(inline)), req)
		if err != nil {
			return resp, fmt.Errorf("%s: %w", callee, err)
		}
		if r, ok := uncheckedResp.(Resp); ok {
			return r, nil
		}
		return resp, fmt.Errorf("%s: overridden verb had invalid response type %T, expected %v", callee, uncheckedResp, reflect.TypeFor[Resp]())
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

// Call a Verb through the FTL Controller.
func Call[Req, Resp any](ctx context.Context, verb Verb[Req, Resp], req Req) (Resp, error) {
	return call[Req, Resp](ctx, reflection.FuncRef(verb), req, verb)
}

// CallSink calls a Sink through the FTL controller.
func CallSink[Req any](ctx context.Context, sink Sink[Req], req Req) error {
	_, err := call[Req, Unit](ctx, reflection.FuncRef(sink), req, func(ctx context.Context, req Req) (Unit, error) {
		return Unit{}, sink(ctx, req)
	})
	return err
}

// CallSource calls a Source through the FTL controller.
func CallSource[Resp any](ctx context.Context, source Source[Resp]) (Resp, error) {
	return call[Unit, Resp](ctx, reflection.FuncRef(source), Unit{}, func(ctx context.Context, req Unit) (Resp, error) {
		return source(ctx)
	})
}

// CallEmpty calls a Verb with no request or response through the FTL controller.
func CallEmpty(ctx context.Context, empty Empty) error {
	_, err := call[Unit, Unit](ctx, reflection.FuncRef(empty), Unit{}, func(ctx context.Context, req Unit) (Unit, error) {
		return Unit{}, empty(ctx)
	})
	return err
}

func widenVerb[Req, Resp any](verb Verb[Req, Resp]) Verb[any, any] {
	return func(ctx context.Context, uncheckedReq any) (any, error) {
		req, ok := uncheckedReq.(Req)
		if !ok {
			return nil, fmt.Errorf("invalid request type %T for %v, expected %v", uncheckedReq, reflection.FuncRef(verb), reflect.TypeFor[Req]())
		}
		return verb(ctx, req)
	}
}
