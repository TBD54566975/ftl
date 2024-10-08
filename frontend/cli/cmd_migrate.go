package main

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/backend/controller/sql"
	"github.com/TBD54566975/ftl/internal/log"
)

type migrateCmd struct {
	DSN string `help:"DSN for the database." default:"${dsn}" env:"DATABASE_URL"`
}

func (c *migrateCmd) Run(ctx context.Context) error {
	logger := log.FromContext(ctx)
	logger.Infof("Migrating database")
	err := sql.Migrate(ctx, c.DSN, log.Info)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	return nil
}
