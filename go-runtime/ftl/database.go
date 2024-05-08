package ftl

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"unsafe"

	_ "github.com/jackc/pgx/v5/stdlib" // Register Postgres driver

	"github.com/TBD54566975/ftl/internal/modulecontext"
)

type Database struct {
	Name   string
	DBType modulecontext.DBType
}

// PostgresDatabase returns a handler for the named database.
func PostgresDatabase(name string) Database {
	return Database{
		Name:   name,
		DBType: modulecontext.DBTypePostgres,
	}
}

func (d Database) String() string { return fmt.Sprintf("database %q", d.Name) }

// Get returns the sql db connection for the database.
func (d Database) Get(ctx context.Context) *sql.DB {
	provider := modulecontext.FromContext(ctx)
	db, err := provider.GetDatabase(d.Name, d.DBType)
	if err != nil {
		panic(err.Error())
	}
	return db
}

var _ HashableHandle[*sql.DB] = Database{}

func (d Database) Hash(ctx context.Context) []byte {
	// Convert the pointer to an integer
	ptrInt := uintptr(unsafe.Pointer(d.Get(ctx)))

	// Convert the integer to a byte slice
	ptrBytes := make([]byte, 8)
	for i := uint(0); i < 8; i++ {
		ptrBytes[i] = byte((ptrInt >> (i * 8)) & 0xff)
	}

	h := sha256.New()
	h.Write(ptrBytes)
	return h.Sum(nil)
}
