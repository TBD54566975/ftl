package ingress_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/backend/controller/dal"
	"github.com/TBD54566975/ftl/backend/controller/ingress"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/encoding"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
)

func TestIngress(t *testing.T) {
	sch, err := schema.ParseString("", `
		module test {
			data AliasRequest {
				aliased String +alias json "alias"
			}

			data PathParameterRequest {
				username String
			}

			data MissingTypes {
				optional String?
				array [String]
				map {String: String}
				any Any
				unit Unit
			}

			data JsonPayload {
				foo String
			}

			verb getAlias(HttpRequest<test.AliasRequest>) HttpResponse<Empty, Empty>
				+ingress http GET /getAlias

			verb getPath(HttpRequest<test.PathParameterRequest>) HttpResponse<Empty, Empty>
				+ingress http GET /getPath/{username}

			verb postMissingTypes(HttpRequest<test.MissingTypes>) HttpResponse<Empty, Empty>
				+ingress http POST /postMissingTypes

			verb postJsonPayload(HttpRequest<test.JsonPayload>) HttpResponse<Empty, Empty>
				+ingress http POST /postJsonPayload
		}
	`)
	assert.NoError(t, err)

	routes := []dal.IngressRoute{
		{Path: "/getAlias", Module: "test", Verb: "getAlias"},
		{Path: "/getPath/{username}", Module: "test", Verb: "getPath"},
		{Path: "/postMissingTypes", Module: "test", Verb: "postMissingTypes"},
		{Path: "/postJsonPayload", Module: "test", Verb: "postJsonPayload"},
	}

	ctx := log.ContextWithNewDefaultLogger(context.Background())

	for _, test := range []struct {
		name       string
		method     string
		path       string
		query      url.Values
		payload    []byte
		response   *ingress.HTTPResponse
		statusCode int
	}{
		{name: "InvalidRoute",
			method:     "GET",
			path:       "/invalid",
			statusCode: http.StatusNotFound},
		{name: "GetAlias",
			method:     "GET",
			path:       "/getAlias",
			query:      url.Values{"alias": {"value"}},
			response:   &ingress.HTTPResponse{Body: []byte(`{}`)},
			statusCode: http.StatusOK},
	} {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			rec.Body = &bytes.Buffer{}
			if test.response == nil {
				test.response = &ingress.HTTPResponse{Body: []byte(`{}`)}
			}
			req := httptest.NewRequest(test.method, test.path, bytes.NewBuffer(test.payload)).WithContext(ctx)
			req.URL.RawQuery = test.query.Encode()
			reqKey := model.NewRequestName(model.OriginIngress, "test")
			ingress.Handle(sch, reqKey, routes, rec, req, func(ctx context.Context, r *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
				body, err := encoding.Marshal(test.response)
				assert.NoError(t, err)
				return connect.NewResponse(&ftlv1.CallResponse{Response: &ftlv1.CallResponse_Body{Body: body}}), nil
			})
			result := rec.Result()
			defer result.Body.Close()
			assert.Equal(t, test.statusCode, rec.Code, "%s: %s", result.Status, rec.Body.Bytes())
			if rec.Code >= 300 {
				return
			}
			assert.Equal(t, test.response.Body, rec.Body.Bytes())
		})
	}
}
