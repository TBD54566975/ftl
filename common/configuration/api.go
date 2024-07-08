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
	"time"

	"github.com/alecthomas/types/optional"
	"github.com/puzpuzpuz/xsync/v3"
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
}

// Provider is a generic interface for storing and retrieving configuration and secrets.
type Provider[R Role] interface {
	Role() R
	Key() string

	// Store a configuration value and return its key.
	Store(ctx context.Context, ref Ref, value []byte) (*url.URL, error)
	// Delete a configuration value.
	Delete(ctx context.Context, ref Ref) error
}

// SyncableProvider is an interface for providers that can load values on-demand.
// This is recommended if the provider allows inexpensive loading of values.
type OnDemandProvider[R Role] interface {
	Provider[R]

	Load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error)
}

// SyncableProvider is an interface for providers that support syncing values.
// This is recommended if the provider allows batch access, or is expensive to load.
type SyncableProvider[R Role] interface {
	Provider[R]

	SyncInterval() time.Duration
	Sync(ctx context.Context, values *xsync.MapOf[Ref, SyncedValue]) error
}

type VersionToken any

type SyncedValue struct {
	Value []byte

	// VersionToken is a way of storing a version provided by the source of truth (eg: lastModified)
	// it is nil when:
	// - the owner of the cache is not using version tokens
	// - the cache is updated after writing
	VersionToken optional.Option[VersionToken]
}
