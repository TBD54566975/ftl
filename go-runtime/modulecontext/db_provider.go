package modulecontext

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib" // SQL driver

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
)

type DBType ftlv1.ModuleContextResponse_DBType

const (
	DBTypePostgres = DBType(ftlv1.ModuleContextResponse_POSTGRES)
)

func (x DBType) String() string {
	switch x {
	case DBTypePostgres:
		return "Postgres"
	default:
		panic(fmt.Sprintf("unknown DB type: %s", strconv.Itoa(int(x))))
	}
}

type dbEntry struct {
	dsn    string
	dbType DBType
	db     *sql.DB
}

// DBProvider takes in DSNs and holds a *sql.DB for each
// this allows us to:
// - pool db connections, rather than initializing anew each time
// - validate DSNs at startup, rather than returning errors or panicking at Database.Get()
type DBProvider struct {
	entries map[string]dbEntry
}

type contextKeyDSNProvider struct{}

func NewDBProvider() *DBProvider {
	return &DBProvider{
		entries: map[string]dbEntry{},
	}
}

// NewDBProviderFromEnvironment creates a new DBProvider from environment variables.
//
// This is a temporary measure until we have a way to load DSNs from the ftl-project.toml file.
func NewDBProviderFromEnvironment(module string) (*DBProvider, error) {
	// TODO: Replace this with loading DSNs from ftl-project.toml.
	dbProvider := NewDBProvider()
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "FTL_POSTGRES_DSN_") {
			continue
		}
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid DSN environment variable: %s", entry)
		}
		key := parts[0]
		value := parts[1]
		// FTL_POSTGRES_DSN_MODULE_DBNAME
		parts = strings.Split(key, "_")
		if len(parts) != 5 {
			return nil, fmt.Errorf("invalid DSN environment variable: %s", entry)
		}
		moduleName := parts[3]
		dbName := parts[4]
		if !strings.EqualFold(moduleName, module) {
			continue
		}
		if err := dbProvider.Add(strings.ToLower(dbName), DBTypePostgres, value); err != nil {
			return nil, err
		}
	}
	return dbProvider, nil
}

func ContextWithDBProvider(ctx context.Context, provider *DBProvider) context.Context {
	return context.WithValue(ctx, contextKeyDSNProvider{}, provider)
}

func DBProviderFromContext(ctx context.Context) *DBProvider {
	m, ok := ctx.Value(contextKeyDSNProvider{}).(*DBProvider)
	if !ok {
		panic("no db provider in context")
	}
	return m
}

func (d *DBProvider) Add(name string, dbType DBType, dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	d.entries[name] = dbEntry{
		dsn:    dsn,
		db:     db,
		dbType: dbType,
	}
	return nil
}

func (d *DBProvider) Get(name string) (*sql.DB, error) {
	if entry, ok := d.entries[name]; ok {
		return entry.db, nil
	}
	return nil, fmt.Errorf("missing DSN for database %s", name)
}
