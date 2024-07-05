package ftltest

import (
	"context"
	"fmt"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/internal"
	"github.com/TBD54566975/ftl/internal/log"
)

func PanicsWithErr(t testing.TB, substr string, fn func()) {
	t.Helper()
	defer func() {
		err := recover()
		if err == nil {
			t.Fatal("Expected panic, but got nil")
		}

		errStr := fmt.Sprintf("%v", err)
		assert.Contains(t, errStr, substr, "Expected panic message to contain %q, but got %q", substr, errStr)
	}()
	fn()
}

func TestFtlTestProjectNotLoadedInContext(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	ctx = internal.WithContext(ctx, newFakeFTL(ctx, t))

	// This should panic suggesting to use ftltest.WithDefaultProjectFile()
	PanicsWithErr(t, "ftltest.WithDefaultProjectFile()", func() {
		_ = ftl.Secret[string]("moo").Get(ctx)
	})
	PanicsWithErr(t, "ftltest.WithDefaultProjectFile()", func() {
		_ = ftl.Config[string]("moo").Get(ctx)
	})
}
