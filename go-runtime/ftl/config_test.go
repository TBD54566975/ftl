package ftl

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/go-runtime/internal"
	"github.com/block/ftl/internal/deploymentcontext"
	"github.com/block/ftl/internal/log"
	. "github.com/block/ftl/internal/testutils/modulecontext"
)

func TestConfig(t *testing.T) {
	type C struct {
		One string
		Two string
	}

	ctx := log.ContextWithNewDefaultLogger(context.Background())

	data, err := json.Marshal(C{"one", "two"})
	assert.NoError(t, err)

	moduleCtx := deploymentcontext.NewBuilder("test").AddConfigs(map[string][]byte{"test": data}).Build()
	ctx = internal.WithContext(ctx, internal.New(MakeDynamic(ctx, moduleCtx)))

	config := Config[C]{Ref: reflection.Ref{Module: "test", Name: "test"}}
	assert.Equal(t, C{"one", "two"}, config.Get(ctx))
}
