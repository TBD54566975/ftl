// Package migrate supports a dbmate-compatible superset of migration files.
//
// The superset is that in addition to a migration being a .sql file, it can
// also be a Go function which is called to execute the migration.
package migrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/TBD54566975/ftl/backend/dal"
	"github.com/TBD54566975/ftl/internal/log"
)

var migrationFileNameRe = regexp.MustCompile(`^.*(\d{14})_(.*)(\.sql)?$`)

type migrateOptions struct {
	logLevel   log.Level
	migrations map[string]MigrationFunc
}

// Option is a configuration option for Migrate.
type Option func(*migrateOptions)

// Migration adds a named migration function to the migration set.
//
// "version" must be in the form "<YYYY><MM><DD><hh><mm><ss>".
func Migration(version, name string, migration MigrationFunc) Option {
	return func(opts *migrateOptions) {
		opts.migrations[version+"_"+name] = migration
	}
}

// LogLevel sets the loggging level of the migrator.
func LogLevel(level log.Level) Option {
	return func(opts *migrateOptions) {
		opts.logLevel = level
	}
}

type MigrationFunc func(ctx context.Context, db *sql.Tx) error

type namedMigration struct {
	name      string
	version   string
	migration MigrationFunc
}

func (m namedMigration) String() string { return m.name }

// Migrate applies all migrations in the provided fs.FS and migration functions
// to the provided database.
func Migrate(ctx context.Context, db *sql.DB, migrationFiles fs.FS, options ...Option) error {
	// Create schema_migrations table if it doesn't exist.
	// This table structure is compatible with dbmate.
	_, _ = db.ExecContext(ctx, `CREATE TABLE schema_migrations (version TEXT PRIMARY KEY)`) //nolint:errcheck

	sqlFiles, err := fs.Glob(migrationFiles, "*.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	opts := migrateOptions{
		logLevel:   log.Debug,
		migrations: make(map[string]MigrationFunc),
	}
	for _, opt := range options {
		opt(&opts)
	}

	migrations := make([]namedMigration, 0, len(sqlFiles)+len(opts.migrations))

	// Collect .sql files.
	for _, sqlFile := range sqlFiles {
		name := filepath.Base(sqlFile)
		groups := migrationFileNameRe.FindStringSubmatch(name)
		if groups == nil {
			return fmt.Errorf("invalid migration file name %q, must be in the form <YYYY><MM><DD><hh><mm><ss>_<detail>.sql", sqlFile)
		}
		version := groups[1]
		migrations = append(migrations, namedMigration{name, version, func(ctx context.Context, db *sql.Tx) error {
			sqlMigration, err := fs.ReadFile(migrationFiles, sqlFile)
			if err != nil {
				return fmt.Errorf("failed to read migration file %q: %w", sqlFile, err)
			}
			return migrateSQLFile(ctx, db, sqlFile, sqlMigration)
		}})
	}
	for name, migration := range opts.migrations {
		groups := migrationFileNameRe.FindStringSubmatch(name)
		if groups == nil {
			return fmt.Errorf("invalid migration name %q, must be in the form <YYYY><MM><DD><hh><mm><ss>_<detail>", name)
		}
		version := groups[1]
		migrations = append(migrations, namedMigration{name, version, migration})
	}
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})
	for _, migration := range migrations {
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("migration %s: failed to begin transaction: %w", migration, err)
		}
		err = applyMigration(ctx, opts.logLevel, tx, migration)
		if err != nil {
			if txerr := tx.Rollback(); txerr != nil {
				return fmt.Errorf("migration %s: failed to rollback transaction: %w", migration, txerr)
			}
			return fmt.Errorf("migration %s: %w", migration, err)
		}
		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("migration %s: failed to commit transaction: %w", migration, err)
		}
	}
	return nil
}

func applyMigration(ctx context.Context, level log.Level, tx *sql.Tx, migration namedMigration) error {
	start := time.Now()
	logger := log.FromContext(ctx).Scope("migrate")
	_, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", migration.version)
	err = dal.TranslatePGError(err)
	if errors.Is(err, dal.ErrConflict) {
		if txerr := tx.Rollback(); txerr != nil {
			return fmt.Errorf("failed to rollback transaction: %w", txerr)
		}
		logger.Logf(level, "Skipping: %s", migration)
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to insert migration: %w", err)
	}
	logger.Logf(level, "Applying: %s", migration)
	if err := migration.migration(ctx, tx); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	logger.Logf(level, "Applied: %s in %s", migration, time.Since(start))
	return nil
}

func migrateSQLFile(ctx context.Context, db *sql.Tx, name string, sqlMigration []byte) error {
	_, err := db.ExecContext(ctx, string(sqlMigration))
	if err != nil {
		return fmt.Errorf("failed to execute migration %q: %w", name, err)
	}
	return nil
}
