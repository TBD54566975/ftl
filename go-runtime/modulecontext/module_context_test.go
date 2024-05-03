package modulecontext

import (
	"context" //nolint:depguard
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/internal/log"
)

func TestGettingAndSettingFromContext(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	moduleCtx := New("test")
	ctx = moduleCtx.ApplyToContext(ctx)
	assert.Equal(t, moduleCtx, FromContext(ctx), "module context should be the same when read from context")
}
