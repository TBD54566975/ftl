//go:build integration

package scaling_test

import (
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"

	in "github.com/TBD54566975/ftl/internal/integration"
)

func TestLocalScaling(t *testing.T) {
	in.Run(t,
		in.CopyModule("echo"),
		in.Deploy("echo"),
		in.Call("echo", "echo", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Hello, Bob!!!", response)
		}),
		in.EditFile("echo", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "Hello", "Bye"))
		}, "echo.go"),

		in.Deploy("echo"),
		in.Call("echo", "echo", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Bye, Bob!!!", response)
		}),
	)
}
