package routers_test

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/internal/configuration"
	"github.com/block/ftl/internal/configuration/manager"
	"github.com/block/ftl/internal/configuration/providers"
	"github.com/block/ftl/internal/configuration/routers"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/projectconfig"
)

func TestSet(t *testing.T) {
	defaultPath, ok := projectconfig.DefaultConfigPath().Get()
	assert.True(t, ok)
	origConfigBytes, err := os.ReadFile(defaultPath)
	assert.NoError(t, err)

	config := filepath.Join(t.TempDir(), "ftl-project.toml")
	existing, err := os.ReadFile("testdata/ftl-project.toml")
	assert.NoError(t, err)
	err = os.WriteFile(config, existing, 0600)
	assert.NoError(t, err)

	setAndAssert(t, "echo", config)
	setAndAssert(t, "echooo", config)

	// Restore the original config file.
	err = os.WriteFile(defaultPath, origConfigBytes, 0600)
	assert.NoError(t, err)
}

func TestGetGlobal(t *testing.T) {
	config := filepath.Join(t.TempDir(), "ftl-project.toml")
	existing, err := os.ReadFile("testdata/ftl-project.toml")
	assert.NoError(t, err)
	err = os.WriteFile(config, existing, 0600)
	assert.NoError(t, err)

	t.Run("ExistingModule", func(t *testing.T) {
		setAndAssert(t, "echo", config)
	})
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	cf, err := manager.New(ctx, routers.ProjectConfig[configuration.Configuration]{Config: config}, providers.Inline[configuration.Configuration]{})
	assert.NoError(t, err)

	var got *url.URL
	want := URL("inline://qwertyqwerty")
	err = cf.Set(ctx, configuration.Ref{Module: optional.None[string](), Name: "default"}, want)
	assert.NoError(t, err)
	err = cf.Get(ctx, configuration.Ref{Module: optional.Some[string]("somemodule"), Name: "default"}, &got)
	assert.NoError(t, err)

	// Get should return `want` even though it was only set globally
	assert.Equal(t, want, got)
}

func setAndAssert(t *testing.T, module string, config string) {
	t.Helper()

	ctx := log.ContextWithNewDefaultLogger(context.Background())

	cf, err := manager.New(ctx, routers.ProjectConfig[configuration.Configuration]{Config: config}, providers.Inline[configuration.Configuration]{})
	assert.NoError(t, err)

	var got *url.URL
	want := URL("inline://asdfasdf")
	err = cf.Set(ctx, configuration.Ref{Module: optional.Some[string](module), Name: "default"}, want)
	assert.NoError(t, err)
	err = cf.Get(ctx, configuration.Ref{Module: optional.Some[string](module), Name: "default"}, &got)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func URL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
