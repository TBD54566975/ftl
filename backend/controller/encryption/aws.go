package encryption

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/tink-crypto/tink-go-awskms/integration/awskms"
	"github.com/tink-crypto/tink-go/v2/aead"
	"github.com/tink-crypto/tink-go/v2/core/registry"
	"github.com/tink-crypto/tink-go/v2/keyderivation"
	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/tink-crypto/tink-go/v2/prf"
	"github.com/tink-crypto/tink-go/v2/testing/fakekms"
	"github.com/tink-crypto/tink-go/v2/tink"

	"github.com/TBD54566975/ftl/backend/controller/encryption/api"
	"github.com/TBD54566975/ftl/internal/mutex"
)

// KMSEncryptor encrypts and decrypts using a KMS key via tink.
type KMSEncryptor struct {
	root            keyset.Handle
	kekAEAD         tink.AEAD
	encryptedKeyset []byte
	cachedDerived   *mutex.Mutex[map[api.SubKey]tink.AEAD]
}

var _ api.DataEncryptor = &KMSEncryptor{}

func newClientWithAEAD(uri string, kms *kms.KMS) (tink.AEAD, error) {
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

func newKey(uri string, v1client *kms.KMS) ([]byte, error) {
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
	return buf.Bytes(), nil
}

func NewKMSEncryptorWithKMS(uri string, v1client *kms.KMS, encryptedKeyset []byte) (*KMSEncryptor, error) {
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
		cachedDerived:   mutex.New(map[api.SubKey]tink.AEAD{}),
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

func (k *KMSEncryptor) getDerivedPrimitive(subKey api.SubKey) (tink.AEAD, error) {
	cachedDerived := k.cachedDerived.Lock()
	defer k.cachedDerived.Unlock()
	primitive, ok := cachedDerived[subKey]
	if ok {
		return primitive, nil
	}

	derived, err := deriveKeyset(k.root, []byte(subKey.SubKey()))
	if err != nil {
		return nil, fmt.Errorf("failed to derive keyset: %w", err)
	}

	primitive, err = aead.New(derived)
	if err != nil {
		return nil, fmt.Errorf("failed to create primitive: %w", err)
	}

	cachedDerived[subKey] = primitive
	return primitive, nil
}

func (k *KMSEncryptor) Encrypt(cleartext []byte, dest api.Encrypted) error {
	primitive, err := k.getDerivedPrimitive(dest)
	if err != nil {
		return fmt.Errorf("%s: failed to get derived primitive: %w", dest.SubKey(), err)
	}

	encrypted, err := primitive.Encrypt(cleartext, nil)
	if err != nil {
		return fmt.Errorf("%s: failed to encrypt: %w", dest.SubKey(), err)
	}

	dest.Set(encrypted)
	return nil
}

func (k *KMSEncryptor) Decrypt(encrypted api.Encrypted) ([]byte, error) {
	primitive, err := k.getDerivedPrimitive(encrypted)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get derived primitive: %w", encrypted.SubKey(), err)
	}

	decrypted, err := primitive.Decrypt(encrypted.Bytes(), nil)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to decrypt: %w", encrypted.SubKey(), err)
	}

	return decrypted, nil
}
