package routerstest

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/internal/configuration/dal"
)

type MockDBSecretResolverDAL struct {
	lock    sync.Mutex
	entries []dal.ModuleSecret
}

func (d *MockDBSecretResolverDAL) FindEntry(module optional.Option[string], name string) (optional.Option[dal.ModuleSecret], int) {
	for i := range d.entries {
		if d.entries[i].Module.Default("") == module.Default("") && d.entries[i].Name == name {
			return optional.Some(d.entries[i]), i
		}
	}
	return optional.None[dal.ModuleSecret](), -1
}

func (d *MockDBSecretResolverDAL) GetModuleSecretURL(ctx context.Context, module optional.Option[string], name string) (string, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	entry, _ := d.FindEntry(module, name)
	if e, ok := entry.Get(); ok {
		return e.Url, nil
	}
	return "", fmt.Errorf("secret not found")
}

func (d *MockDBSecretResolverDAL) ListModuleSecrets(ctx context.Context) ([]dal.ModuleSecret, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	return slices.Clone(d.entries), nil
}

func (d *MockDBSecretResolverDAL) SetModuleSecretURL(ctx context.Context, module optional.Option[string], name string, url string) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.remove(module, name)
	d.entries = append(d.entries, dal.ModuleSecret{Module: module, Name: name, Url: url})
	return nil
}

func (d *MockDBSecretResolverDAL) UnsetModuleSecret(ctx context.Context, module optional.Option[string], name string) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.remove(module, name)
	return nil
}

func (d *MockDBSecretResolverDAL) remove(module optional.Option[string], name string) {
	entry, i := d.FindEntry(module, name)
	if _, ok := entry.Get(); ok {
		d.entries = append(d.entries[:i], d.entries[i+1:]...)
	}
}
