package manager

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
	"github.com/zalando/go-keyring"

	"github.com/block/ftl/internal/configuration"
	"github.com/block/ftl/internal/configuration/providers"
	"github.com/block/ftl/internal/configuration/routers"
	"github.com/block/ftl/internal/log"
)

func TestManager(t *testing.T) {
	t.Skip("This will be replaced soon")
	keyring.MockInit() // There's no way to undo this :\
	config := tempConfigPath(t, "testdata/ftl-project.toml", "manager")
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	t.Run("Secrets", func(t *testing.T) {
		cf, err := New(
			ctx,
			routers.ProjectConfig[configuration.Secrets]{Config: config},
			providers.NewInline[configuration.Secrets](),
		)
		assert.NoError(t, err)
		testManager(t, ctx, cf, []configuration.Entry{
			{Ref: configuration.Ref{Name: "baz"}, Accessor: URL("inline://MDcwZFFEVmoxVmRuQ1NoRGFyRjR4bENiZFpwMGdkTzFqd2M2aTRRPQ")},
			{Ref: configuration.Ref{Name: "foo"}, Accessor: URL("inline://ImJhciI")},
			{Ref: configuration.Ref{Name: "mutable"}, Accessor: URL("inline://eXJYMWNNVmtwVHhUa1FkcFJ2bTNLaXVESTJEZ1lBRT0")},
		})
	})
	t.Run("Configuration", func(t *testing.T) {
		cf, err := New(
			ctx,
			routers.ProjectConfig[configuration.Configuration]{Config: config},
			providers.Inline[configuration.Configuration]{},
		)
		assert.NoError(t, err)
		testManager(t, ctx, cf, []configuration.Entry{
			{Ref: configuration.Ref{Name: "baz"}, Accessor: URL("inline://eyJiYXoiOiJ3YXoifQ")},
			{Ref: configuration.Ref{Name: "foo"}, Accessor: URL("inline://ImJhciI")},
			{Ref: configuration.Ref{Name: "mutable"}, Accessor: URL("inline://ImhlbGxvIg")},
			{Ref: configuration.Ref{Module: optional.Some[string]("echo"), Name: "default"}, Accessor: URL("inline://ImFub255bW91cyI")},
		})
	})
}

// TestMapPriority checks that module specific configs beat global configs when flattening for a module
func TestMapPriority(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	config := tempConfigPath(t, "", "map")
	cm, err := New[configuration.Configuration](
		ctx,
		routers.ProjectConfig[configuration.Configuration]{Config: config},
		providers.Inline[configuration.Configuration]{},
	)
	assert.NoError(t, err)
	moduleName := "test"

	// Set 50 configs and 50 global configs
	// It's hard to tell if module config beats global configs because we are dealing with unordered maps, or because the logic is correct
	// Repeating it 50 times hopefully gives us a good chance of catching inconsistencies
	for i := range 50 {
		key := fmt.Sprintf("key%d", i)

		strValue := "HelloWorld"
		globalStrValue := "GlobalHelloWorld"
		if i%2 == 0 {
			// sometimes try setting the module config first
			assert.NoError(t, cm.Set(ctx, configuration.Ref{Module: optional.Some(moduleName), Name: key}, strValue))
			assert.NoError(t, cm.Set(ctx, configuration.Ref{Module: optional.None[string](), Name: key}, globalStrValue))
		} else {
			// other times try setting the global config first
			assert.NoError(t, cm.Set(ctx, configuration.Ref{Module: optional.None[string](), Name: key}, globalStrValue))
			assert.NoError(t, cm.Set(ctx, configuration.Ref{Module: optional.Some(moduleName), Name: key}, strValue))
		}
	}
	result, err := cm.MapForModule(ctx, moduleName)
	assert.NoError(t, err)

	for i := range 50 {
		key := fmt.Sprintf("key%d", i)
		assert.Equal(t, `"HelloWorld"`, string(result[key]), "module configs should beat global configs")
	}
}

func tempConfigPath(t *testing.T, existingPath string, prefix string) string {
	t.Helper()

	config := filepath.Join(t.TempDir(), fmt.Sprintf("%s-ftl-project.toml", prefix))
	var existing []byte
	var err error
	if existingPath == "" {
		existing = []byte(`
			name = "generated"
			secrets-provider = "inline"
			config-provider = "inline"
		`)
	} else {
		existing, err = os.ReadFile(existingPath)
		assert.NoError(t, err)
	}
	err = os.WriteFile(config, existing, 0600)
	assert.NoError(t, err)
	return config
}

// nolint
func testManager[R configuration.Role](
	t *testing.T,
	ctx context.Context,
	cf *Manager[R],
	expectedListing []configuration.Entry,
) {
	actualListing, err := cf.List(ctx)
	assert.NoError(t, err)

	assert.Equal(t, expectedListing, actualListing)
	// Try to get value from missing envar
	var bazValue map[string]string

	err = cf.Get(ctx, configuration.Ref{Name: "baz"}, &bazValue)
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"baz": "waz"}, bazValue)

	var fooValue string
	err = cf.Get(ctx, configuration.Ref{Name: "foo"}, &fooValue)
	assert.NoError(t, err)
	assert.Equal(t, "bar", fooValue)

	err = cf.Get(ctx, configuration.Ref{Name: "nonexistent"}, &fooValue)
	assert.IsError(t, err, configuration.ErrNotFound)

	// Change value.
	err = cf.Set(ctx, configuration.Ref{Name: "mutable"}, "hello")
	assert.NoError(t, err)

	err = cf.Get(ctx, configuration.Ref{Name: "mutable"}, &fooValue)
	assert.NoError(t, err)
	assert.Equal(t, "hello", fooValue)

	actualListing, err = cf.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedListing, actualListing)

	// Delete value
	err = cf.Unset(ctx, configuration.Ref{Name: "foo"})
	assert.NoError(t, err)
	err = cf.Get(ctx, configuration.Ref{Name: "foo"}, &fooValue)
	assert.IsError(t, err, configuration.ErrNotFound)
}

func URL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
