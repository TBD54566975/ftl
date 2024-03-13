package ftl

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"connectrpc.com/connect"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	"github.com/TBD54566975/ftl/go-runtime/encoding"
	"github.com/TBD54566975/ftl/internal/rpc"
)

func call[Req, Resp any](ctx context.Context, callee *schemapb.VerbRef, req Req) (resp Resp, err error) {
	client := rpc.ClientFromContext[ftlv1connect.VerbServiceClient](ctx)
	reqData, err := encoding.Marshal(req)
	if err != nil {
		return resp, fmt.Errorf("%s: failed to marshal request: %w", callee, err)
	}
	cresp, err := client.Call(ctx, connect.NewRequest(&ftlv1.CallRequest{Verb: callee, Body: reqData}))
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
func Call[Req, Resp any](ctx context.Context, verb Verb[Req, Resp], req Req) (resp Resp, err error) {
	callee := VerbToRef(verb)
	return call[Req, Resp](ctx, callee.ToProto(), req)
}

// CallSink calls a Sink through the FTL controller.
func CallSink[Req any](ctx context.Context, sink Sink[Req], req Req) error {
	callee := SinkToRef(sink)
	verbRef := &schemapb.VerbRef{Module: callee.Module, Name: callee.Name}
	_, err := call[Req, Unit](ctx, verbRef, req)
	return err
}

// CallSource calls a Source through the FTL controller.
func CallSource[Resp any](ctx context.Context, source Source[Resp]) (Resp, error) {
	callee := SourceToRef(source)
	verbRef := &schemapb.VerbRef{Module: callee.Module, Name: callee.Name}
	fmt.Printf("source ref2: %v\n", verbRef)
	return call[Unit, Resp](ctx, verbRef, Unit{})
}

// CallEmpty calls a Verb with no request or response through the FTL controller.
func CallEmpty(ctx context.Context, empty Empty) error {
	callee := EmptyToRef(empty)
	verbRef := &schemapb.VerbRef{Module: callee.Module, Name: callee.Name}

	_, err := call[Unit, Unit](ctx, verbRef, Unit{})
	return err
}

// VerbToRef returns the FTL reference for a Verb.
func VerbToRef[Req, Resp any](verb Verb[Req, Resp]) Ref {
	ref := runtime.FuncForPC(reflect.ValueOf(verb).Pointer()).Name()
	return goRefToFTLRef(ref)
}

// SinkToRef returns the FTL reference for a Sink.
func SinkToRef[Req any](sink Sink[Req]) Ref {
	ref := runtime.FuncForPC(reflect.ValueOf(sink).Pointer()).Name()
	return goRefToFTLRef(ref)
}

// SourceToRef returns the FTL reference for a Source.
func SourceToRef[Resp any](source Source[Resp]) Ref {
	ref := runtime.FuncForPC(reflect.ValueOf(source).Pointer()).Name()
	return goRefToFTLRef(ref)
}

// EmptyToRef returns the FTL reference for an Empty.
func EmptyToRef(empty Empty) Ref {
	ref := runtime.FuncForPC(reflect.ValueOf(empty).Pointer()).Name()
	return goRefToFTLRef(ref)
}

func goRefToFTLRef(ref string) Ref {
	parts := strings.Split(ref[strings.LastIndex(ref, "/")+1:], ".")
	return Ref{parts[len(parts)-2], strcase.ToLowerCamel(parts[len(parts)-1])}
}
