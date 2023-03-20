package sdkgo

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/alecthomas/errors"
	"github.com/iancoleman/strcase"

	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
)

// Call a Verb through the Agent.
func Call[Req, Resp any](ctx context.Context, verb func(ctx context.Context, req Req) (Resp, error), req Req) (resp Resp, err error) {
	callee := VerbRef(verb)
	client := ClientFromContext(ctx)
	reqData, err := json.Marshal(req)
	if err != nil {
		return resp, errors.Wrapf(err, "%s: failed to marshal request", callee)
	}
	cresp, err := client.Call(ctx, &ftlv1.CallRequest{Verb: callee, Body: reqData})
	if err != nil {
		return resp, errors.Wrapf(err, "%s: failed to call Verb", callee)
	}
	switch cresp := cresp.Response.(type) {
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

// VerbRef returns the FTL reference for a Verb.
func VerbRef[Req, Resp any](verb func(ctx context.Context, req Req) (Resp, error)) string {
	ref := runtime.FuncForPC(reflect.ValueOf(verb).Pointer()).Name()
	return goRefToFTLRef(ref)
}

func goRefToFTLRef(ref string) string {
	parts := strings.Split(ref[strings.LastIndex(ref, "/")+1:], ".")
	return parts[len(parts)-2] + "." + strcase.ToLowerCamel(parts[len(parts)-1])
}
