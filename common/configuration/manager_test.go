package configuration

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

	"github.com/TBD54566975/ftl/internal/log"
)

func TestManager(t *testing.T) {
	keyring.MockInit() // There's no way to undo this :\
	config := tempConfigPath(t, "testdata/ftl-project.toml", "manager")
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	t.Run("Secrets", func(t *testing.T) {
		kcp := KeychainProvider{Keychain: true}
		_, err := kcp.Store(ctx, Ref{Name: "mutable"}, []byte("hello"))
		assert.NoError(t, err)
		cf, err := New(ctx,
			ProjectConfigResolver[Secrets]{Config: []string{config}},
			[]Provider[Secrets]{
				EnvarProvider[Secrets]{},
				InlineProvider[Secrets]{},
				kcp,
			})
		assert.NoError(t, err)
		testManager(t, ctx, cf, "FTL_SECRET_YmF6", []Entry{
			{Ref: Ref{Name: "baz"}, Accessor: URL("envar://baz")},
			{Ref: Ref{Name: "foo"}, Accessor: URL("inline://ImJhciI")},
			{Ref: Ref{Name: "mutable"}, Accessor: URL("keychain://mutable")},
		})
	})
	t.Run("Configuration", func(t *testing.T) {
		cf, err := New(ctx,
			ProjectConfigResolver[Configuration]{Config: []string{config}},
			[]Provider[Configuration]{
				EnvarProvider[Configuration]{},
				InlineProvider[Configuration]{Inline: true}, // Writer
			})
		assert.NoError(t, err)
		testManager(t, ctx, cf, "FTL_CONFIG_YmF6", []Entry{
			{Ref: Ref{Name: "baz"}, Accessor: URL("envar://baz")},
			{Ref: Ref{Name: "foo"}, Accessor: URL("inline://ImJhciI")},
			{Ref: Ref{Name: "mutable"}, Accessor: URL("inline://ImhlbGxvIg")},
			{Ref: Ref{Module: optional.Some[string]("echo"), Name: "default"}, Accessor: URL("inline://ImFub255bW91cyI")},
		})
	})
}

// TestMapPriority checks that module specific configs beat global configs when flattening for a module
func TestMapPriority(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	config := tempConfigPath(t, "", "map")
	cm, err := New(ctx,
		ProjectConfigResolver[Configuration]{Config: []string{config}},
		[]Provider[Configuration]{
			InlineProvider[Configuration]{
				Inline: true,
			},
		})
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
			assert.NoError(t, cm.Set(ctx, Ref{Module: optional.Some(moduleName), Name: key}, strValue))
			assert.NoError(t, cm.Set(ctx, Ref{Module: optional.None[string](), Name: key}, globalStrValue))
		} else {
			// other times try setting the global config first
			assert.NoError(t, cm.Set(ctx, Ref{Module: optional.None[string](), Name: key}, globalStrValue))
			assert.NoError(t, cm.Set(ctx, Ref{Module: optional.Some(moduleName), Name: key}, strValue))
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
		existing = []byte{}
	} else {
		existing, err = os.ReadFile(existingPath)
		assert.NoError(t, err)
	}
	err = os.WriteFile(config, existing, 0600)
	assert.NoError(t, err)
	return config
}

// nolint
func testManager[R Role](
	t *testing.T,
	ctx context.Context,
	cf *Manager[R],
	envarName string,
	expectedListing []Entry,
) {
	actualListing, err := cf.List(ctx)
	assert.NoError(t, err)

	assert.Equal(t, expectedListing, actualListing)
	// Try to get value from missing envar
	var bazValue map[string]string

	err = cf.Get(ctx, Ref{Name: "baz"}, &bazValue)
	assert.IsError(t, err, ErrNotFound)

	// Set the envar and try again.
	t.Setenv(envarName, "eyJiYXoiOiJ3YXoifQ") // baz={"baz": "waz"}

	err = cf.Get(ctx, Ref{Name: "baz"}, &bazValue)
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"baz": "waz"}, bazValue)

	var fooValue string
	err = cf.Get(ctx, Ref{Name: "foo"}, &fooValue)
	assert.NoError(t, err)
	assert.Equal(t, "bar", fooValue)

	err = cf.Get(ctx, Ref{Name: "nonexistent"}, &fooValue)
	assert.IsError(t, err, ErrNotFound)

	// Change value.
	err = cf.Set(ctx, Ref{Name: "mutable"}, "hello")
	assert.NoError(t, err)

	err = cf.Get(ctx, Ref{Name: "mutable"}, &fooValue)
	assert.NoError(t, err)
	assert.Equal(t, "hello", fooValue)

	actualListing, err = cf.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedListing, actualListing)

	// Delete value
	err = cf.Unset(ctx, Ref{Name: "foo"})
	assert.NoError(t, err)
	err = cf.Get(ctx, Ref{Name: "foo"}, &fooValue)
	assert.IsError(t, err, ErrNotFound)
}

func URL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
