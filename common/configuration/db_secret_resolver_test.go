package configuration

import (
	"context"
	"fmt"
	"net/url"
	"slices"
	"sync"
	"testing"

	"github.com/TBD54566975/ftl/common/configuration/dal"
	"github.com/alecthomas/assert/v2"
	. "github.com/alecthomas/types/optional"
)

type mockDBSecretResolverDAL struct {
	lock    sync.Mutex
	entries []dal.ModuleSecret
}

func (d *mockDBSecretResolverDAL) findEntry(module Option[string], name string) (Option[dal.ModuleSecret], int) {
	for i := range d.entries {
		if d.entries[i].Module.Default("") == module.Default("") && d.entries[i].Name == name {
			return Some(d.entries[i]), i
		}
	}
	return None[dal.ModuleSecret](), -1
}

func (d *mockDBSecretResolverDAL) GetModuleSecretURL(ctx context.Context, module Option[string], name string) (string, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	entry, _ := d.findEntry(module, name)
	if e, ok := entry.Get(); ok {
		return e.Url, nil
	}
	return "", fmt.Errorf("secret not found")
}

func (d *mockDBSecretResolverDAL) ListModuleSecrets(ctx context.Context) ([]dal.ModuleSecret, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	return slices.Clone(d.entries), nil
}

func (d *mockDBSecretResolverDAL) SetModuleSecretURL(ctx context.Context, module Option[string], name string, url string) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.remove(module, name)
	d.entries = append(d.entries, dal.ModuleSecret{Module: module, Name: name, Url: url})
	return nil
}

func (d *mockDBSecretResolverDAL) UnsetModuleSecret(ctx context.Context, module Option[string], name string) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.remove(module, name)
	return nil
}

func (d *mockDBSecretResolverDAL) remove(module Option[string], name string) {
	entry, i := d.findEntry(module, name)
	if _, ok := entry.Get(); ok {
		d.entries = append(d.entries[:i], d.entries[i+1:]...)
	}
}

func TestDBSecretResolverList(t *testing.T) {
	ctx := context.Background()
	resolver := NewDBSecretResolver(&mockDBSecretResolverDAL{})

	rone := Ref{Module: Some("foo"), Name: "one"}
	err := resolver.Set(ctx, rone, &url.URL{Scheme: "asm", Host: rone.String()})
	assert.NoError(t, err)

	rtwo := Ref{Module: Some("foo"), Name: "two"}
	err = resolver.Set(ctx, rtwo, &url.URL{Scheme: "asm", Host: rtwo.String()})
	assert.NoError(t, err)

	entries, err := resolver.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(entries), 2)

	err = resolver.Unset(ctx, rone)
	assert.NoError(t, err)

	entries, err = resolver.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(entries), 1)

	url, err := resolver.Get(ctx, rtwo)
	assert.NoError(t, err)
	assert.Equal(t, url.String(), "asm://foo.two")
}
