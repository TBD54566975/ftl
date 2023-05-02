package drivego

import (
	"context"
	"encoding/json"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"

	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	sdkgo "github.com/TBD54566975/ftl/sdk-go"
)

type UserVerbConfig struct{}

// NewUserVerbServer starts a new code-generated drive for user Verbs.
//
// This function is intended to be used by the code generator.
func NewUserVerbServer(handlers ...Handler) func(context.Context, UserVerbConfig) (ftlv1connect.VerbServiceHandler, error) {
	return func(ctx context.Context, mc UserVerbConfig) (ftlv1connect.VerbServiceHandler, error) {
		hmap := map[sdkgo.VerbRef]Handler{}
		for _, handler := range handlers {
			hmap[handler.ref] = handler
		}
		return &moduleServer{handlers: hmap}, nil
	}
}

// Handler for a Verb.
type Handler struct {
	ref sdkgo.VerbRef
	fn  func(ctx context.Context, req []byte) ([]byte, error)
}

// Handle creates a Handler from a Verb.
func Handle[Req, Resp any](verb func(ctx context.Context, req Req) (Resp, error)) Handler {
	ref := sdkgo.ToVerbRef(verb)
	return Handler{
		ref: ref,
		fn: func(ctx context.Context, reqdata []byte) ([]byte, error) {
			// Decode request.
			var req Req
			err := json.Unmarshal(reqdata, &req)
			if err != nil {
				return nil, errors.Wrapf(err, "invalid request to verb %s", ref)
			}

			// Call Verb.
			resp, err := verb(ctx, req)
			if err != nil {
				return nil, errors.Wrapf(err, "call to verb %s failed", ref)
			}

			respdata, err := json.Marshal(resp)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			return respdata, nil
		},
	}
}

var _ ftlv1connect.VerbServiceHandler = (*moduleServer)(nil)

// This is the server that is compiled into the same binary as user-defined Verbs.
type moduleServer struct {
	handlers map[sdkgo.VerbRef]Handler
}

func (m *moduleServer) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	handler, ok := m.handlers[sdkgo.VerbRefFromProto(req.Msg.Verb)]

	if !ok {
		return nil, errors.Errorf("verb %q not found", req.Msg.Verb)
	}

	respdata, err := handler.fn(ctx, req.Msg.Body)
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

func (m *moduleServer) List(ctx context.Context, req *connect.Request[ftlv1.ListRequest]) (*connect.Response[ftlv1.ListResponse], error) {
	out := &ftlv1.ListResponse{}
	for handler := range m.handlers {
		out.Verbs = append(out.Verbs, handler.ToProto())
	}
	return connect.NewResponse(out), nil
}

func (m *moduleServer) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (m *moduleServer) Send(ctx context.Context, req *connect.Request[ftlv1.SendRequest]) (*connect.Response[ftlv1.SendResponse], error) {
	panic("unimplemented")
}
