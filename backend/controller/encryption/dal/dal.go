package dal

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/backend/controller/encryption/api"
	"github.com/TBD54566975/ftl/backend/controller/encryption/dal/internal/sql"
	"github.com/TBD54566975/ftl/backend/libdal"
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

// VerificationKeys contains the verification keys for the timeline and async encryption.
type VerificationKeys struct {
	VerifyTimeline api.OptionalEncryptedTimelineColumn
	VerifyAsync    api.OptionalEncryptedAsyncColumn
}

func (d *DAL) GetVerificationKeys(ctx context.Context) (keys VerificationKeys, err error) {
	tx, err := d.Begin(ctx)
	if err != nil {
		return VerificationKeys{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.CommitOrRollback(ctx, &err)

	row, err := tx.db.GetOnlyEncryptionKey(ctx)
	if err != nil {
		return VerificationKeys{}, fmt.Errorf("failed to get encryption key from the db: %w", err)
	}

	return VerificationKeys{
		VerifyTimeline: row.VerifyTimeline,
		VerifyAsync:    row.VerifyAsync,
	}, nil
}

func (d *DAL) UpdateVerificationKeys(ctx context.Context, keys VerificationKeys) (err error) {
	tx, err := d.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.CommitOrRollback(ctx, &err)

	err = tx.db.UpdateEncryptionVerification(ctx, keys.VerifyTimeline, keys.VerifyAsync)
	if err != nil {
		return fmt.Errorf("failed to update encryption verification: %w", err)
	}

	return nil
}
