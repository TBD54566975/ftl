package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"connectrpc.com/connect"
	"github.com/iancoleman/strcase"

	"github.com/TBD54566975/ftl/backend/common/rpc"
	"github.com/TBD54566975/ftl/go-runtime/encoding"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

// Call a Verb through the FTL Controller.
func Call[Req, Resp any](ctx context.Context, verb Verb[Req, Resp], req Req) (resp Resp, err error) {
	callee := ToVerbRef(verb)
	client := rpc.ClientFromContext[ftlv1connect.VerbServiceClient](ctx)
	reqData, err := encoding.Marshal(req)
	if err != nil {
		return resp, fmt.Errorf("%s: failed to marshal request: %w", callee, err)
	}
	cresp, err := client.Call(ctx, connect.NewRequest(&ftlv1.CallRequest{Verb: callee.ToProto(), Body: reqData}))
	if err != nil {
		return resp, fmt.Errorf("%s: failed to call Verb: %w", callee, err)
	}
	switch cresp := cresp.Msg.Response.(type) {
	case *ftlv1.CallResponse_Error_:
		return resp, fmt.Errorf("%s: %s", callee, cresp.Error.Message)

	case *ftlv1.CallResponse_Body:
		err = json.Unmarshal(cresp.Body, &resp)
		if err != nil {
			return resp, fmt.Errorf("%s: failed to decode response: %w", callee, err)
		}
		return resp, nil

	default:
		panic(fmt.Sprintf("%s: invalid response type %T", callee, cresp))
	}
}

// ToVerbRef returns the FTL reference for a Verb.
func ToVerbRef[Req, Resp any](verb Verb[Req, Resp]) VerbRef {
	ref := runtime.FuncForPC(reflect.ValueOf(verb).Pointer()).Name()
	return goRefToFTLRef(ref)
}

func goRefToFTLRef(ref string) VerbRef {
	parts := strings.Split(ref[strings.LastIndex(ref, "/")+1:], ".")
	return VerbRef{parts[len(parts)-2], strcase.ToLowerCamel(parts[len(parts)-1])}
}
