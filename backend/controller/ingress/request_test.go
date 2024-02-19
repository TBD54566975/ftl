package ingress

import (
	"bytes"
	"encoding/json"
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
type HTTPRequest[Body any] struct {
	Body           Body
	Headers        map[string][]string `json:"headers,omitempty"`
	Method         string
	Path           string
	PathParameters map[string]string   `json:"pathParameters,omitempty"`
	Query          map[string][]string `json:"query,omitempty"`
}

func TestBuildRequestBody(t *testing.T) {
	sch, err := schema.ParseString("test", `
		module test {
			data AliasRequest {
				aliased String alias json "alias"
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

			verb getAlias(HttpRequest<AliasRequest>) HttpResponse<Empty, Empty>
				ingress http GET /getAlias

			verb getPath(HttpRequest<PathParameterRequest>) HttpResponse<Empty, Empty>
				ingress http GET /getPath/{username}

			verb postMissingTypes(HttpRequest<MissingTypes>) HttpResponse<Empty, Empty>
				ingress http POST /postMissingTypes

			verb postJsonPayload(HttpRequest<JsonPayload>) HttpResponse<Empty, Empty>
				ingress http POST /postJsonPayload
		}
	`)
	assert.NoError(t, err)
	for _, test := range []struct {
		name     string
		verb     string
		method   string
		path     string
		query    url.Values
		body     obj
		expected any
		err      string
	}{
		{name: "UnknownVerb",
			verb: "unknown",
			err:  `unknown verb "unknown"`},
		{name: "UnknownModule",
			verb: "unknown",
			err:  `unknown verb "unknown"`},
		//FIXME: Query parameter decoding doesn't work?
		//
		// {name: "QueryParameterDecoding",
		// 	verb:   "getAlias",
		// 	method: "GET",
		// 	path:   "/getAlias",
		// 	query: map[string][]string{
		// 		"alias": {"value"},
		// 	},
		// 	expected: HTTPRequest[AliasRequest]{
		// 		Method: "GET",
		// 		Path:   "/getAlias",
		// 		Query: map[string][]string{
		// 			"alias": {"value"},
		// 		},
		// 		Body: AliasRequest{
		// 			Aliased: "value",
		// 		},
		// 	},
		// },
		{name: "AllowMissingFieldTypes",
			verb:   "postMissingTypes",
			method: "POST",
			path:   "/postMissingTypes",
			expected: HTTPRequest[MissingTypes]{
				Method: "POST",
				Path:   "/postMissingTypes",
				Body:   MissingTypes{},
			},
		},
		{name: "JSONPayload",
			verb:   "postJsonPayload",
			method: "POST",
			path:   "/postJsonPayload",
			body:   obj{"foo": "bar"},
			expected: HTTPRequest[PostJSONPayload]{
				Method: "POST",
				Path:   "/postJsonPayload",
				Body:   PostJSONPayload{Foo: "bar"},
			},
		},
		// FIXME: Path parameters don't seem to be interpolated into body fields?
		//
		// {name: "PathParameterDecoding",
		// 	verb:  "getPath",
		// 	method: "GET",
		// 	path:   "/getPath/bob",
		// 	expected: HTTPRequest[PathParameterRequest]{
		// 		Method: "GET",
		// 		Path:   "/getPath/bob",
		// 		PathParameters: map[string]string{
		// 			"username": "bob",
		// 		},
		// 		Body: PathParameterRequest{
		// 			Username: "bob",
		// 		},
		// 	},
		// },
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
				Path:   test.path,
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
			err = json.Unmarshal(requestBody, actual)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, actualrv.Elem().Interface(), assert.OmitEmpty())
		})
	}
}
