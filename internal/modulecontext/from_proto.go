package modulecontext

import (
	"fmt"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
)

func DBTypeFromProto(x ftlv1.GetModuleContextResponse_DbType) DBType {
	switch x {
	case ftlv1.GetModuleContextResponse_DB_TYPE_UNSPECIFIED:
		return DBTypeUnspecified
	case ftlv1.GetModuleContextResponse_DB_TYPE_POSTGRES:
		return DBTypePostgres
	case ftlv1.GetModuleContextResponse_DB_TYPE_MYSQL:
		return DBTypeMySQL
	default:
		panic(fmt.Sprintf("unknown DB type: %d", x))
	}
}

func FromProto(response *ftlv1.GetModuleContextResponse) (ModuleContext, error) {
	databases := map[string]Database{}
	for name, entry := range response.Databases {
		db, err := NewDatabase(DBTypeFromProto(entry.Type), entry.Dsn)
		if err != nil {
			return ModuleContext{}, fmt.Errorf("could not create database %q with DSN %q: %w", name, entry.Dsn, err)
		}
		databases[entry.Name] = db
	}
	return NewBuilder(response.Module).AddConfigs(response.Configs).AddSecrets(response.Secrets).AddDatabases(databases).Build(), nil
}
