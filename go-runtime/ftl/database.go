package ftl

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/alecthomas/types/once"
	_ "github.com/go-sql-driver/mysql" // Register MySQL driver
	_ "github.com/jackc/pgx/v5/stdlib" // Register Postgres driver
)

type DatabaseConfig interface {
	// Name returns the name of the database.
	Name() string
	/*
		// RAM returns the amount of memory (in bytes) allocated to the database.
		RAM() int64
		// Disk returns the path or identifier for the disk where the database data is stored.
		Disk() string
		// Timeout returns the timeout value (in milliseconds) for database operations, such as queries or connections.
		Timeout() int64
		// MaxConnections returns the maximum number of concurrent database connections allowed.
		MaxConnections() int
	*/
	db()
}

type PostgresDatabaseConfig interface {
	DatabaseConfig
	pg()
}

// DefaultPostgresDatabaseConfig is a default implementation of PostgresDatabaseConfig. It does not provide
// an implementation for the Name method and should be embedded in a struct that does.
type DefaultPostgresDatabaseConfig struct{}

func (DefaultPostgresDatabaseConfig) db() {} //nolint:unused
func (DefaultPostgresDatabaseConfig) pg() {} //nolint:unused

type MySQLDatabaseConfig interface {
	DatabaseConfig
	mysql()
}

// DefaultMySQLDatabaseConfig is a default implementation of MySQLDatabaseConfig. It does not provide
// an implementation for the Name method and should be embedded in a struct that does.
type DefaultMySQLDatabaseConfig struct{}

func (DefaultMySQLDatabaseConfig) db()    {} //nolint:unused
func (DefaultMySQLDatabaseConfig) mysql() {} //nolint:unused

type DatabaseType string

const (
	DatabaseTypePostgres DatabaseType = "postgres"
	DatabaseTypeMysql    DatabaseType = "mysql"
)

type DatabaseHandle[T DatabaseConfig] struct {
	name  string
	_type DatabaseType
	db    *once.Handle[*sql.DB]
}

// Name returns the name of the database.
func (d DatabaseHandle[T]) Name() string { return d.name }

// Type returns the type of the database, e.g. "postgres"
func (d DatabaseHandle[T]) Type() DatabaseType {
	return d._type
}

// String returns a string representation of the database handle.
func (d DatabaseHandle[T]) String() string {
	return fmt.Sprintf("database %q", d.name)
}

// Get returns the SQL DB connection for the database.
func (d DatabaseHandle[T]) Get(ctx context.Context) *sql.DB {
	db, err := d.db.Get(ctx)
	if err != nil {
		panic(err)
	}
	return db
}

// NewDatabaseHandle is managed by FTL.
func NewDatabaseHandle[T DatabaseConfig](config T, dbType DatabaseType, db *once.Handle[*sql.DB]) DatabaseHandle[T] {
	return DatabaseHandle[T]{name: config.Name(), db: db, _type: dbType}
}
