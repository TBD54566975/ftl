package ftl

import (
	"connectrpc.com/connect"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	"github.com/TBD54566975/ftl/go-runtime/encoding"
	"github.com/TBD54566975/ftl/internal/rpc"

	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

func call[Req, Resp any](ctx context.Context, callee *schemapb.Ref, req Req) (resp Resp, err error) {
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
func Call[Req, Resp any](ctx context.Context, verb Verb[Req, Resp], req Req) (Resp, error) {
	return call[Req, Resp](ctx, CallToSchemaRef(verb), req)
}

// CallSink calls a Sink through the FTL controller.
func CallSink[Req any](ctx context.Context, sink Sink[Req], req Req) error {
	_, err := call[Req, Unit](ctx, CallToSchemaRef(sink), req)
	return err
}

// CallSource calls a Source through the FTL controller.
func CallSource[Resp any](ctx context.Context, source Source[Resp]) (Resp, error) {
	return call[Unit, Resp](ctx, CallToSchemaRef(source), Unit{})
}

// CallEmpty calls a Verb with no request or response through the FTL controller.
func CallEmpty(ctx context.Context, empty Empty) error {
	_, err := call[Unit, Unit](ctx, CallToSchemaRef(empty), Unit{})
	return err
}

// CallToRef returns the Ref for a Verb, Sink, Source, or Empty.
func CallToRef(call any) Ref {
	ref := runtime.FuncForPC(reflect.ValueOf(call).Pointer()).Name()
	return goRefToFTLRef(ref)
}

// CallToSchemaRef returns the Ref for a Verb, Sink, Source, or Empty as a Schema Ref.
func CallToSchemaRef(call any) *schemapb.Ref {
	ref := CallToRef(call)
	return ref.ToProto()
}

func goRefToFTLRef(ref string) Ref {
	parts := strings.Split(ref[strings.LastIndex(ref, "/")+1:], ".")
	return Ref{parts[len(parts)-2], strcase.ToLowerCamel(parts[len(parts)-1])}
}
