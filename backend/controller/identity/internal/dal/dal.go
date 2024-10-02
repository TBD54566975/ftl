package dal

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/backend/controller/identity/internal/sql"
	"github.com/TBD54566975/ftl/backend/libdal"
)

type DAL struct {
	*libdal.Handle[DAL]
	db sql.Querier
}

func New(conn libdal.Connection) *DAL {
	return &DAL{
		db: sql.New(conn),
		Handle: libdal.New(conn, func(h *libdal.Handle[DAL]) *DAL {
			return &DAL{Handle: h, db: sql.New(h.Connection)}
		}),
	}
}

type EncryptedIdentity = sql.GetIdentityKeysRow

func (d *DAL) GetOnlyIdentityKey(ctx context.Context) (EncryptedIdentity, error) {
	rows, err := d.db.GetIdentityKeys(ctx)
	if err != nil {
		return EncryptedIdentity{}, fmt.Errorf("failed to get only identity key: %w", err)
	}
	if len(rows) == 0 {
		return EncryptedIdentity{}, libdal.ErrNotFound
	}
	if len(rows) > 1 {
		return EncryptedIdentity{}, fmt.Errorf("too many identity keys found: %d", len(rows))
	}

	return rows[0], nil
}

func (d *DAL) CreateOnlyIdentityKey(ctx context.Context, e EncryptedIdentity) error {
	if err := d.db.CreateOnlyIdentityKey(ctx, e.Private, e.Public, e.VerifySignature); err != nil {
		return fmt.Errorf("failed to create only identity key: %w", err)
	}

	return nil
}
