package modulecontext

import (
	"context" //nolint:depguard
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/internal/log"
)

func TestGettingAndSettingFromContext(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	moduleCtx := NewBuilder("test").Build()
	ctx = moduleCtx.MakeDynamic(ctx).ApplyToContext(ctx)
	assert.Equal(t, moduleCtx, FromContext(ctx).CurrentContext(), "module context should be the same when read from context")
}
