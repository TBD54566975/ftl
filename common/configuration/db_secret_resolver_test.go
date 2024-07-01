package configuration

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/TBD54566975/ftl/common/configuration/sql"
	"github.com/alecthomas/assert/v2"
	. "github.com/alecthomas/types/optional"
)

type mockDBSecretResolverDAL struct {
	entries []sql.ModuleSecret
}

func (d *mockDBSecretResolverDAL) findEntry(module Option[string], name string) (Option[sql.ModuleSecret], int) {
	for i := range d.entries {
		if d.entries[i].Module.Default("") == module.Default("") && d.entries[i].Name == name {
			return Some(d.entries[i]), i
		}
	}
	return None[sql.ModuleSecret](), -1
}

func (d *mockDBSecretResolverDAL) GetModuleSecretURL(ctx context.Context, module Option[string], name string) (string, error) {
	entry, _ := d.findEntry(module, name)
	if e, ok := entry.Get(); ok {
		return e.Url, nil
	}
	return "", fmt.Errorf("secret not found")
}

func (d *mockDBSecretResolverDAL) ListModuleSecrets(ctx context.Context) ([]sql.ModuleSecret, error) {
	return d.entries, nil
}

func (d *mockDBSecretResolverDAL) SetModuleSecretURL(ctx context.Context, module Option[string], name string, url string) error {
	d.UnsetModuleSecret(ctx, module, name)
	d.entries = append(d.entries, sql.ModuleSecret{Module: module, Name: name, Url: url})
	return nil
}

func (d *mockDBSecretResolverDAL) UnsetModuleSecret(ctx context.Context, module Option[string], name string) error {
	entry, i := d.findEntry(module, name)
	if _, ok := entry.Get(); ok {
		d.entries = append(d.entries[:i], d.entries[i+1:]...)
	}
	return nil
}

func TestDBSecretResolverList(t *testing.T) {
	ctx := context.Background()
	resolver := NewDBSecretResolver(&mockDBSecretResolverDAL{})

	rone := Ref{Module: Some("foo"), Name: "one"}
	resolver.Set(ctx, rone, &url.URL{Scheme: "asm", Host: rone.String()})

	rtwo := Ref{Module: Some("foo"), Name: "two"}
	resolver.Set(ctx, rtwo, &url.URL{Scheme: "asm", Host: rtwo.String()})

	entries, err := resolver.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(entries), 2)

	resolver.Unset(ctx, rone)

	entries, err = resolver.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(entries), 1)

	url, err := resolver.Get(ctx, rtwo)
	assert.NoError(t, err)
	assert.Equal(t, url.String(), "asm://foo.two")
}
