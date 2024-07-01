// Package dal provides a data abstraction layer for managing module configurations
package dal

import (
	"context"
	"fmt"

	"github.com/alecthomas/types/optional"
	"github.com/jackc/pgx/v5/pgxpool"

	dalerrs "github.com/TBD54566975/ftl/backend/dal"
	"github.com/TBD54566975/ftl/common/configuration/sql"
)

type DAL struct {
	db sql.DBI
}

func New(ctx context.Context, pool *pgxpool.Pool) (*DAL, error) {
	dal := &DAL{db: sql.NewDB(pool)}
	return dal, nil
}

func (d *DAL) GetModuleConfiguration(ctx context.Context, module optional.Option[string], name string) ([]byte, error) {
	b, err := d.db.GetModuleConfiguration(ctx, module, name)
	if err != nil {
		return nil, dalerrs.TranslatePGError(err)
	}
	return b, nil
}

func (d *DAL) SetModuleConfiguration(ctx context.Context, module optional.Option[string], name string, value []byte) error {
	err := d.db.SetModuleConfiguration(ctx, module, name, value)
	return dalerrs.TranslatePGError(err)
}

func (d *DAL) UnsetModuleConfiguration(ctx context.Context, module optional.Option[string], name string) error {
	err := d.db.UnsetModuleConfiguration(ctx, module, name)
	return dalerrs.TranslatePGError(err)
}

func (d *DAL) ListModuleConfiguration(ctx context.Context) ([]sql.ModuleConfiguration, error) {
	l, err := d.db.ListModuleConfiguration(ctx)
	if err != nil {
		return nil, dalerrs.TranslatePGError(err)
	}
	return l, nil
}

func (d *DAL) GetModuleSecretURL(ctx context.Context, module optional.Option[string], name string) (string, error) {
	b, err := d.db.GetModuleSecretURL(ctx, module, name)
	if err != nil {
		return "", fmt.Errorf("could not get secret URL: %w", dalerrs.TranslatePGError(err))
	}
	return b, nil
}

func (d *DAL) SetModuleSecretURL(ctx context.Context, module optional.Option[string], name string, url string) error {
	err := d.db.SetModuleSecretURL(ctx, module, name, url)
	if err != nil {
		return fmt.Errorf("could not set secret URL: %w", dalerrs.TranslatePGError(err))
	}
	return nil
}

func (d *DAL) UnsetModuleSecret(ctx context.Context, module optional.Option[string], name string) error {
	err := d.db.UnsetModuleSecret(ctx, module, name)
	if err != nil {
		return fmt.Errorf("could not unset secret: %w", dalerrs.TranslatePGError(err))
	}
	return nil
}

func (d *DAL) ListModuleSecrets(ctx context.Context) ([]sql.ModuleSecret, error) {
	l, err := d.db.ListModuleSecrets(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not list secrets: %w", dalerrs.TranslatePGError(err))
	}
	return l, nil
}
