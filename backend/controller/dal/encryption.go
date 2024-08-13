package dal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/TBD54566975/ftl/backend/dal"
	"github.com/TBD54566975/ftl/internal/encryption"
)

func (d *DAL) encrypt(subKey encryption.SubKey, cleartext []byte) ([]byte, error) {
	if d.encryptor == nil {
		return nil, fmt.Errorf("encryptor not set")
	}

	v, err := d.encryptor.Encrypt(subKey, cleartext)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt: %w", err)
	}

	return v, nil
}

func (d *DAL) decrypt(subKey encryption.SubKey, encrypted []byte) ([]byte, error) {
	if d.encryptor == nil {
		return nil, fmt.Errorf("encryptor not set")
	}

	v, err := d.encryptor.Decrypt(subKey, encrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return v, nil
}

func (d *DAL) encryptJSON(subKey encryption.SubKey, v any) ([]byte, error) {
	serialized, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return d.encrypt(subKey, serialized)
}

func (d *DAL) decryptJSON(subKey encryption.SubKey, encrypted []byte, v any) error { //nolint:unparam
	decrypted, err := d.decrypt(subKey, encrypted)
	if err != nil {
		return fmt.Errorf("failed to decrypt: %w", err)
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
	tx, err := d.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.CommitOrRollback(ctx, &err)

	url, ok := d.kmsURL.Get()
	if !ok {
		d.encryptor = encryption.NewNoOpEncryptor()
		return nil
	}

	encryptedKey, err := tx.db.GetOnlyEncryptionKey(ctx)
	if err != nil {
		if dal.IsNotFound(err) {
			encryptor, err := encryption.NewKMSEncryptorGenerateKey(url, nil)
			if err != nil {
				return fmt.Errorf("failed to create encryptor for generation: %w", err)
			}
			d.encryptor = encryptor
			return nil
		}
		return fmt.Errorf("failed to get only encryption key: %w", err)
	}

	encryptor, err := encryption.NewKMSEncryptorWithKMS(url, nil, encryptedKey)
	if err != nil {
		return fmt.Errorf("failed to create encryptor with encrypted key: %w", err)
	}
	d.encryptor = encryptor

	if err = tx.db.CreateOnlyEncryptionKey(ctx, encryptedKey); err != nil {
		return fmt.Errorf("failed to create only encryption key: %w", err)
	}

	return nil
}
