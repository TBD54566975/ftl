package ingress

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/backend/controller/dal"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/encoding"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

type AliasRequest struct {
	Aliased string `json:"alias"`
}

type PathParameterRequest struct {
	Username string
}

type QueryParameterRequest struct {
	Foo ftl.Option[string] `json:"foo,omitempty"`
}

type MissingTypes struct {
	Optional ftl.Option[string] `json:"optional,omitempty"`
	Array    []string           `json:"array,omitempty"`
	Map      map[string]string  `json:"map,omitempty"`
	Any      any                `json:"any,omitempty"`
	Unit     ftl.Unit           `json:"unit,omitempty"`
}

type PostJSONPayload struct {
	Foo string
}

// HTTPRequest mirrors builtin.HttpRequest.
type HTTPRequest[Body any, Path any, Query any] struct {
	Body           Body
	Headers        map[string][]string `json:"headers,omitempty"`
	Method         string
	Path           string
	PathParameters Path  `json:"pathParameters,omitempty"`
	Query          Query `json:"query,omitempty"`
}

func TestBuildRequestBody(t *testing.T) {
	sch, err := schema.ParseString("test", `
		module test {
			data AliasRequest {
				aliased String +alias json "alias"
			}

			data QueryParameterRequest {
				foo String?
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

			export verb optionalQuery(HttpRequest<Unit, Unit, test.QueryParameterRequest>) HttpResponse<Empty, Empty>
				+ingress http GET /optionalQuery

			export verb postMissingTypes(HttpRequest<test.MissingTypes, Unit, Unit>) HttpResponse<Empty, Empty>
				+ingress http POST /postMissingTypes

			export verb postJsonPayload(HttpRequest<test.JsonPayload, Unit, Unit>) HttpResponse<Empty, Empty>
				+ingress http POST /postJsonPayload

			export verb getById(HttpRequest<Unit, Int, Unit>) HttpResponse<Empty, Empty>
				+ingress http GET /getbyid/{id}

			export verb mapQuery(HttpRequest<Unit, Unit, {String: String}>) HttpResponse<Empty, Empty>
				+ingress http GET /mapQuery

			export verb multiMapQuery(HttpRequest<Unit, Unit, {String: [String]}>) HttpResponse<Empty, Empty>
				+ingress http GET /multiMapQuery
		}
	`)
	assert.NoError(t, err)
	for _, test := range []struct {
		name      string
		verb      string
		method    string
		path      string
		routePath string
		query     url.Values
		body      obj
		expected  any
		err       string
	}{
		{name: "UnknownVerb",
			verb: "unknown",
			err:  fmt.Errorf("could not resolve reference test.unknown: %w", schema.ErrNotFound).Error()},
		{name: "UnknownModule",
			verb: "unknown",
			err:  fmt.Errorf("could not resolve reference test.unknown: %w", schema.ErrNotFound).Error()},
		{name: "QueryParameterDecoding",
			verb:      "getAlias",
			method:    "GET",
			path:      "/getAlias",
			routePath: "/getAlias",
			query: map[string][]string{
				"alias": {"value"},
			},
			expected: HTTPRequest[ftl.Unit, map[string]string, AliasRequest]{
				Method: "GET",
				Path:   "/getAlias",
				Query: AliasRequest{
					Aliased: "value",
				},
			},
		},
		{name: "AllowMissingFieldTypes",
			verb:      "postMissingTypes",
			method:    "POST",
			path:      "/postMissingTypes",
			routePath: "/postMissingTypes",
			expected: HTTPRequest[MissingTypes, map[string]string, map[string][]string]{
				Method: "POST",
				Path:   "/postMissingTypes",
				Body:   MissingTypes{},
			},
		},
		{name: "JSONPayload",
			verb:      "postJsonPayload",
			method:    "POST",
			path:      "/postJsonPayload",
			routePath: "/postJsonPayload",
			body:      obj{"foo": "bar"},
			expected: HTTPRequest[PostJSONPayload, map[string]string, map[string][]string]{
				Method: "POST",
				Path:   "/postJsonPayload",
				Body:   PostJSONPayload{Foo: "bar"},
			},
		},
		{name: "OptionalQueryParameter",
			verb:      "optionalQuery",
			method:    "GET",
			path:      "/optionalQuery",
			routePath: "/optionalQuery",
			query: map[string][]string{
				"foo": {"bar"},
			},
			expected: HTTPRequest[map[string]string, map[string][]string, QueryParameterRequest]{
				Method: "GET",
				Path:   "/optionalQuery",
				Query: QueryParameterRequest{
					Foo: ftl.Some("bar"),
				},
			},
		},
		{name: "PathParameterDecoding",
			verb:      "getPath",
			method:    "GET",
			path:      "/getPath/bob",
			routePath: "/getPath/{username}",
			expected: HTTPRequest[ftl.Unit, PathParameterRequest, map[string][]string]{
				Method: "GET",
				Path:   "/getPath/bob",
				PathParameters: PathParameterRequest{
					Username: "bob",
				},
			},
		},
		{name: "GetById",
			verb:      "getById",
			method:    "GET",
			path:      "/getbyid/100",
			routePath: "/getbyid/{id}",
			expected: HTTPRequest[ftl.Unit, int, ftl.Unit]{
				Method:         "GET",
				Path:           "/getbyid/100",
				PathParameters: 100,
			},
		},
		{name: "MapQuery",
			verb:      "mapQuery",
			method:    "GET",
			path:      "/mapQuery",
			routePath: "/mapQuery",
			query: map[string][]string{
				"alias": {"value"},
			},
			expected: HTTPRequest[ftl.Unit, ftl.Unit, map[string]string]{
				Method: "GET",
				Path:   "/mapQuery",
				Query:  map[string]string{"alias": "value"},
			},
		},
		{name: "MultiMapQuery",
			verb:      "multiMapQuery",
			method:    "GET",
			path:      "/multiMapQuery",
			routePath: "/multiMapQuery",
			query: map[string][]string{
				"alias": {"value"},
			},
			expected: HTTPRequest[ftl.Unit, ftl.Unit, map[string][]string]{
				Method: "GET",
				Path:   "/multiMapQuery",
				Query:  map[string][]string{"alias": []string{"value"}},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if test.body == nil {
				test.body = obj{}
			}
			body, err := encoding.Marshal(test.body)
			assert.NoError(t, err)
			requestURL := "http://127.0.0.1" + test.path
			if test.query != nil {
				requestURL += "?" + test.query.Encode()
			}
			r, err := http.NewRequest(test.method, requestURL, bytes.NewReader(body)) //nolint:noctx
			assert.NoError(t, err)
			requestBody, err := BuildRequestBody(&dal.IngressRoute{
				Path:   test.routePath,
				Module: "test",
				Verb:   test.verb,
			}, r, sch)
			if test.err != "" {
				assert.EqualError(t, err, test.err)
				return
			}
			assert.NoError(t, err)
			actualrv := reflect.New(reflect.TypeOf(test.expected))
			actual := actualrv.Interface()
			err = encoding.Unmarshal(requestBody, actual)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, actualrv.Elem().Interface(), assert.OmitEmpty())
		})
	}
}
