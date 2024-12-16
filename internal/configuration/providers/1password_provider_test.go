//go:build 1password

// 1password needs to be running and will temporarily make a "ftl-test" vault.
//
// These tests will clean up before and after itself.

package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/internal/configuration"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/log"
)

const vault = "ftl-test"

func createVault(ctx context.Context) (string, error) {
	args := []string{
		"vault", "create", vault,
		"--format", "json",
	}
	output, err := exec.Capture(ctx, ".", "op", args...)
	if err != nil {
		return "", err
	}
	var parsed map[string]any
	if err := json.Unmarshal(output, &parsed); err != nil {
		return "", fmt.Errorf("could not decode 1Password create vault response: %w", err)
	}
	id, ok := parsed["id"].(string)
	if !ok {
		return "", fmt.Errorf("could not find id in 1Password create vault response: %w", err)
	}
	return id, nil
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

	vauldId, err := createVault(ctx)
	assert.NoError(t, err)

	provider := OnePassword{
		ProjectName: "unittest",
		Vault:       vauldId,
	}

	_, err = provider.getItem(ctx, vault)
	assert.Error(t, err)

	var pw1 = []byte("hunter1")
	var pw2 = []byte(`{
	  "user": "root",
	  "password": "hun\\ter🪤"
	}`)

	ref := configuration.Ref{Module: optional.Some("mod"), Name: "example"}

	err = provider.createItem(ctx, vault)
	assert.NoError(t, err)

	err = provider.storeSecret(ctx, vault, ref, pw1)
	assert.NoError(t, err)

	item, err := provider.getItem(ctx, vault)
	assert.NoError(t, err)
	secret, ok := item.value(ref)
	assert.True(t, ok)
	assert.Equal(t, pw1, secret)

	err = provider.storeSecret(ctx, vault, ref, pw2)
	assert.NoError(t, err)

	item, err = provider.getItem(ctx, vault)
	assert.NoError(t, err)
	secret, ok = item.value(ref)
	assert.True(t, ok)
	assert.Equal(t, pw2, secret)
}
