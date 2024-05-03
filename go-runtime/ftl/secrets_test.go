package ftl

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/go-runtime/modulecontext"
	"github.com/TBD54566975/ftl/internal/log"
)

func TestSecret(t *testing.T) {
	type C struct {
		One string
		Two string
	}

	ctx := log.ContextWithNewDefaultLogger(context.Background())

	data, err := json.Marshal(C{"one", "two"})
	assert.NoError(t, err)

	moduleCtx := modulecontext.New("test").Update(
		map[string][]byte{},
		map[string][]byte{
			"test": data,
		},
		map[string]modulecontext.Database{},
	)
	ctx = moduleCtx.ApplyToContext(ctx)

	secret := Secret[C]("test")
	assert.Equal(t, C{"one", "two"}, secret.Get(ctx))
}
