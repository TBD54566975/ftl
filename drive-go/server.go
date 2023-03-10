package drivego

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"

	"github.com/alecthomas/errors"

	ftlv1 "github.com/TBD54566975/ftl/common/gen/xyz/block/ftl/v1"
)

type UserVerbConfig struct{}

// NewUserVerbServer starts a new code-generated FTL.drive server for user Verbs.
//
// This function is intended to be used by the code generator.
func NewUserVerbServer(handlers ...Handler) func(context.Context, UserVerbConfig) (ftlv1.DriveServiceServer, error) {
	return func(ctx context.Context, mc UserVerbConfig) (ftlv1.DriveServiceServer, error) {
		hmap := map[string]Handler{}
		for _, handler := range handlers {
			hmap[handler.path] = handler
		}
		return &moduleServer{handlers: hmap}, nil
	}
}

// Handler for a Verb.
type Handler struct {
	path string
	fn   func(ctx context.Context, req []byte) ([]byte, error)
}

// Handle creates a Handler from a Verb.
func Handle[Req, Resp any](verb func(ctx context.Context, req Req) (Resp, error)) Handler {
	name := runtime.FuncForPC(reflect.ValueOf(verb).Pointer()).Name()
	fmt.Println(name)
	return Handler{
		path: name,
		fn: func(ctx context.Context, reqdata []byte) ([]byte, error) {
			// Decode request.
			var req Req
			err := json.Unmarshal(reqdata, &req)
			if err != nil {
				return nil, errors.Wrapf(err, "invalid request to verb %s", name)
			}

			// Call Verb.
			resp, err := verb(ctx, req)
			if err != nil {
				return nil, errors.Wrapf(err, "call to verb %s failed", name)
			}

			respdata, err := json.Marshal(resp)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			return respdata, nil
		},
	}
}

var _ ftlv1.DriveServiceServer = (*moduleServer)(nil)

type moduleServer struct {
	handlers map[string]Handler
}

func (m *moduleServer) List(ctx context.Context, req *ftlv1.ListRequest) (*ftlv1.ListResponse, error) {
	out := &ftlv1.ListResponse{}
	for handler := range m.handlers {
		out.Verbs = append(out.Verbs, handler)
	}
	return out, nil
}

func (*moduleServer) FileChange(context.Context, *ftlv1.FileChangeRequest) (*ftlv1.FileChangeResponse, error) {
	return &ftlv1.FileChangeResponse{}, nil
}

func (*moduleServer) Ping(context.Context, *ftlv1.PingRequest) (*ftlv1.PingResponse, error) {
	return &ftlv1.PingResponse{}, nil
}

func (m *moduleServer) Call(ctx context.Context, req *ftlv1.CallRequest) (*ftlv1.CallResponse, error) {
	handler, ok := m.handlers[req.Verb]
	if !ok {
		return nil, errors.Errorf("verb %q not found", req.Verb)
	}
	respdata, err := handler.fn(ctx, req.Body)
	if err != nil {
		// This makes me slightly ill.
		return &ftlv1.CallResponse{
			Response: &ftlv1.CallResponse_Error_{Error: &ftlv1.CallResponse_Error{Message: err.Error()}},
		}, nil
	}
	return &ftlv1.CallResponse{
		Response: &ftlv1.CallResponse_Body{Body: respdata},
	}, nil
}
