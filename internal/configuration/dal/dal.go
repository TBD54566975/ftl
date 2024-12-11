// Package dal provides a data abstraction layer for managing module configurations
package dal

import (
	"context"
	"fmt"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/common/slices"
	"github.com/TBD54566975/ftl/internal/configuration/dal/internal/sql"
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

func (d *DAL) GetModuleConfiguration(ctx context.Context, module optional.Option[string], name string) ([]byte, error) {
	b, err := d.db.GetModuleConfiguration(ctx, module, name)
	if err != nil {
		return nil, libdal.TranslatePGError(err)
	}
	return b, nil
}

func (d *DAL) SetModuleConfiguration(ctx context.Context, module optional.Option[string], name string, value []byte) error {
	err := d.db.SetModuleConfiguration(ctx, module, name, value)
	return libdal.TranslatePGError(err)
}

func (d *DAL) UnsetModuleConfiguration(ctx context.Context, module optional.Option[string], name string) error {
	err := d.db.UnsetModuleConfiguration(ctx, module, name)
	return libdal.TranslatePGError(err)
}

type ModuleConfiguration sql.ModuleConfiguration

func (d *DAL) ListModuleConfiguration(ctx context.Context) ([]ModuleConfiguration, error) {
	l, err := d.db.ListModuleConfiguration(ctx)
	if err != nil {
		return nil, libdal.TranslatePGError(err)
	}
	return slices.Map(l, func(t sql.ModuleConfiguration) ModuleConfiguration {
		return ModuleConfiguration(t)
	}), nil
}

func (d *DAL) GetModuleSecretURL(ctx context.Context, module optional.Option[string], name string) (string, error) {
	b, err := d.db.GetModuleSecretURL(ctx, module, name)
	if err != nil {
		return "", fmt.Errorf("could not get secret URL: %w", libdal.TranslatePGError(err))
	}
	return b, nil
}

func (d *DAL) SetModuleSecretURL(ctx context.Context, module optional.Option[string], name string, url string) error {
	err := d.db.SetModuleSecretURL(ctx, module, name, url)
	if err != nil {
		return fmt.Errorf("could not set secret URL: %w", libdal.TranslatePGError(err))
	}
	return nil
}

func (d *DAL) UnsetModuleSecret(ctx context.Context, module optional.Option[string], name string) error {
	err := d.db.UnsetModuleSecret(ctx, module, name)
	if err != nil {
		return fmt.Errorf("could not unset secret: %w", libdal.TranslatePGError(err))
	}
	return nil
}

type ModuleSecret sql.ModuleSecret

func (d *DAL) ListModuleSecrets(ctx context.Context) ([]ModuleSecret, error) {
	l, err := d.db.ListModuleSecrets(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not list secrets: %w", libdal.TranslatePGError(err))
	}

	// Convert []sql.ModuleSecret to []ModuleSecret
	ms := make([]ModuleSecret, len(l))
	for i, secret := range l {
		ms[i] = ModuleSecret(secret)
	}

	return ms, nil
}
