//go:build integration

package ingress

import (
	"net/http"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/repr"

	"github.com/TBD54566975/ftl/integration"
)

func TestHttpIngress(t *testing.T) {
	integration.Run(t, "",
		integration.CopyModule("httpingress"),
		integration.Deploy("httpingress"),
		integration.HttpCall(http.MethodGet, "/users/123/posts/456", integration.JsonData(t, obj{}), func(t testing.TB, resp *integration.HTTPResponse) {
			assert.Equal(t, 200, resp.Status)
			assert.Equal(t, []string{"Header from FTL"}, resp.Headers["Get"])
			assert.Equal(t, []string{"application/json; charset=utf-8"}, resp.Headers["Content-Type"])

			message, ok := resp.JsonBody["msg"].(string)
			assert.True(t, ok, "msg is not a string: %s", repr.String(resp.JsonBody))
			assert.Equal(t, "UserID: 123, PostID: 456", message)

			nested, ok := resp.JsonBody["nested"].(map[string]any)
			assert.True(t, ok, "nested is not a map: %s", repr.String(resp.JsonBody))
			goodStuff, ok := nested["good_stuff"].(string)
			assert.True(t, ok, "good_stuff is not a string: %s", repr.String(resp.JsonBody))
			assert.Equal(t, "This is good stuff", goodStuff)
		}),
		integration.HttpCall(http.MethodPost, "/users", integration.JsonData(t, obj{"userId": 123, "postId": 345}), func(t testing.TB, resp *integration.HTTPResponse) {
			assert.Equal(t, 201, resp.Status)
			assert.Equal(t, []string{"Header from FTL"}, resp.Headers["Post"])
			success, ok := resp.JsonBody["success"].(bool)
			assert.True(t, ok, "success is not a bool: %s", repr.String(resp.JsonBody))
			assert.True(t, success)
		}),
		// contains aliased field
		integration.HttpCall(http.MethodPost, "/users", integration.JsonData(t, obj{"user_id": 123, "postId": 345}), func(t testing.TB, resp *integration.HTTPResponse) {
			assert.Equal(t, 201, resp.Status)
		}),
		integration.HttpCall(http.MethodPut, "/users/123", integration.JsonData(t, obj{"postId": "346"}), func(t testing.TB, resp *integration.HTTPResponse) {
			assert.Equal(t, 200, resp.Status)
			assert.Equal(t, []string{"Header from FTL"}, resp.Headers["Put"])
			assert.Equal(t, map[string]any{}, resp.JsonBody)
		}),
		integration.HttpCall(http.MethodDelete, "/users/123", integration.JsonData(t, obj{}), func(t testing.TB, resp *integration.HTTPResponse) {
			assert.Equal(t, 200, resp.Status)
			assert.Equal(t, []string{"Header from FTL"}, resp.Headers["Delete"])
			assert.Equal(t, map[string]any{}, resp.JsonBody)
		}),

		integration.HttpCall(http.MethodGet, "/html", integration.JsonData(t, obj{}), func(t testing.TB, resp *integration.HTTPResponse) {
			assert.Equal(t, 200, resp.Status)
			assert.Equal(t, []string{"text/html; charset=utf-8"}, resp.Headers["Content-Type"])
			assert.Equal(t, "<html><body><h1>HTML Page From FTL 🚀!</h1></body></html>", string(resp.BodyBytes))
		}),

		integration.HttpCall(http.MethodPost, "/bytes", []byte("Hello, World!"), func(t testing.TB, resp *integration.HTTPResponse) {
			assert.Equal(t, 200, resp.Status)
			assert.Equal(t, []string{"application/octet-stream"}, resp.Headers["Content-Type"])
			assert.Equal(t, []byte("Hello, World!"), resp.BodyBytes)
		}),

		integration.HttpCall(http.MethodGet, "/empty", nil, func(t testing.TB, resp *integration.HTTPResponse) {
			assert.Equal(t, 200, resp.Status)
			assert.Equal(t, nil, resp.Headers["Content-Type"])
			assert.Equal(t, nil, resp.BodyBytes)
		}),

		integration.HttpCall(http.MethodGet, "/string", []byte("Hello, World!"), func(t testing.TB, resp *integration.HTTPResponse) {
			assert.Equal(t, 200, resp.Status)
			assert.Equal(t, []string{"text/plain; charset=utf-8"}, resp.Headers["Content-Type"])
			assert.Equal(t, []byte("Hello, World!"), resp.BodyBytes)
		}),

		integration.HttpCall(http.MethodGet, "/int", []byte("1234"), func(t testing.TB, resp *integration.HTTPResponse) {
			assert.Equal(t, 200, resp.Status)
			assert.Equal(t, []string{"text/plain; charset=utf-8"}, resp.Headers["Content-Type"])
			assert.Equal(t, []byte("1234"), resp.BodyBytes)
		}),
		integration.HttpCall(http.MethodGet, "/float", []byte("1234.56789"), func(t testing.TB, resp *integration.HTTPResponse) {
			assert.Equal(t, 200, resp.Status)
			assert.Equal(t, []string{"text/plain; charset=utf-8"}, resp.Headers["Content-Type"])
			assert.Equal(t, []byte("1234.56789"), resp.BodyBytes)
		}),
		integration.HttpCall(http.MethodGet, "/bool", []byte("true"), func(t testing.TB, resp *integration.HTTPResponse) {
			assert.Equal(t, 200, resp.Status)
			assert.Equal(t, []string{"text/plain; charset=utf-8"}, resp.Headers["Content-Type"])
			assert.Equal(t, []byte("true"), resp.BodyBytes)
		}),
		integration.HttpCall(http.MethodGet, "/error", nil, func(t testing.TB, resp *integration.HTTPResponse) {
			assert.Equal(t, 500, resp.Status)
			assert.Equal(t, []string{"text/plain; charset=utf-8"}, resp.Headers["Content-Type"])
			assert.Equal(t, []byte("Error from FTL"), resp.BodyBytes)
		}),
		integration.HttpCall(http.MethodGet, "/array/string", integration.JsonData(t, []string{"hello", "world"}), func(t testing.TB, resp *integration.HTTPResponse) {
			assert.Equal(t, 200, resp.Status)
			assert.Equal(t, []string{"application/json; charset=utf-8"}, resp.Headers["Content-Type"])
			assert.Equal(t, integration.JsonData(t, []string{"hello", "world"}), resp.BodyBytes)
		}),
		integration.HttpCall(http.MethodPost, "/array/data", integration.JsonData(t, []obj{{"item": "a"}, {"item": "b"}}), func(t testing.TB, resp *integration.HTTPResponse) {
			assert.Equal(t, 200, resp.Status)
			assert.Equal(t, []string{"application/json; charset=utf-8"}, resp.Headers["Content-Type"])
			assert.Equal(t, integration.JsonData(t, []obj{{"item": "a"}, {"item": "b"}}), resp.BodyBytes)
		}),
		integration.HttpCall(http.MethodGet, "/typeenum", integration.JsonData(t, obj{"name": "A", "value": "hello"}), func(t testing.TB, resp *integration.HTTPResponse) {
			assert.Equal(t, 200, resp.Status)
			assert.Equal(t, []string{"application/json; charset=utf-8"}, resp.Headers["Content-Type"])
			assert.Equal(t, integration.JsonData(t, obj{"name": "A", "value": "hello"}), resp.BodyBytes)
		}),
	)
}
