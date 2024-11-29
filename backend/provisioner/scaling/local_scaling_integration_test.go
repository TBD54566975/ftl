//go:build integration

package scaling_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/atomic"

	in "github.com/TBD54566975/ftl/internal/integration"
)

func TestLocalScaling(t *testing.T) {
	failure := atomic.Value[error]{}
	done := atomic.Value[bool]{}
	routineStopped := sync.WaitGroup{}
	routineStopped.Add(1)
	done.Store(false)
	in.Run(t,
		in.CopyModule("echo"),
		in.Deploy("echo"),
		in.Call("echo", "echo", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Hello, Bob!!!", response)
		}),
		in.EditFile("echo", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "Hello", "Bye"))
		}, "echo.go"),
		func(t testing.TB, ic in.TestContext) {
			// Hit the verb constantly to test rolling updates.
			go func() {
				defer routineStopped.Done()
				for !done.Load() {
					in.Call("echo", "echo", "Bob", func(t testing.TB, response string) {
						if !strings.Contains(response, "Bob") {
							failure.Store(fmt.Errorf("unexpected response: %s", response))
							return
						}
					})(t, ic)
				}
			}()
		},
		in.Deploy("echo"),
		in.Call("echo", "echo", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Bye, Bob!!!", response)
		}),
		func(t testing.TB, ic in.TestContext) {
			done.Store(true)
			routineStopped.Wait()
			err := failure.Load()
			assert.NoError(t, err)
		},
	)
}
