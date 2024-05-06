package modulecontext

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// DatabasesFromEnvironment finds DSNs in environment variables and creates a map of databases.
func DatabasesFromEnvironment(ctx context.Context, module string) (map[string]Database, error) {
	databases := map[string]Database{}
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
		dbName = strings.ToLower(dbName)
		db, err := NewDatabase(DBTypePostgres, value)
		if err != nil {
			return nil, fmt.Errorf("could not create database %q with DSN %q: %w", dbName, value, err)
		}
		databases[dbName] = db
	}
	return databases, nil
}
