package main

import (
	"context"
	"github.com/TBD54566975/ftl/internal/log"

	"github.com/TBD54566975/ftl/backend/controller/sql"
)

type migrateCmd struct {
	DSN string `help:"DSN for the database." default:"postgres://localhost:15432/ftl?sslmode=disable&user=postgres&password=secret" env:"DATABASE_URL"`
}

func (c *migrateCmd) Run(ctx context.Context) error {
	logger := log.FromContext(ctx)
	logger.Infof("Migrating database")
	return sql.Migrate(ctx, c.DSN)
}
