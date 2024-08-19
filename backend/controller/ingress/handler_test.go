package ingress_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

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

			export verb getAlias(HttpRequest<Unit, Unit, test.AliasRequest>) HttpResponse<Empty, Empty>
				+ingress http GET /getAlias

			export verb getPath(HttpRequest<Unit, test.PathParameterRequest, Unit>) HttpResponse<Empty, Empty>
				+ingress http GET /getPath/{username}

			export verb postMissingTypes(HttpRequest<test.MissingTypes, Unit, Unit>) HttpResponse<Empty, Empty>
				+ingress http POST /postMissingTypes

			export verb postJsonPayload(HttpRequest<test.JsonPayload, Unit, Unit>) HttpResponse<Empty, Empty>
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
		response   optional.Option[ingress.HTTPResponse]
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
			response:   optional.Some(ingress.HTTPResponse{Body: []byte(`{}`)}),
			statusCode: http.StatusOK},
	} {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			rec.Body = &bytes.Buffer{}
			var response ingress.HTTPResponse
			var ok bool
			if response, ok = test.response.Get(); ok {
				response = ingress.HTTPResponse{Body: []byte(`{}`)}
			}
			req := httptest.NewRequest(test.method, test.path, bytes.NewBuffer(test.payload)).WithContext(ctx)
			req.URL.RawQuery = test.query.Encode()
			reqKey := model.NewRequestKey(model.OriginIngress, "test")
			ingress.Handle(time.Now(), sch, reqKey, routes, rec, req, func(ctx context.Context, r *connect.Request[ftlv1.CallRequest], requestKey optional.Option[model.RequestKey], parentRequestKey optional.Option[model.RequestKey], requestSource string) (*connect.Response[ftlv1.CallResponse], error) {
				body, err := encoding.Marshal(response)
				assert.NoError(t, err)
				return connect.NewResponse(&ftlv1.CallResponse{Response: &ftlv1.CallResponse_Body{Body: body}}), nil
			})
			result := rec.Result()
			defer result.Body.Close()
			assert.Equal(t, test.statusCode, rec.Code, "%s: %s", result.Status, rec.Body.Bytes())
			if rec.Code >= 300 {
				return
			}
			assert.Equal(t, response.Body, rec.Body.Bytes())
		})
	}
}
