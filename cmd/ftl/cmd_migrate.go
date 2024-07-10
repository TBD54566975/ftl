package main

import (
	"context"
	"fmt"
	"github.com/TBD54566975/ftl/internal/log"

	"github.com/TBD54566975/ftl/backend/controller/sql"
)

type migrateCmd struct {
	DSN string `help:"DSN for the database." default:"postgres://localhost:15432/ftl?sslmode=disable&user=postgres&password=secret" env:"DATABASE_URL"`
}

func (c *migrateCmd) Run(ctx context.Context) error {
	logger := log.FromContext(ctx)
	logger.Infof("Migrating database")
	err := sql.Migrate(ctx, c.DSN)
	return fmt.Errorf("failed to migrate database: %w", err)
}
