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

// DSNSecretKey returns the key for the secret that is expected to hold the DSN for a database.
//
// The format is FTL_DSN_<MODULE>_<DBNAME>
func DSNSecretKey(module, name string) string {
	return fmt.Sprintf("FTL_DSN_%s_%s", strings.ToUpper(module), strings.ToUpper(name))
}

// GetDSNFromSecret returns the DSN for a database from the relevant secret
func GetDSNFromSecret(module, name string, secrets map[string][]byte) (string, error) {
	key := DSNSecretKey(module, name)
	dsn, ok := secrets[key]
	if !ok {
		return "", fmt.Errorf("secrets map %v is missing DSN with key %q", secrets, key)
	}
	dsnStr := string(dsn)
	return dsnStr[1 : len(dsnStr)-1], nil // chop leading + trailing quotes
}
