// Package configuration is the FTL configuration and secret management API.
//
// The full design is documented [here].
//
// A [Manager] is the high-level interface to storing, listing, and retrieving
// secrets and configuration. A [Resolver] is the next layer, mapping
// names to a storage location key such as environment variables, keychain, etc.
// The [Provider] is the final layer, responsible for actually storing and
// retrieving values in concrete storage.
//
// A constructed [Manager] and its providers are parametric on either secrets or
// configuration and thus cannot be used interchangeably.
//
// [here]: https://hackmd.io/@ftl/S1e6YVEuq6
package configuration

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/alecthomas/types/optional"
)

// ErrNotFound is returned when a configuration entry is not found or cannot be resolved.
var ErrNotFound = errors.New("not found")

// Entry in the configuration store.
type Entry struct {
	Ref
	Accessor *url.URL
}

// A Ref is a reference to a configuration value.
type Ref struct {
	Module optional.Option[string]
	Name   string
}

// NewRef creates a new Ref.
//
// If [module] is empty, the Ref is considered to be a global configuration value.
func NewRef(module, name string) Ref {
	return Ref{Module: optional.Zero(module), Name: name}
}

func ParseRef(s string) (Ref, error) {
	ref := Ref{}
	err := ref.UnmarshalText([]byte(s))
	return ref, err
}

func (k Ref) String() string {
	if m, ok := k.Module.Get(); ok {
		return m + "." + k.Name
	}
	return k.Name
}

func (k *Ref) UnmarshalText(text []byte) error {
	s := string(text)
	if i := strings.Index(s, "."); i != -1 {
		k.Module = optional.Some(s[:i])
		k.Name = s[i+1:]
	} else {
		k.Name = s
	}
	return nil
}

// A Router resolves configuration names to keys that are then used to load
// values from a Provider.
//
// This indirection allows for the storage of configuration values to be
// abstracted from the configuration itself. For example, the ftl-project.toml
// file contains per-module and global configuration maps, but the secrets
// themselves may be stored in a separate secret store such as a system keychain.
type Router[R Role] interface {
	Role() R
	Get(ctx context.Context, ref Ref) (key *url.URL, err error)
	Set(ctx context.Context, ref Ref, key *url.URL) error
	Unset(ctx context.Context, ref Ref) error
	List(ctx context.Context) ([]Entry, error)

	// TODO: Routers and Providers have become conflated, and this method is a
	// hack to allow mutation of values to occur in a Provider only if the
	// Router is compatible with it.
	//
	// (i.e. all providers are compatible with the projectconfig_resolver, but
	// only the asm provider is compatible with the asm_resolver)
	//
	// This should be refactored to use a single entity, and optionally also
	// write to the projectconfig.
	UseWithProvider(ctx context.Context, pkey string) bool
}

// Provider is a generic interface for storing and retrieving configuration and secrets.
type Provider[R Role] interface {
	Role() R
	Key() string
	Load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error)
	// Store a configuration value and return its key.
	Store(ctx context.Context, ref Ref, value []byte) (*url.URL, error)
	// Delete a configuration value.
	Delete(ctx context.Context, ref Ref) error
}
