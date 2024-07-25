package inner_test

import (
	"testing"

	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/alecthomas/assert/v2"
)

func TestInner(t *testing.T) {
	// make sure that packages within a module are correctly identified as being part of the correct module
	assert.Equal(t, "outer", reflection.Module())
}
