package runtimereflection

import (
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
)

func TestTypeRef(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.Equal(t, reflection.Ref{Module: "runtimereflection", Name: "echoRequest"}, reflection.TypeRef[EchoRequest]())
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.Panics(t, func() {
			reflection.TypeRef[int]()
		})
	})
}

func TestVerbRef(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.Equal(t, reflection.Ref{Module: "runtimereflection", Name: "echo"}, reflection.FuncRef(Echo))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.Panics(t, func() {
			reflection.FuncRef(time.Now)
		})
	})
}

func TestModule(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.Equal(t, "runtimereflection", reflection.Module())
	})
}
