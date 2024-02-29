package configuration

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/zalando/go-keyring"

	"github.com/TBD54566975/ftl/internal/log"
)

func TestManager(t *testing.T) {
	keyring.MockInit() // There's no way to undo this :\

	config := filepath.Join(t.TempDir(), "ftl-project.toml")
	existing, err := os.ReadFile("testdata/ftl-project.toml")
	assert.NoError(t, err)
	err = os.WriteFile(config, existing, 0600)
	assert.NoError(t, err)

	ctx := log.ContextWithNewDefaultLogger(context.Background())
	cf, err := New(ctx,
		ProjectConfigResolver[FromConfig]{Config: config},
		[]Provider{
			EnvarProvider[EnvarTypeConfig]{},
			InlineProvider{Inline: true}, // Writer
			KeychainProvider{},
		})
	assert.NoError(t, err)

	actual, err := cf.List(ctx)
	assert.NoError(t, err)

	expected := []Entry{
		{Ref: Ref{Name: "baz"}, Accessor: URL("envar://baz")},
		{Ref: Ref{Name: "foo"}, Accessor: URL("inline://ImJhciI")},
		{Ref: Ref{Name: "keychain"}, Accessor: URL("keychain://keychain")},
	}

	assert.Equal(t, expected, actual)

	// Try to get value from missing envar
	var bazValue map[string]string

	err = cf.Get(ctx, Ref{Name: "baz"}, &bazValue)
	assert.IsError(t, err, ErrNotFound)

	// Set the envar and try again.
	t.Setenv("FTL_CONFIG_YmF6", "eyJiYXoiOiJ3YXoifQ") // baz={"baz": "waz"}

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
	err = cf.Set(ctx, Ref{Name: "foo"}, "hello")
	assert.NoError(t, err)

	err = cf.Get(ctx, Ref{Name: "foo"}, &fooValue)
	assert.NoError(t, err)
	assert.Equal(t, "hello", fooValue)

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
