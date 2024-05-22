//go:build 1password

// 1password needs to be running and will temporarily make a "ftl-test" vault.
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

const vault = "ftl-test"
const module = "test.module"

func createVault(ctx context.Context) error {
	args := []string{"vault", "create", vault}
	_, err := exec.Capture(ctx, ".", "op", args...)
	return err
}

func clean(ctx context.Context) bool {
	args := []string{"vault", "delete", vault}
	_, err := exec.Capture(ctx, ".", "op", args...)
	return err == nil
}

func Test1PasswordProvider(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	// OK to fail the initial clean.
	_ = clean(ctx)
	t.Cleanup(func() {
		if !clean(ctx) {
			t.Fatal("clean failed")
		}
	})

	err := createVault(ctx)
	assert.NoError(t, err)

	_, err = getItem(ctx, vault, Ref{Name: module})
	assert.Error(t, err)

	var pw1 = []byte("hunter1")
	var pw2 = []byte(`{
	  "user": "root",
	  "password": "hunter🪤"
	}`)

	err = createItem(ctx, vault, Ref{Name: module}, pw1)
	assert.NoError(t, err)

	value, err := getItem(ctx, vault, Ref{Name: module})
	assert.NoError(t, err)
	secret, ok := value.password()
	assert.True(t, ok)
	assert.Equal(t, pw1, secret)

	err = editItem(ctx, vault, Ref{Name: module}, pw2)
	assert.NoError(t, err)

	value, err = getItem(ctx, vault, Ref{Name: module})
	assert.NoError(t, err)
	secret, ok = value.password()
	assert.True(t, ok)
	assert.Equal(t, pw2, secret)
}
