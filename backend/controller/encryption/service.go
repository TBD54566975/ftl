package encryption

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/encryption/api"
	"github.com/TBD54566975/ftl/backend/controller/encryption/dal"
	"github.com/TBD54566975/ftl/backend/libdal"
)

type Service struct {
	encryptor api.DataEncryptor
}

func New(ctx context.Context, conn libdal.Connection, encryptionBuilder Builder) (*Service, error) {
	d := dal.New(ctx, conn)

	encryptor, err := encryptionBuilder.Build(ctx, d)
	if err != nil {
		return nil, fmt.Errorf("build encryptor: %w", err)
	}

	if err := verifyEncryptor(ctx, d, encryptor); err != nil {
		return nil, fmt.Errorf("verify encryptor: %w", err)
	}

	return &Service{encryptor: encryptor}, nil
}

// EncryptJSON encrypts the given JSON object and stores it in the provided destination.
func (s *Service) EncryptJSON(v any, dest api.Encrypted) error {
	serialized, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return s.Encrypt(serialized, dest)
}

// DecryptJSON decrypts the given encrypted object and stores it in the provided destination.
func (s *Service) DecryptJSON(encrypted api.Encrypted, v any) error {
	decrypted, err := s.Decrypt(encrypted)
	if err != nil {
		return fmt.Errorf("failed to decrypt json with subkey %s: %w", encrypted.SubKey(), err)
	}

	if err = json.Unmarshal(decrypted, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

func (s *Service) Encrypt(cleartext []byte, dest api.Encrypted) error {
	err := s.encryptor.Encrypt(cleartext, dest)
	if err != nil {
		return fmt.Errorf("failed to encrypt binary with subkey %s: %w", dest.SubKey(), err)
	}

	return nil
}

func (s *Service) Decrypt(encrypted api.Encrypted) ([]byte, error) {
	v, err := s.encryptor.Decrypt(encrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt binary with subkey %s: %w", encrypted.SubKey(), err)
	}

	return v, nil
}

const verification = "FTL - Towards a 𝝺-calculus for large-scale systems"

func verifyEncryptor(ctx context.Context, d *dal.DAL, encryptor api.DataEncryptor) (err error) {
	tx, err := d.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.CommitOrRollback(ctx, &err)

	keys, err := tx.GetVerificationKeys(ctx)
	if err != nil {
		if libdal.IsNotFound(err) {
			if _, ok := encryptor.(NoOpEncryptor); ok {
				return nil
			}
			return fmt.Errorf("no encryption key found in the db for encryptor %T: %w", encryptor, err)
		}
		return fmt.Errorf("failed to get encryption row from the db: %w", err)
	}

	needsUpdate := false
	newTimeline, err := verifySubkey(encryptor, keys.VerifyTimeline)
	if err != nil {
		return fmt.Errorf("failed to verify timeline subkey: %w", err)
	}
	if newTimeline.Ok() {
		needsUpdate = true
		keys.VerifyTimeline = newTimeline
	}

	newAsync, err := verifySubkey(encryptor, keys.VerifyAsync)
	if err != nil {
		return fmt.Errorf("failed to verify async subkey: %w", err)
	}
	if newAsync.Ok() {
		needsUpdate = true
		keys.VerifyAsync = newAsync
	}

	if !needsUpdate {
		return nil
	}

	if !keys.VerifyTimeline.Ok() || !keys.VerifyAsync.Ok() {
		panic("should be unreachable. verifySubkey should have set the subkey")
	}

	err = tx.UpdateVerificationKeys(ctx, keys)
	if err != nil {
		return fmt.Errorf("failed to update encryption verification: %w", err)
	}

	return nil
}

// verifySubkey checks if the subkey is set and if not, sets it to a verification string.
// returns (nil, nil) if verified and not changed
func verifySubkey[SK api.SubKey](
	encryptor api.DataEncryptor,
	encrypted optional.Option[api.EncryptedColumn[SK]],
) (optional.Option[api.EncryptedColumn[SK]], error) {
	type EC = api.EncryptedColumn[SK]

	verifyField, ok := encrypted.Get()
	if !ok {
		err := encryptor.Encrypt([]byte(verification), &verifyField)
		if err != nil {
			return optional.None[EC](), fmt.Errorf("failed to encrypt verification sanity string: %w", err)
		}
		return optional.Some(verifyField), nil
	}

	decrypted, err := encryptor.Decrypt(&verifyField)
	if err != nil {
		return optional.None[EC](), fmt.Errorf("failed to decrypt verification sanity string: %w", err)
	}

	if string(decrypted) != verification {
		return optional.None[EC](), fmt.Errorf("decrypted verification string does not match expected value")
	}

	// verified, no need to update
	return optional.None[EC](), nil
}
