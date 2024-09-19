package api

import "context"

type KeyPair interface {
	Signer() (IdentitySigner, error)
	Verifier() (IdentityVerifier, error)
}

type IdentityVerifier interface {
	Verify(signedData SignedData) error
}

type IdentitySigner interface {
	Sign(data []byte) (*SignedData, error)
}

type SignedData struct {
	Data      []byte
	Signature []byte
}

type IdentityKeyStoreProvider interface {
	// EnsureKey asks a provider to check for an identity key.
	// If not available, call the generateKey function to create a new key.
	// The provider should handle transactions around checking and setting the key, to prevent race conditions.
	EnsureKey(ctx context.Context, generateKey func() ([]byte, error)) ([]byte, error)
}
