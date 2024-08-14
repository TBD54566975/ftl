//go:build integration

package ftl_test

import (
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"

	in "github.com/TBD54566975/ftl/integration"

	"github.com/alecthomas/repr"
)

func TestJavaToGoCall(t *testing.T) {
	in.Run(t,
		in.WithJava(),
		in.CopyModule("gomodule"),
		in.CopyDir("javamodule", "javamodule"),
		in.Deploy("gomodule"),
		in.Deploy("javamodule"),
		in.Call("javamodule", "timeVerb", in.Obj{}, func(t testing.TB, response in.Obj) {
			message, ok := response["time"].(string)
			assert.True(t, ok, "time is not a string: %s", repr.String(response))
			result, err := time.Parse(time.RFC3339, message)
			assert.NoError(t, err, "time is not a valid RFC3339 time: %s", message)
			assert.True(t, result.After(time.Now().Add(-time.Minute)), "time is not recent: %s", message)
		}),
		// We call both the go and pass through Java versions
		// To make sure the response is the same
		in.Call("gomodule", "emptyVerb", in.Obj{}, func(t testing.TB, response in.Obj) {
			assert.Equal(t, map[string]any{}, response, "expecting empty response, got %s", repr.String(response))
		}),
		in.Call("javamodule", "emptyVerb", in.Obj{}, func(t testing.TB, response in.Obj) {
			assert.Equal(t, map[string]any{}, response, "expecting empty response, got %s", repr.String(response))
		}),
		in.Call("gomodule", "sinkVerb", "ignored", func(t testing.TB, response in.Obj) {
			assert.Equal(t, map[string]any{}, response, "expecting empty response, got %s", repr.String(response))
		}),
		in.Call("javamodule", "sinkVerb", "ignored", func(t testing.TB, response in.Obj) {
			assert.Equal(t, map[string]any{}, response, "expecting empty response, got %s", repr.String(response))
		}),
		in.Call("gomodule", "sourceVerb", in.Obj{}, func(t testing.TB, response string) {
			assert.Equal(t, "Source Verb", response, "expecting empty response, got %s", response)
		}),
		in.Call("javamodule", "sourceVerb", in.Obj{}, func(t testing.TB, response string) {
			assert.Equal(t, "Source Verb", response, "expecting empty response, got %s", response)
		}),
		in.Fail(
			in.Call("gomodule", "errorEmptyVerb", in.Obj{}, func(t testing.TB, response in.Obj) {
				assert.Equal(t, map[string]any{}, response, "expecting empty response, got %s", repr.String(response))
			}), "verb failed"),
		in.Fail(
			in.Call("gomodule", "errorEmptyVerb", in.Obj{}, func(t testing.TB, response in.Obj) {
				assert.Equal(t, map[string]any{}, response, "expecting empty response, got %s", repr.String(response))
			}), "verb failed"),
	)
}
