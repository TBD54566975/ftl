package ftl

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/go-runtime/modulecontext"
	"github.com/TBD54566975/ftl/internal/log"
)

func TestConfig(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	moduleCtx := modulecontext.New("test")
	ctx = moduleCtx.ApplyToContext(ctx)

	type C struct {
		One string
		Two string
	}
	config := Config[C]("test")
	assert.NoError(t, moduleCtx.SetConfig("test", C{"one", "two"}))
	assert.Equal(t, C{"one", "two"}, config.Get(ctx))
}
