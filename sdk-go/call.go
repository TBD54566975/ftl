package sdkgo

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"
	"github.com/iancoleman/strcase"

	"github.com/TBD54566975/ftl/internal/rpc"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

// Call a Verb through the Agent.
func Call[Req, Resp any](ctx context.Context, verb Verb[Req, Resp], req Req) (resp Resp, err error) {
	callee := ToVerbRef(verb)
	client := rpc.ClientFromContext[ftlv1connect.VerbServiceClient](ctx)
	reqData, err := json.Marshal(req)
	if err != nil {
		return resp, errors.Wrapf(err, "%s: failed to marshal request", callee)
	}
	cresp, err := client.Call(ctx, connect.NewRequest(&ftlv1.CallRequest{Verb: callee.ToProto(), Body: reqData}))
	if err != nil {
		return resp, errors.Wrapf(err, "%s: failed to call Verb", callee)
	}
	switch cresp := cresp.Msg.Response.(type) {
	case *ftlv1.CallResponse_Error_:
		return resp, errors.Errorf("%s: %s", callee, cresp.Error.Message)

	case *ftlv1.CallResponse_Body:
		err = json.Unmarshal(cresp.Body, &resp)
		if err != nil {
			return resp, errors.Wrapf(err, "%s: failed to decode response", callee)
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
