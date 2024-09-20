package dal

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/backend/controller/identity/dal/internal/sql"
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

type EncryptedIdentity = sql.GetOnlyIdentityKeyRow

func (d *DAL) GetOnlyIdentityKey(ctx context.Context) (*EncryptedIdentity, error) {
	row, err := d.db.GetOnlyIdentityKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get only identity key: %w", err)
	}

	return &row, nil
}

func (d *DAL) CreateOnlyIdentityKey(ctx context.Context, e EncryptedIdentity) error {
	if err := d.db.CreateOnlyIdentityKey(ctx, e.Private, e.Public, e.VerifySignature); err != nil {
		return fmt.Errorf("failed to create only identity key: %w", err)
	}

	return nil
}
