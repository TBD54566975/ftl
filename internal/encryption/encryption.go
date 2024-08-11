package encryption

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	awsv1kms "github.com/aws/aws-sdk-go/service/kms"
	"github.com/tink-crypto/tink-go-awskms/integration/awskms"
	"github.com/tink-crypto/tink-go/v2/aead"
	"github.com/tink-crypto/tink-go/v2/core/registry"
	"github.com/tink-crypto/tink-go/v2/insecurecleartextkeyset"
	"github.com/tink-crypto/tink-go/v2/keyderivation"
	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/tink-crypto/tink-go/v2/prf"
	"github.com/tink-crypto/tink-go/v2/streamingaead"
	"github.com/tink-crypto/tink-go/v2/testing/fakekms"
	"github.com/tink-crypto/tink-go/v2/tink"
)

// Encryptable is an interface for encrypting and decrypting JSON payloads.
// Deprecated: This is will be changed or removed very soon.
type Encryptable interface {
	EncryptJSON(input any) (json.RawMessage, error)
	DecryptJSON(input json.RawMessage, output any) error
}

// NewForKeyOrURI creates a new encryptor using the provided key or URI.
// Deprecated: This is will be changed or removed very soon.
func NewForKeyOrURI(keyOrURI string) (Encryptable, error) {
	if len(keyOrURI) == 0 {
		return NoOpEncryptor{}, nil
	}

	// If keyOrUri is a JSON string, it is a clear text key set.
	if strings.TrimSpace(keyOrURI)[0] == '{' {
		return NewClearTextEncryptor(keyOrURI)
		// Otherwise should be a URI for KMS.
		// aws-kms://arn:aws:kms:[region]:[account-id]:key/[key-id]
	} else if strings.HasPrefix(keyOrURI, "aws-kms://") {
		panic("not implemented")
	}
	return nil, fmt.Errorf("unsupported key or uri: %s", keyOrURI)
}

// NoOpEncryptor does not encrypt or decrypt and just passes the input as is.
// Deprecated: This is will be changed or removed very soon.
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

	encryptor, err := NewDeprecatedEncryptor(*keySetHandle)
	if err != nil {
		return nil, fmt.Errorf("failed to create clear text encryptor: %w", err)
	}

	return encryptor, nil
}

// NewDeprecatedEncryptor encrypts and decrypts JSON payloads using the provided key set.
// The key set must contain a primary key that is a streaming AEAD primitive.
func NewDeprecatedEncryptor(keySet keyset.Handle) (Encryptable, error) {
	primitive, err := streamingaead.New(&keySet)
	if err != nil {
		return nil, fmt.Errorf("failed to create primitive during encryption: %w", err)
	}

	return Encryptor{keySetHandle: keySet, primitive: primitive}, nil
}

// Encryptor uses streaming with JSON payloads.
// Deprecated: This is will be changed or removed very soon.
type Encryptor struct {
	keySetHandle keyset.Handle
	primitive    tink.StreamingAEAD
}

// EncryptedPayload is a JSON payload that contains the encrypted data to put into the database.
// Deprecated: This is will be changed or removed very soon.
type EncryptedPayload struct {
	Encrypted []byte `json:"encrypted"`
}

func (e Encryptor) EncryptJSON(input any) (json.RawMessage, error) {
	msg, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	encrypted, err := encryptBytesForStreaming(e.primitive, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	out, err := json.Marshal(EncryptedPayload{Encrypted: encrypted})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal encrypted data: %w", err)
	}
	return out, nil
}

func (e Encryptor) DecryptJSON(input json.RawMessage, output any) error {
	var payload EncryptedPayload
	if err := json.Unmarshal(input, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal encrypted payload: %w", err)
	}

	decryptedBuffer, err := decryptBytesForStreaming(e.primitive, payload.Encrypted)
	if err != nil {
		return fmt.Errorf("failed to decrypt data: %w", err)
	}

	if err := json.Unmarshal(decryptedBuffer, output); err != nil {
		return fmt.Errorf("failed to unmarshal decrypted data: %w", err)
	}

	return nil
}

func encryptBytesForStreaming(streamingPrimitive tink.StreamingAEAD, clearText []byte) ([]byte, error) {
	encryptedBuffer := &bytes.Buffer{}
	msgBuffer := bytes.NewBuffer(clearText)
	writer, err := streamingPrimitive.NewEncryptingWriter(encryptedBuffer, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create encrypting writer: %w", err)
	}
	if _, err := io.Copy(writer, msgBuffer); err != nil {
		return nil, fmt.Errorf("failed to copy encrypted data: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close encrypted writer: %w", err)
	}

	return encryptedBuffer.Bytes(), nil
}

func decryptBytesForStreaming(streamingPrimitive tink.StreamingAEAD, encrypted []byte) ([]byte, error) {
	encryptedBuffer := bytes.NewReader(encrypted)
	decryptedBuffer := &bytes.Buffer{}
	reader, err := streamingPrimitive.NewDecryptingReader(encryptedBuffer, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create decrypting reader: %w", err)
	}
	if _, err := io.Copy(decryptedBuffer, reader); err != nil {
		return nil, fmt.Errorf("failed to copy decrypted data: %w", err)
	}
	return decryptedBuffer.Bytes(), nil
}

type SubKey string

const (
	Logs  SubKey = "logs"
	Async SubKey = "async"
)

type EncryptorNext interface {
	Encrypt(subKey SubKey, cleartext []byte) ([]byte, error)
	Decrypt(subKey SubKey, encrypted []byte) ([]byte, error)
}

// NoOpEncryptorNext does not encrypt and just passes the input as is.
type NoOpEncryptorNext struct{}

func (n NoOpEncryptorNext) Encrypt(_ SubKey, cleartext []byte) ([]byte, error) {
	return cleartext, nil
}

func (n NoOpEncryptorNext) Decrypt(_ SubKey, encrypted []byte) ([]byte, error) {
	return encrypted, nil
}

type PlaintextEncryptor struct {
	root keyset.Handle
}

func NewPlaintextEncryptor(key string) (*PlaintextEncryptor, error) {
	handle, err := insecurecleartextkeyset.Read(
		keyset.NewJSONReader(bytes.NewBufferString(key)))
	if err != nil {
		return nil, fmt.Errorf("failed to read clear text keyset: %w", err)
	}

	return &PlaintextEncryptor{root: *handle}, nil
}

func (p PlaintextEncryptor) Encrypt(subKey SubKey, cleartext []byte) ([]byte, error) {
	encrypted, err := derivedEncrypt(p.root, subKey, cleartext)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt with derive: %w", err)
	}

	return encrypted, nil
}

func (p PlaintextEncryptor) Decrypt(subKey SubKey, encrypted []byte) ([]byte, error) {
	decrypted, err := derivedDecrypt(p.root, subKey, encrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt with derive: %w", err)
	}

	return decrypted, nil
}

// KMSEncryptor
// TODO: maybe change to DerivableEncryptor and integrate plaintext and kms encryptor.
type KMSEncryptor struct {
	root            keyset.Handle
	kekAEAD         tink.AEAD
	encryptedKeyset []byte
}

func newClientWithAEAD(uri string, kms *awsv1kms.KMS) (tink.AEAD, error) {
	var client registry.KMSClient
	var err error

	if strings.HasPrefix(strings.ToLower(uri), "fake-kms://") {
		client, err = fakekms.NewClient(uri)
		if err != nil {
			return nil, fmt.Errorf("failed to create fake KMS client: %w", err)
		}

	} else {
		// tink does not support awsv2 yet
		// https://github.com/tink-crypto/tink-go-awskms/issues/2
		var opts []awskms.ClientOption
		if kms != nil {
			opts = append(opts, awskms.WithKMS(kms))
		}
		client, err = awskms.NewClientWithOptions(uri, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create KMS client: %w", err)
		}
	}

	kekAEAD, err := client.GetAEAD(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to get aead: %w", err)
	}

	return kekAEAD, nil
}

func NewKMSEncryptorGenerateKey(uri string, v1client *awsv1kms.KMS) (*KMSEncryptor, error) {
	kekAEAD, err := newClientWithAEAD(uri, v1client)
	if err != nil {
		return nil, fmt.Errorf("failed to create KMS client: %w", err)
	}

	// Create a PRF key template using HKDF-SHA256
	prfKeyTemplate := prf.HKDFSHA256PRFKeyTemplate()

	// Create an AES-256-GCM key template
	aeadKeyTemplate := aead.AES256GCMKeyTemplate()

	template, err := keyderivation.CreatePRFBasedKeyTemplate(prfKeyTemplate, aeadKeyTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to create PRF based key template: %w", err)
	}

	handle, err := keyset.NewHandle(template)
	if err != nil {
		return nil, fmt.Errorf("failed to create keyset handle: %w", err)
	}

	// Encrypt the keyset with the KEK AEAD.
	buf := new(bytes.Buffer)
	writer := keyset.NewBinaryWriter(buf)
	err = handle.Write(writer, kekAEAD)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt DEK: %w", err)
	}
	encryptedKeyset := buf.Bytes()

	return &KMSEncryptor{
		root:            *handle,
		kekAEAD:         kekAEAD,
		encryptedKeyset: encryptedKeyset,
	}, nil
}

func NewKMSEncryptorWithKMS(uri string, v1client *awsv1kms.KMS, encryptedKeyset []byte) (*KMSEncryptor, error) {
	kekAEAD, err := newClientWithAEAD(uri, v1client)
	if err != nil {
		return nil, fmt.Errorf("failed to create KMS client: %w", err)
	}

	reader := keyset.NewBinaryReader(bytes.NewReader(encryptedKeyset))
	handle, err := keyset.Read(reader, kekAEAD)
	if err != nil {
		return nil, fmt.Errorf("failed to read keyset: %w", err)
	}

	return &KMSEncryptor{
		root:            *handle,
		kekAEAD:         kekAEAD,
		encryptedKeyset: encryptedKeyset,
	}, nil
}

func (k *KMSEncryptor) GetEncryptedKeyset() []byte {
	return k.encryptedKeyset
}

func deriveKeyset(root keyset.Handle, salt []byte) (*keyset.Handle, error) {
	deriver, err := keyderivation.New(&root)
	if err != nil {
		return nil, fmt.Errorf("failed to create deriver: %w", err)
	}

	derived, err := deriver.DeriveKeyset(salt)
	if err != nil {
		return nil, fmt.Errorf("failed to derive keyset: %w", err)
	}

	return derived, nil
}

func (k *KMSEncryptor) Encrypt(subKey SubKey, cleartext []byte) ([]byte, error) {
	encrypted, err := derivedEncrypt(k.root, subKey, cleartext)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt with derive: %w", err)
	}

	return encrypted, nil
}

func (k *KMSEncryptor) Decrypt(subKey SubKey, encrypted []byte) ([]byte, error) {
	decrypted, err := derivedDecrypt(k.root, subKey, encrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt with derive: %w", err)
	}

	return decrypted, nil
}

func derivedDecrypt(root keyset.Handle, subKey SubKey, encrypted []byte) ([]byte, error) {
	derived, err := deriveKeyset(root, []byte(subKey))
	if err != nil {
		return nil, fmt.Errorf("failed to derive keyset: %w", err)
	}

	primitive, err := aead.New(derived)
	if err != nil {
		return nil, fmt.Errorf("failed to create primitive: %w", err)
	}

	bytes, err := primitive.Decrypt(encrypted, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return bytes, nil
}

func derivedEncrypt(root keyset.Handle, subKey SubKey, cleartext []byte) ([]byte, error) {
	// TODO: Deriving might be expensive, consider caching the derived keyset.
	derived, err := deriveKeyset(root, []byte(subKey))
	if err != nil {
		return nil, fmt.Errorf("failed to derive keyset: %w", err)
	}

	primitive, err := aead.New(derived)
	if err != nil {
		return nil, fmt.Errorf("failed to create primitive: %w", err)
	}

	bytes, err := primitive.Encrypt(cleartext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt: %w", err)
	}

	return bytes, nil
}
