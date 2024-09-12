package encryption

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/TBD54566975/ftl/backend/controller/encryption/dal"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/internal/encryption"
)

type Service struct {
	encryptor encryption.DataEncryptor
}

func New(ctx context.Context, conn libdal.Connection, encryptionBuilder encryption.Builder) (*Service, error) {
	d := dal.New(ctx, conn)

	encryptor, err := encryptionBuilder.Build(ctx, d)
	if err != nil {
		return nil, fmt.Errorf("build encryptor: %w", err)
	}

	if err := d.VerifyEncryptor(ctx, encryptor); err != nil {
		return nil, fmt.Errorf("verify encryptor: %w", err)
	}

	return &Service{encryptor: encryptor}, nil
}

// EncryptJSON encrypts the given JSON object and stores it in the provided destination.
func (s *Service) EncryptJSON(v any, dest encryption.Encrypted) error {
	serialized, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return s.Encrypt(serialized, dest)
}

// DecryptJSON decrypts the given encrypted object and stores it in the provided destination.
func (s *Service) DecryptJSON(encrypted encryption.Encrypted, v any) error {
	decrypted, err := s.Decrypt(encrypted)
	if err != nil {
		return fmt.Errorf("failed to decrypt json with subkey %s: %w", encrypted.SubKey(), err)
	}

	if err = json.Unmarshal(decrypted, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

func (s *Service) Encrypt(cleartext []byte, dest encryption.Encrypted) error {
	err := s.encryptor.Encrypt(cleartext, dest)
	if err != nil {
		return fmt.Errorf("failed to encrypt binary with subkey %s: %w", dest.SubKey(), err)
	}

	return nil
}

func (s *Service) Decrypt(encrypted encryption.Encrypted) ([]byte, error) {
	v, err := s.encryptor.Decrypt(encrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt binary with subkey %s: %w", encrypted.SubKey(), err)
	}

	return v, nil
}
