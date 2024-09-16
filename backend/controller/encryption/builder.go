package encryption

import (
	"context"
	"fmt"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/encryption/api"
)

// Builder constructs a DataEncryptor when used with a provider.
// Use a chain of With* methods to configure the builder.
type Builder struct {
	kmsURI optional.Option[string]
}

func NewBuilder() Builder {
	return Builder{
		kmsURI: optional.None[string](),
	}
}

// WithKMSURI sets the URI for the KMS key to use. Omitting this call or using None will create a NoOpEncryptor.
func (b Builder) WithKMSURI(kmsURI optional.Option[string]) Builder {
	b.kmsURI = kmsURI
	return b
}

func (b Builder) Build(ctx context.Context, provider api.KeyStoreProvider) (api.DataEncryptor, error) {
	kmsURI, ok := b.kmsURI.Get()
	if !ok {
		return NewNoOpEncryptor(), nil
	}

	key, err := provider.EnsureKey(ctx, func() ([]byte, error) {
		return newKey(kmsURI, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to ensure key from provider: %w", err)
	}

	encryptor, err := NewKMSEncryptorWithKMS(kmsURI, nil, key)
	if err != nil {
		return nil, fmt.Errorf("failed to create KMS encryptor: %w", err)
	}

	return encryptor, nil
}
