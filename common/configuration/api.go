// Package configuration is a generic configuration and secret management API.
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

// A Resolver resolves configuration names to keys that are then used to load
// values from a Provider.
//
// This indirection allows for the storage of configuration values to be
// abstracted from the configuration itself. For example, the ftl-project.toml
// file contains per-module and global configuration maps, but the secrets
// themselves may be stored in a separate secret store such as a system keychain.
type Resolver interface {
	Get(ctx context.Context, ref Ref) (key *url.URL, err error)
	Set(ctx context.Context, ref Ref, key *url.URL) error
	Unset(ctx context.Context, ref Ref) error
	List(ctx context.Context) ([]Entry, error)
}

// Provider is a generic interface for storing and retrieving configuration and secrets.
type Provider interface {
	Key() string
	Load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error)
}

// A MutableProvider is a Provider that can update configuration.
type MutableProvider interface {
	Provider
	// Writer returns true if this provider should be used to store configuration.
	//
	// Only one provider should return true.
	//
	// To be usable from the CLI, each provider must be a Kong-compatible struct
	// containing a flag that this method should return. For example:
	//
	// 	type InlineProvider struct {
	// 		Inline bool `help:"Write values inline." group:"Provider:" xor:"configwriter"`
	// 	}
	//
	//	func (i InlineProvider) Writer() bool { return i.Inline }
	//
	// The "xor" tag is used to ensure that only one writer is selected.
	Writer() bool
	// Store a configuration value and return its key.
	Store(ctx context.Context, ref Ref, value []byte) (*url.URL, error)
	// Delete a configuration value.
	Delete(ctx context.Context, ref Ref) error
}
