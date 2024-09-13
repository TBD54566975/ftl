package dal

import (
	"context"
	"fmt"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/encryption/dal/internal/sql"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/internal/encryption"
	"github.com/TBD54566975/ftl/internal/log"
)

type DAL struct {
	*libdal.Handle[DAL]
	db sql.Querier
}

func New(ctx context.Context, conn libdal.Connection) *DAL {
	return &DAL{
		Handle: libdal.New(conn, func(h *libdal.Handle[DAL]) *DAL {
			return &DAL{
				Handle: h,
				db:     sql.New(h.Connection),
			}
		}),
		db: sql.New(conn),
	}
}

func (d *DAL) EnsureKey(ctx context.Context, generateKey func() ([]byte, error)) (encryptedKey []byte, err error) {
	logger := log.FromContext(ctx)
	tx, err := d.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.CommitOrRollback(ctx, &err)

	var key []byte
	row, err := tx.db.GetOnlyEncryptionKey(ctx)
	if err != nil && libdal.IsNotFound(err) {
		logger.Debugf("No encryption key found, generating a new one")
		key, err = generateKey()
		if err != nil {
			return nil, fmt.Errorf("failed to generate key: %w", err)
		}

		if err = tx.db.CreateOnlyEncryptionKey(ctx, key); err != nil {
			return nil, fmt.Errorf("failed to save the encryption key: %w", err)
		}

		return key, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to load the encryption key from the db: %w", err)
	}

	logger.Debugf("Encryption key found, using it")

	return row.Key, nil
}

const verification = "FTL - Towards a 𝝺-calculus for large-scale systems"

func (d *DAL) VerifyEncryptor(ctx context.Context, encryptor encryption.DataEncryptor) (err error) {
	tx, err := d.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.CommitOrRollback(ctx, &err)

	row, err := tx.db.GetOnlyEncryptionKey(ctx)
	if err != nil {
		if libdal.IsNotFound(err) {
			// No encryption key found, probably using noop.
			return nil
		}
		return fmt.Errorf("failed to get encryption row from the db: %w", err)
	}

	needsUpdate := false
	newTimeline, err := verifySubkey(encryptor, row.VerifyTimeline)
	if err != nil {
		return fmt.Errorf("failed to verify timeline subkey: %w", err)
	}
	if newTimeline != nil {
		needsUpdate = true
		row.VerifyTimeline = optional.Some(newTimeline)
	}

	newAsync, err := verifySubkey(encryptor, row.VerifyAsync)
	if err != nil {
		return fmt.Errorf("failed to verify async subkey: %w", err)
	}
	if newAsync != nil {
		needsUpdate = true
		row.VerifyAsync = optional.Some(newAsync)
	}

	if !needsUpdate {
		return nil
	}

	if !row.VerifyTimeline.Ok() || !row.VerifyAsync.Ok() {
		panic("should be unreachable. verifySubkey should have set the subkey")
	}

	err = tx.db.UpdateEncryptionVerification(ctx, row.VerifyTimeline, row.VerifyAsync)
	if err != nil {
		return fmt.Errorf("failed to update encryption verification: %w", err)
	}

	return nil
}

// verifySubkey checks if the subkey is set and if not, sets it to a verification string.
// returns (nil, nil) if verified and not changed
func verifySubkey[SK encryption.SubKey](encryptor encryption.DataEncryptor, encrypted optional.Option[encryption.EncryptedColumn[SK]]) (encryption.EncryptedColumn[SK], error) {
	verifyField, ok := encrypted.Get()
	if !ok {
		err := encryptor.Encrypt([]byte(verification), &verifyField)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt verification sanity string: %w", err)
		}
		return verifyField, nil
	}

	decrypted, err := encryptor.Decrypt(&verifyField)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt verification sanity string: %w", err)
	}

	if string(decrypted) != verification {
		return nil, fmt.Errorf("decrypted verification string does not match expected value")
	}

	// verified, no need to update
	return nil, nil
}
