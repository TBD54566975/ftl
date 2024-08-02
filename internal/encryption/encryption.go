package encryption

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/tink-crypto/tink-go/v2/insecurecleartextkeyset"
	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/tink-crypto/tink-go/v2/streamingaead"
	"github.com/tink-crypto/tink-go/v2/tink"
)

type Encryptable interface {
	EncryptJSON(input any) (json.RawMessage, error)
	DecryptJSON(input json.RawMessage, output any) error
}

func NewForKeyOrUri(keyOrUri string) (Encryptable, error) {
	if len(keyOrUri) == 0 {
		return NoOpEncryptor{}, nil
	}

	// If keyOrUri is a JSON string, it is a clear text key set.
	if strings.TrimSpace(keyOrUri)[0] == '{' {
		return NewClearTextEncryptor(keyOrUri)
		// Otherwise should be a URI for KMS.
		// aws-kms://arn:aws:kms:[region]:[account-id]:key/[key-id]
	} else if strings.HasPrefix(keyOrUri, "aws-kms://") {
		return nil, fmt.Errorf("AWS KMS is not supported yet")
	} else {
		return nil, fmt.Errorf("unsupported key or uri: %s", keyOrUri)
	}
}

// NoOpEncryptor does not encrypt and just passes the input as is.
type NoOpEncryptor struct {
}

func (n NoOpEncryptor) EncryptJSON(input any) (json.RawMessage, error) {
	msg, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	return msg, nil
}

func (n NoOpEncryptor) DecryptJSON(input json.RawMessage, output any) error {
	err := json.Unmarshal(input, output)
	if err != nil {
		return fmt.Errorf("failed to unmarshal input: %w", err)
	}

	return nil
}

func NewClearTextEncryptor(key string) (Encryptable, error) {
	keySetHandle, err := insecurecleartextkeyset.Read(
		keyset.NewJSONReader(bytes.NewBufferString(key)))
	if err != nil {
		return nil, fmt.Errorf("failed to read clear text keyset: %w", err)
	}

	encryptor, err := NewEncryptor(*keySetHandle)
	if err != nil {
		return nil, fmt.Errorf("failed to create clear text encryptor: %w", err)
	}

	return encryptor, nil
}

// NewEncryptor encrypts and decrypts JSON payloads using the provided key set.
// The key set must contain a primary key that is a streaming AEAD primitive.
func NewEncryptor(keySet keyset.Handle) (Encryptable, error) {
	primitive, err := streamingaead.New(&keySet)
	if err != nil {
		return nil, fmt.Errorf("failed to create primitive during encryption: %w", err)
	}

	return Encryptor{keySetHandle: keySet, primitive: primitive}, nil
}

type Encryptor struct {
	keySetHandle keyset.Handle
	primitive    tink.StreamingAEAD
}

type EncryptedPayload struct {
	Encrypted []byte `json:"encrypted"`
}

func (e Encryptor) EncryptJSON(input any) (json.RawMessage, error) {
	msg, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	encryptedBuffer := &bytes.Buffer{}
	msgBuffer := bytes.NewBuffer(msg)
	writer, err := e.primitive.NewEncryptingWriter(encryptedBuffer, nil)
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

func (e Encryptor) DecryptJSON(input json.RawMessage, output any) error {
	var payload EncryptedPayload
	if err := json.Unmarshal(input, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal encrypted payload: %w", err)
	}

	inputBytesReader := bytes.NewReader(payload.Encrypted)
	reader, err := e.primitive.NewDecryptingReader(inputBytesReader, nil)
	if err != nil {
		return fmt.Errorf("failed to create decrypting reader: %w", err)
	}

	decryptedBuffer := &bytes.Buffer{}
	if _, err := io.Copy(decryptedBuffer, reader); err != nil {
		return fmt.Errorf("failed to copy decrypted data: %w", err)
	}

	if err := json.Unmarshal(decryptedBuffer.Bytes(), output); err != nil {
		return fmt.Errorf("failed to unmarshal decrypted data: %w", err)
	}

	return nil
}
