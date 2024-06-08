package modulecontext

import (
	"context"
	"fmt"
	"strings"
)

// DatabasesFromSecrets finds DSNs in secrets and creates a map of databases.
//
// Secret keys should be in the format FTL_DSN_<MODULENAME>_<DBNAME>
func DatabasesFromSecrets(ctx context.Context, module string, secrets map[string][]byte) (map[string]Database, error) {
	databases := map[string]Database{}
	for sName, maybeDSN := range secrets {
		if !strings.HasPrefix(sName, "FTL_DSN_") {
			continue
		}
		// FTL_DSN_<MODULE>_<DBNAME>
		parts := strings.Split(sName, "_")
		if len(parts) != 4 {
			return nil, fmt.Errorf("invalid DSN secret key %q should have format FTL_DSN_<MODULE>_<DBNAME>", sName)
		}
		moduleName := strings.ToLower(parts[2])
		dbName := strings.ToLower(parts[3])
		if !strings.EqualFold(moduleName, module) {
			continue
		}
		dsnStr := string(maybeDSN)
		dsn := dsnStr[1 : len(dsnStr)-1] // chop leading + trailing quotes
		db, err := NewDatabase(DBTypePostgres, dsn)
		if err != nil {
			return nil, fmt.Errorf("could not create database %q with DSN %q: %w", dbName, maybeDSN, err)
		}
		databases[dbName] = db
	}
	return databases, nil
}
