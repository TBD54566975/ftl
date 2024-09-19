package identity

import "context"

type KeyPair interface {
	Signer() (Signer, error)
	Verifier() (Verifier, error)
}

type Verifier interface {
	Verify(signedData SignedData) error
}

type Signer interface {
	Sign(data []byte) (*SignedData, error)
}

type SignedData struct {
	Data      []byte
	Signature []byte
}

type KeyStoreProvider interface {
	// EnsureKey asks a provider to check for an identity key.
	// If not available, call the generateKey function to create a new key.
	// The provider should handle transactions around checking and setting the key, to prevent race conditions.
	EnsureKey(ctx context.Context, generateKey func() ([]byte, error)) ([]byte, error)
}
