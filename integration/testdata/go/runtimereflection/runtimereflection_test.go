package runtimereflection

import (
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

func TestTypeRef(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.Equal(t, ftl.Ref{Module: "runtimereflection", Name: "echoRequest"}, ftl.TypeRef[EchoRequest]())
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.Panics(t, func() {
			ftl.TypeRef[int]()
		})
	})
}

func TestVerbRef(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.Equal(t, ftl.Ref{Module: "runtimereflection", Name: "echo"}, ftl.FuncRef(Echo))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.Panics(t, func() {
			ftl.FuncRef(time.Now)
		})
	})
}

func TestModule(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.Equal(t, "runtimereflection", ftl.Module())
	})
}
