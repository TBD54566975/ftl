//go:build integration

package ingress_test

import (
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/repr"

	in "github.com/block/ftl/internal/integration"
)

func TestHttpIngress(t *testing.T) {
	in.Run(t,
		in.WithLanguages("go", "java"),
		in.CopyModule("httpingress"),
		in.Deploy("httpingress"),
		in.SubTests(
			in.SubTest{Name: "GetWithPathParams", Action: in.HttpCall(http.MethodGet, "/users/123/posts/456", nil, nil, func(t testing.TB, resp *in.HTTPResponse) {
				assert.Equal(t, 200, resp.Status)
				assert.Equal(t, []string{"Header from FTL"}, resp.Headers["Get"])
				expectContentType(t, resp, "application/json;charset=utf-8")

				message, ok := resp.JsonBody["msg"].(string)
				assert.True(t, ok, "msg is not a string: %s", repr.String(resp.JsonBody))
				assert.Equal(t, "UserID: 123, PostID: 456", message)

				nested, ok := resp.JsonBody["nested"].(map[string]any)
				assert.True(t, ok, "nested is not a map: %s", repr.String(resp.JsonBody))
				goodStuff, ok := nested["good_stuff"].(string)
				assert.True(t, ok, "good_stuff is not a string: %s", repr.String(resp.JsonBody))
				assert.Equal(t, "This is good stuff", goodStuff)
			})},
			in.SubTest{Name: "GetWithNumericQueryParams", Action: in.HttpCall(http.MethodGet, "/getquery?userId=123&postId=456", nil, nil, func(t testing.TB, resp *in.HTTPResponse) {
				assert.Equal(t, 200, resp.Status)
				assert.Equal(t, []string{"Header from FTL"}, resp.Headers["Get"])
				expectContentType(t, resp, "application/json;charset=utf-8")

				message, ok := resp.JsonBody["msg"].(string)
				assert.True(t, ok, "msg is not a string: %s", repr.String(resp.JsonBody))
				assert.Equal(t, "UserID: 123, PostID: 456", message)

				nested, ok := resp.JsonBody["nested"].(map[string]any)
				assert.True(t, ok, "nested is not a map: %s", repr.String(resp.JsonBody))
				goodStuff, ok := nested["good_stuff"].(string)
				assert.True(t, ok, "good_stuff is not a string: %s", repr.String(resp.JsonBody))
				assert.Equal(t, "This is good stuff", goodStuff)
			})},
			in.SubTest{Name: "PostUsers", Action: in.HttpCall(http.MethodPost, "/users", nil, in.JsonData(t, in.Obj{"userId": 123, "postId": 345}), func(t testing.TB, resp *in.HTTPResponse) {
				assert.Equal(t, 201, resp.Status)
				assert.Equal(t, []string{"Header from FTL"}, resp.Headers["Post"])
				success, ok := resp.JsonBody["success"].(bool)
				assert.True(t, ok, "success is not a bool: %s", repr.String(resp.JsonBody))
				assert.True(t, success)
			})},
			// contains aliased field
			in.SubTest{Name: "PostUsersAliased", Action: in.HttpCall(http.MethodPost, "/users", nil, in.JsonData(t, in.Obj{"user_id": 123, "postId": 345}), func(t testing.TB, resp *in.HTTPResponse) {
				assert.Equal(t, 201, resp.Status)
			})},
			in.SubTest{Name: "PutUsers", Action: in.HttpCall(http.MethodPut, "/users/123", nil, in.JsonData(t, in.Obj{"postId": "346"}), func(t testing.TB, resp *in.HTTPResponse) {
				assert.Equal(t, 200, resp.Status)
				assert.Equal(t, []string{"Header from FTL"}, resp.Headers["Put"])
				assert.Equal(t, map[string]any{}, resp.JsonBody)
			})},
			in.SubTest{Name: "DeleteUsers", Action: in.HttpCall(http.MethodDelete, "/users/123", nil, nil, func(t testing.TB, resp *in.HTTPResponse) {
				assert.Equal(t, 200, resp.Status)
				assert.Equal(t, []string{"Header from FTL"}, resp.Headers["Delete"])
				assert.Equal(t, map[string]any{}, resp.JsonBody)
			})},

			in.SubTest{Name: "GetQueryParams", Action: in.HttpCall(http.MethodGet, "/queryparams?foo=bar", nil, nil, func(t testing.TB, resp *in.HTTPResponse) {
				assert.Equal(t, 200, resp.Status)
				assert.Equal(t, "bar", string(resp.BodyBytes))
			})},

			in.SubTest{Name: "GetMissingQueryParams", Action: in.HttpCall(http.MethodGet, "/queryparams", nil, nil, func(t testing.TB, resp *in.HTTPResponse) {
				assert.Equal(t, 200, resp.Status)
				assert.Equal(t, "No value", string(resp.BodyBytes))
			})},

			in.SubTest{Name: "GetHTML", Action: in.HttpCall(http.MethodGet, "/html", nil, nil, func(t testing.TB, resp *in.HTTPResponse) {
				assert.Equal(t, 200, resp.Status)
				expectContentType(t, resp, "text/html;charset=utf-8")
				assert.Equal(t, "<html><body><h1>HTML Page From FTL 🚀!</h1></body></html>", string(resp.BodyBytes))
			})},

			in.SubTest{Name: "PostBytes", Action: in.HttpCall(http.MethodPost, "/bytes", nil, []byte("Hello, World!"), func(t testing.TB, resp *in.HTTPResponse) {
				assert.Equal(t, 200, resp.Status)
				expectContentType(t, resp, "application/octet-stream")
				assert.Equal(t, []byte("Hello, World!"), resp.BodyBytes)
			})},

			in.SubTest{Name: "GetEmpty", Action: in.HttpCall(http.MethodGet, "/empty", nil, nil, func(t testing.TB, resp *in.HTTPResponse) {
				assert.Equal(t, 200, resp.Status)
				assert.Equal(t, nil, resp.Headers["Content-Type"])
				assert.Equal(t, nil, resp.BodyBytes)
			})},

			in.SubTest{Name: "PostString", Action: in.HttpCall(http.MethodPost, "/string", nil, []byte("Hello, World!"), func(t testing.TB, resp *in.HTTPResponse) {
				assert.Equal(t, 200, resp.Status)
				expectContentType(t, resp, "text/plain;charset=utf-8")
				assert.Equal(t, []byte("Hello, World!"), resp.BodyBytes)
			})},
			in.SubTest{Name: "GetError", Action: in.HttpCall(http.MethodGet, "/error", nil, nil, func(t testing.TB, resp *in.HTTPResponse) {
				assert.Equal(t, 500, resp.Status)
				expectContentType(t, resp, "text/plain;charset=utf-8")
				assert.True(t, strings.Contains(string(resp.BodyBytes), "Error"))
			})},
			in.SubTest{Name: "PostArrayString", Action: in.HttpCall(http.MethodPost, "/array/string", nil, in.JsonData(t, []string{"hello", "world"}), func(t testing.TB, resp *in.HTTPResponse) {
				assert.Equal(t, 200, resp.Status)
				expectContentType(t, resp, "application/json;charset=utf-8")
				assert.Equal(t, in.JsonData(t, []string{"hello", "world"}), resp.BodyBytes)
			})},
			in.SubTest{Name: "PostArrayJSON", Action: in.HttpCall(http.MethodPost, "/array/data", nil, in.JsonData(t, []in.Obj{{"item": "a"}, {"item": "b"}}), func(t testing.TB, resp *in.HTTPResponse) {
				assert.Equal(t, 200, resp.Status)
				expectContentType(t, resp, "application/json;charset=utf-8")
				assert.Equal(t, in.JsonData(t, []in.Obj{{"item": "a"}, {"item": "b"}}), resp.BodyBytes)
			})},
			// CORS preflight request without CORS middleware enabled
			in.SubTest{Name: "TestOptionsRequest", Action: in.HttpCall(http.MethodOptions, "/typeenum", map[string][]string{
				"Origin":                         {"http://localhost:8892"},
				"Access-Control-Request-Method":  {"GET"},
				"Access-Control-Request-Headers": {"x-forwarded-capabilities"},
			}, nil, func(t testing.TB, resp *in.HTTPResponse) {
				// should not return access control headers because we have not set up cors in this controller
				assert.Equal(t, nil, resp.Headers["Access-Control-Allow-Origin"])
				assert.Equal(t, nil, resp.Headers["Access-Control-Allow-Methods"])
				assert.Equal(t, nil, resp.Headers["Access-Control-Allow-Headers"])
			})},
		),
		in.IfLanguage("go", in.SubTests(
			// Double, Int etc work in java with JSON encoding, but test/plain is not implemented yet
			in.SubTest{Name: "PostInt", Action: in.HttpCall(http.MethodPost, "/int", nil, []byte("1234"), func(t testing.TB, resp *in.HTTPResponse) {
				assert.Equal(t, 200, resp.Status)
				expectContentType(t, resp, "text/plain;charset=utf-8")
				assert.Equal(t, []byte("1234"), resp.BodyBytes)
			})},
			in.SubTest{Name: "PostFloat", Action: in.HttpCall(http.MethodPost, "/float", nil, []byte("1234.56789"), func(t testing.TB, resp *in.HTTPResponse) {
				assert.Equal(t, 200, resp.Status)
				expectContentType(t, resp, "text/plain;charset=utf-8")
				assert.Equal(t, []byte("1234.56789"), resp.BodyBytes)
			})},
			in.SubTest{Name: "PostBool", Action: in.HttpCall(http.MethodPost, "/bool", nil, []byte("true"), func(t testing.TB, resp *in.HTTPResponse) {
				assert.Equal(t, 200, resp.Status)
				expectContentType(t, resp, "text/plain;charset=utf-8")
				assert.Equal(t, []byte("true"), resp.BodyBytes)
			})},
			in.SubTest{Name: "PostTypeEnum", Action: in.HttpCall(http.MethodPost, "/typeenum", nil, in.JsonData(t, in.Obj{"name": "A", "value": "hello"}), func(t testing.TB, resp *in.HTTPResponse) {
				assert.Equal(t, 200, resp.Status)
				expectContentType(t, resp, "application/json;charset=utf-8")
				assert.Equal(t, in.JsonData(t, in.Obj{"name": "A", "value": "hello"}), resp.BodyBytes)
			})},
			in.SubTest{Name: "TestExternalType", Action: in.HttpCall(http.MethodPost, "/external", nil, in.JsonData(t, in.Obj{"message": "hello"}), func(t testing.TB, resp *in.HTTPResponse) {
				assert.Equal(t, 200, resp.Status)
				expectContentType(t, resp, "application/json;charset=utf-8")
				assert.Equal(t, in.JsonData(t, in.Obj{"message": "hello"}), resp.BodyBytes)
			})},
			in.SubTest{Name: "TestExternalType2", Action: in.HttpCall(http.MethodPost, "/external2", nil, in.JsonData(t, in.Obj{"message": "hello"}), func(t testing.TB, resp *in.HTTPResponse) {
				assert.Equal(t, 200, resp.Status)
				expectContentType(t, resp, "application/json;charset=utf-8")
				assert.Equal(t, in.JsonData(t, in.Obj{"Message": "hello"}), resp.BodyBytes)
			})},
			// not lenient
			in.SubTest{Name: "TestExtraFieldStrict", Action: in.HttpCall(http.MethodPost, "/users", nil, in.JsonData(t, in.Obj{"user_id": 123, "postId": 345, "extra": "blah"}), func(t testing.TB, resp *in.HTTPResponse) {
				assert.Equal(t, 400, resp.Status)
			})},
			// lenient
			in.SubTest{Name: "TestExtraFieldLenient", Action: in.HttpCall(http.MethodPost, "/lenient", nil, in.JsonData(t, in.Obj{"user_id": 123, "postId": 345, "extra": "blah"}), func(t testing.TB, resp *in.HTTPResponse) {
				assert.Equal(t, 201, resp.Status)
			})},
		)),
	)
}

func expectContentType(t testing.TB, resp *in.HTTPResponse, expected string) {
	t.Helper()
	headers := resp.Headers["Content-Type"]
	for k, v := range headers {
		headers[k] = strings.ReplaceAll(strings.ToLower(v), " ", "")
	}
	expected = strings.ReplaceAll(strings.ToLower(expected), " ", "")
	assert.Equal(t, []string{expected}, headers)
}

// Run with CORS enabled via FTL_CONTROLLER_ALLOW_ORIGIN and FTL_CONTROLLER_ALLOW_HEADERS
// This test is similar to TestHttpIngress above with the addition of CORS enabled in the controller.
func TestHttpIngressWithCors(t *testing.T) {
	os.Setenv("FTL_INGRESS_ALLOW_ORIGIN", "http://localhost:8892")
	os.Setenv("FTL_INGRESS_ALLOW_HEADERS", "x-forwarded-capabilities")
	in.Run(t,
		in.CopyModule("httpingress"),
		in.Deploy("httpingress"),
		// A correct CORS preflight request
		in.HttpCall(http.MethodOptions, "/typeenum", map[string][]string{
			"Origin":                         {"http://localhost:8892"},
			"Access-Control-Request-Method":  {"GET"},
			"Access-Control-Request-Headers": {"x-forwarded-capabilities"},
		}, nil, func(t testing.TB, resp *in.HTTPResponse) {
			assert.Equal(t, []string{"http://localhost:8892"}, resp.Headers["Access-Control-Allow-Origin"])
			assert.Equal(t, []string{"GET"}, resp.Headers["Access-Control-Allow-Methods"])
			assert.Equal(t, []string{"x-forwarded-capabilities"}, resp.Headers["Access-Control-Allow-Headers"])
		}),
		// Not allowed headers
		in.HttpCall(http.MethodOptions, "/typeenum", map[string][]string{
			"Origin":                         {"http://localhost:8892"},
			"Access-Control-Request-Method":  {"GET"},
			"Access-Control-Request-Headers": {"moo"},
		}, nil, func(t testing.TB, resp *in.HTTPResponse) {
			assert.Equal(t, nil, resp.Headers["Access-Control-Allow-Origin"])
			assert.Equal(t, nil, resp.Headers["Access-Control-Allow-Methods"])
			assert.Equal(t, nil, resp.Headers["Access-Control-Allow-Headers"])
		}),
		// Not allowed origin
		in.HttpCall(http.MethodOptions, "/typeenum", map[string][]string{
			"Origin":                         {"http://localhost:4444"},
			"Access-Control-Request-Method":  {"GET"},
			"Access-Control-Request-Headers": {"x-forwarded-capabilities"},
		}, nil, func(t testing.TB, resp *in.HTTPResponse) {
			assert.Equal(t, nil, resp.Headers["Access-Control-Allow-Origin"])
			assert.Equal(t, nil, resp.Headers["Access-Control-Allow-Methods"])
			assert.Equal(t, nil, resp.Headers["Access-Control-Allow-Headers"])
		}),
	)
}
