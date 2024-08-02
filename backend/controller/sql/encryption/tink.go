package encryption

import (
	"bytes"
	"encoding/base64"
	"errors"
	"os"

	"github.com/google/tink/go/aead"
	"github.com/google/tink/go/insecurecleartextkeyset"
	"github.com/google/tink/go/keyset"
)

func CreateTinkEncryptionManager() (TinkEncryptionManager, error) {
	key := os.Getenv("FTL_TINK_ENCRYPTION_KEY")
	if key == "" {
		return TinkEncryptionManager{}, errors.New("FTL_TINK_ENCRYPTION_KEY was not set")
	}

	// Parse the keyset.
	parsedHandle, err := insecurecleartextkeyset.Read(
		keyset.NewBinaryReader(bytes.NewBuffer([]byte(key))))
	if err != nil {
		return TinkEncryptionManager{}, err
	}
	return TinkEncryptionManager{key: parsedHandle}, nil
}

type TinkEncryptionManager struct {
	key *keyset.Handle
}

func (t TinkEncryptionManager) Encrypt(plain string) (string, error) {

	// Get the primitive.
	primitive, err := aead.New(t.key)
	if err != nil {
		return "", err
	}

	ciphertext, err := primitive.Encrypt([]byte(plain), []byte{})
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}
func (t TinkEncryptionManager) Decrypt(ciphertext string) (string, error) {

	primitive, err := aead.New(t.key)
	decodeString, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	decrypted, err := primitive.Decrypt(decodeString, []byte{})
	if err != nil {
		return "", err
	}
	return string(decrypted), err
}
