package ftl

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/go-runtime/internal"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/modulecontext"
)

func TestConfig(t *testing.T) {
	type C struct {
		One string
		Two string
	}

	ctx := log.ContextWithNewDefaultLogger(context.Background())

	data, err := json.Marshal(C{"one", "two"})
	assert.NoError(t, err)

	ctx = internal.WithContext(ctx, internal.New(modulecontext.NewBuilder("test").AddConfigs(map[string][]byte{"test": data}).Build()))

	config := Config[C]("test")
	assert.Equal(t, C{"one", "two"}, config.Get(ctx))
}
