package dal

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/TBD54566975/ftl/backend/dal"
	"github.com/TBD54566975/ftl/internal/encryption"
	"github.com/TBD54566975/ftl/internal/log"
)

func (d *DAL) encrypt(cleartext []byte, dest encryption.Encrypted) error {
	if d.encryptor == nil {
		return fmt.Errorf("encryptor not set")
	}

	err := d.encryptor.Encrypt(cleartext, dest)
	if err != nil {
		return fmt.Errorf("failed to encrypt binary with subkey %s: %w", dest.SubKey(), err)
	}

	return nil
}

func (d *DAL) decrypt(encrypted encryption.Encrypted) ([]byte, error) {
	if d.encryptor == nil {
		return nil, fmt.Errorf("encryptor not set")
	}

	v, err := d.encryptor.Decrypt(encrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt binary with subkey %s: %w", encrypted.SubKey(), err)
	}

	return v, nil
}

func (d *DAL) encryptJSON(v any, dest encryption.Encrypted) error {
	serialized, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return d.encrypt(serialized, dest)
}

func (d *DAL) decryptJSON(encrypted encryption.Encrypted, v any) error { //nolint:unparam
	decrypted, err := d.decrypt(encrypted)
	if err != nil {
		return fmt.Errorf("failed to decrypt json with subkey %s: %w", encrypted.SubKey(), err)
	}

	if err = json.Unmarshal(decrypted, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

// setupEncryptor sets up the encryptor for the DAL.
// It will either create a key or load the existing one.
// If the KMS URL is not set, it will use a NoOpEncryptor which does not encrypt anything.
func (d *DAL) setupEncryptor(ctx context.Context) (err error) {
	logger := log.FromContext(ctx)
	tx, err := d.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.CommitOrRollback(ctx, &err)

	url, ok := d.kmsURL.Get()
	if !ok {
		logger.Infof("KMS URL not set, encryption not enabled")
		d.encryptor = encryption.NewNoOpEncryptor()
		return nil
	}

	encryptedKey, err := tx.db.GetOnlyEncryptionKey(ctx)
	if err != nil {
		if dal.IsNotFound(err) {
			logger.Infof("No encryption key found, generating a new one")
			encryptor, err := encryption.NewKMSEncryptorGenerateKey(url, nil)
			if err != nil {
				return fmt.Errorf("failed to create encryptor for generation: %w", err)
			}
			d.encryptor = encryptor

			if err = tx.db.CreateOnlyEncryptionKey(ctx, encryptor.GetEncryptedKeyset()); err != nil {
				return fmt.Errorf("failed to create only encryption key: %w", err)
			}

			return nil
		}
		return fmt.Errorf("failed to get only encryption key: %w", err)
	}

	logger.Debugf("Encryption key found, using it")
	encryptor, err := encryption.NewKMSEncryptorWithKMS(url, nil, encryptedKey)
	if err != nil {
		return fmt.Errorf("failed to create encryptor with encrypted key: %w", err)
	}
	d.encryptor = encryptor

	return nil
}
