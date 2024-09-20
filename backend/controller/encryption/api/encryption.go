package api

import (
	"context"
)

// Encrypted is an interface for values that contain encrypted data.
type Encrypted interface {
	SubKey() string
	Bytes() []byte
	Set(data []byte)
}

type KeyStoreProvider interface {
	// EnsureKey asks a provider to check for an encrypted key.
	// If not available, call the generateKey function to create a new key.
	// The provider should handle transactions around checking and setting the key, to prevent race conditions.
	EnsureKey(ctx context.Context, generateKey func() ([]byte, error)) ([]byte, error)
}

type DataEncryptor interface {
	Encrypt(cleartext []byte, dest Encrypted) error
	Decrypt(encrypted Encrypted) ([]byte, error)
	// EncryptIdentityKey(keyPair internalidentity.KeyPair, dest api.Encrypted) error
}
