package configuration

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	keyring "github.com/zalando/go-keyring"
)

type KeychainProvider struct{}

var _ SynchronousProvider[Secrets] = KeychainProvider{}

func (KeychainProvider) Role() Secrets { return Secrets{} }
func (k KeychainProvider) Key() string { return "keychain" }

func (k KeychainProvider) Load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error) {
	value, err := keyring.Get(k.serviceName(ref), key.Host)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil, fmt.Errorf("no keychain entry for %q: %w", key.Host, ErrNotFound)
		}
		return nil, err
	}
	return []byte(value), nil
}

func (k KeychainProvider) Store(ctx context.Context, ref Ref, value []byte) (*url.URL, error) {
	err := keyring.Set(k.serviceName(ref), ref.Name, string(value))
	if err != nil {
		return nil, err
	}
	return &url.URL{Scheme: "keychain", Host: ref.Name}, nil
}

func (k KeychainProvider) Delete(ctx context.Context, ref Ref) error {
	err := keyring.Delete(k.serviceName(ref), ref.Name)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return fmt.Errorf("no keychain entry for %q: %w", ref, ErrNotFound)
		}
		return err
	}
	return nil
}

func (k KeychainProvider) serviceName(ref Ref) string {
	return "ftl-secret-" + strings.ReplaceAll(ref.String(), ".", "-")
}
