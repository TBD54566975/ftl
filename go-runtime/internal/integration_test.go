//go:build integration

package internal

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	. "github.com/TBD54566975/ftl/internal/integration"
)

func TestRealMap(t *testing.T) {
	Run(t,
		CopyModule("mapper"),
		Deploy("mapper"),
		Call("mapper", "get", Obj{}, func(t testing.TB, response Obj) {
			assert.Equal(t, Obj{"underlyingCounter": 1.0, "mapCounter": 1.0, "mapped": "0"}, response)
		}),
		Call("mapper", "get", Obj{}, func(t testing.TB, response Obj) {
			assert.Equal(t, Obj{"underlyingCounter": 2.0, "mapCounter": 1.0, "mapped": "0"}, response)
		}),
		Call[Obj, Obj]("mapper", "inc", Obj{}, nil),
		Call("mapper", "get", Obj{}, func(t testing.TB, response Obj) {
			assert.Equal(t, Obj{"underlyingCounter": 3.0, "mapCounter": 2.0, "mapped": "1"}, response)
		}),
	)
}
