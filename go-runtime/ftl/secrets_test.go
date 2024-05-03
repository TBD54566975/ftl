package ftl

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/go-runtime/modulecontext"
	"github.com/TBD54566975/ftl/internal/log"
)

func TestSecret(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	moduleCtx := modulecontext.New()
	ctx = moduleCtx.ApplyToContext(ctx)

	type C struct {
		One string
		Two string
	}
	secret := Secret[C]("test")
	assert.NoError(t, moduleCtx.SetSecret("test", C{"one", "two"}))
	assert.Equal(t, C{"one", "two"}, secret.Get(ctx))
}
