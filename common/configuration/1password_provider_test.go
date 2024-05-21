//go:build 1password

// 1password needs to be running and have a vault named "ftl test".
//
// These tests will clean up before and after itself.

package configuration

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
)

const vault = "ftl test"
const module = "test.module"

func clean(ctx context.Context) bool {
	args := []string{"item", "delete", "--vault", vault, module}
	_, err := exec.Capture(ctx, ".", "op", args...)
	return err == nil
}

func Test1PasswordProvider(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	// OK to fail the initial clean.
	_ = clean(ctx)
	defer func() {
		if !clean(ctx) {
			t.Fatal("clean failed")
		}
	}()

	_, err := getItem(ctx, vault, Ref{Name: module})
	assert.Error(t, err)

	err = createItem(ctx, vault, Ref{Name: module}, "hunter1")
	assert.NoError(t, err)

	value, err := getItem(ctx, vault, Ref{Name: module})
	assert.NoError(t, err)
	secret, ok := value.value("password")
	assert.True(t, ok)
	assert.Equal(t, "hunter1", secret)

	err = editItem(ctx, vault, Ref{Name: module}, "hunter2")
	assert.NoError(t, err)

	value, err = getItem(ctx, vault, Ref{Name: module})
	assert.NoError(t, err)
	secret, ok = value.value("password")
	assert.True(t, ok)
	assert.Equal(t, "hunter2", secret)
}
