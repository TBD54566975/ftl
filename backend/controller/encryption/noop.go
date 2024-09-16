package encryption

import (
	"github.com/TBD54566975/ftl/backend/controller/encryption/api"
)

// NoOpEncryptor does not encrypt and just passes the input as is.
type NoOpEncryptor struct{}

func NewNoOpEncryptor() NoOpEncryptor {
	return NoOpEncryptor{}
}

var _ api.DataEncryptor = NoOpEncryptor{}

func (n NoOpEncryptor) Encrypt(cleartext []byte, dest api.Encrypted) error {
	dest.Set(cleartext)
	return nil
}

func (n NoOpEncryptor) Decrypt(encrypted api.Encrypted) ([]byte, error) {
	return encrypted.Bytes(), nil
}
