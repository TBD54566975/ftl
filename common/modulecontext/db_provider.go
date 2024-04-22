package modulecontext

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"

	_ "github.com/jackc/pgx/v5/stdlib" // SQL driver
)

type DBType int32

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

func (d *DBProvider) AddDSN(name string, dbType DBType, dsn string) error {
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

func (d *DBProvider) GetDB(name string) (*sql.DB, error) {
	if entry, ok := d.entries[name]; ok {
		return entry.db, nil
	}
	return nil, fmt.Errorf("missing DSN for database %s", name)
}
