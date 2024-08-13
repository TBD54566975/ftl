package encryption

import (
	"bytes"
	"fmt"
	"strings"

	awsv1kms "github.com/aws/aws-sdk-go/service/kms"
	"github.com/tink-crypto/tink-go-awskms/integration/awskms"
	"github.com/tink-crypto/tink-go/v2/aead"
	"github.com/tink-crypto/tink-go/v2/core/registry"
	"github.com/tink-crypto/tink-go/v2/keyderivation"
	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/tink-crypto/tink-go/v2/prf"
	"github.com/tink-crypto/tink-go/v2/testing/fakekms"
	"github.com/tink-crypto/tink-go/v2/tink"
)

type SubKey string

const (
	LogsSubKey  SubKey = "logs"
	AsyncSubKey SubKey = "async"
)

type DataEncryptor interface {
	Encrypt(subKey SubKey, cleartext []byte) ([]byte, error)
	Decrypt(subKey SubKey, encrypted []byte) ([]byte, error)
}

// NoOpEncryptorNext does not encrypt and just passes the input as is.
type NoOpEncryptorNext struct{}

func NewNoOpEncryptor() NoOpEncryptorNext {
	return NoOpEncryptorNext{}
}

func (n NoOpEncryptorNext) Encrypt(_ SubKey, cleartext []byte) ([]byte, error) {
	return cleartext, nil
}

func (n NoOpEncryptorNext) Decrypt(_ SubKey, encrypted []byte) ([]byte, error) {
	return encrypted, nil
}

// KMSEncryptor encrypts and decrypts using a KMS key via tink.
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
