package drivego

import (
	"context"
	"encoding/json"
	"reflect"
	"runtime"

	"github.com/alecthomas/errors"

	ftlv1 "github.com/TBD54566975/ftl/common/gen/xyz/block/ftl/v1"
)

type ModuleConfig struct{}

func NewModule(handlers ...Handler) func(context.Context, ModuleConfig) (ftlv1.ModuleServiceServer, error) {
	return func(ctx context.Context, mc ModuleConfig) (ftlv1.ModuleServiceServer, error) {
		hmap := map[string]Handler{}
		for _, handler := range handlers {
			hmap[handler.path] = handler
		}
		return &moduleServer{handlers: hmap}, nil
	}
}

type Handler struct {
	path string
	fn   func(ctx context.Context, req []byte) ([]byte, error)
}

// Handle creates a Handler from a Verb.
func Handle[Req, Resp any](verb func(ctx context.Context, req Req) (Resp, error)) Handler {
	name := runtime.FuncForPC(reflect.ValueOf(verb).Pointer()).Name()
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

var _ ftlv1.ModuleServiceServer = (*moduleServer)(nil)

type moduleServer struct {
	handlers map[string]Handler
}

func (*moduleServer) Ping(context.Context, *ftlv1.PingRequest) (*ftlv1.PingResponse, error) {
	return &ftlv1.PingResponse{}, nil
}

func (s *moduleServer) Call(ctx context.Context, req *ftlv1.CallRequest) (*ftlv1.CallResponse, error) {
	handler, ok := s.handlers[req.Verb]
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
