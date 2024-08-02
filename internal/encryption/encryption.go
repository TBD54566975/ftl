package encryption

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/tink-crypto/tink-go/v2/insecurecleartextkeyset"
	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/tink-crypto/tink-go/v2/streamingaead"
	"github.com/tink-crypto/tink-go/v2/tink"
	"io"
)

type Encryptable interface {
	EncryptJSON(input any) (json.RawMessage, error)
	DecryptJSON(input json.RawMessage, output any) error
}

func NewForKey(key string) (Encryptable, error) {
	if len(key) == 0 {
		return NoOpEncrypter{}, nil
	}

	keySetHandle, err := insecurecleartextkeyset.Read(
		keyset.NewJSONReader(bytes.NewBufferString(key)))
	if err != nil {
		return nil, fmt.Errorf("failed to read keyset: %w", err)
	}

	return NewEncrypter(*keySetHandle), nil
}

type NoOpEncrypter struct {
}

func (n NoOpEncrypter) EncryptJSON(input any) (json.RawMessage, error) {
	msg, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	return msg, nil
}

func (n NoOpEncrypter) DecryptJSON(input json.RawMessage, output any) error {
	err := json.Unmarshal(input, output)
	if err != nil {
		return fmt.Errorf("failed to unmarshal input: %w", err)
	}

	return nil
}

// NewEncrypter encrypts fields using AES256_GCM_HKDF_1MB
func NewEncrypter(keySet keyset.Handle) (Encryptable, error) {
	primitive, err := streamingaead.New(&keySet)
	if err != nil {
		return nil, fmt.Errorf("failed to create primitive during encryption: %w", err)
	}

	return Encrypter{keySetHandle: keySet, primitive: primitive}, nil
}

type Encrypter struct {
	keySetHandle keyset.Handle
	primitive    tink.StreamingAEAD
}

type EncryptedPayload struct {
	Encrypted []byte
}

func (e Encrypter) EncryptJSON(input any) (json.RawMessage, error) {
	// Retrieve the StreamingAEAD primitive we want to use from the keyset handle.

	msg, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	encryptedBuffer := &bytes.Buffer{}
	msgBuffer := bytes.NewBuffer(msg)
	writer, err := primitive.NewEncryptingWriter(encryptedBuffer, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create encrypting writer: %w", err)
	}

	if _, err := io.Copy(writer, msgBuffer); err != nil {
		return nil, fmt.Errorf("failed to copy encrypted data: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close encrypted writer: %w", err)
	}

	return json.Marshal(EncryptedPayload{Encrypted: encryptedBuffer.Bytes()})
}

func (e Encrypter) DecryptJSON(input json.RawMessage, output any) error {

}
