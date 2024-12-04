package deploymentcontext

import (
	"fmt"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/deployment/v1"
)

func DBTypeFromProto(x ftlv1.GetDeploymentContextResponse_DbType) DBType {
	switch x {
	case ftlv1.GetDeploymentContextResponse_DB_TYPE_UNSPECIFIED:
		return DBTypeUnspecified
	case ftlv1.GetDeploymentContextResponse_DB_TYPE_POSTGRES:
		return DBTypePostgres
	case ftlv1.GetDeploymentContextResponse_DB_TYPE_MYSQL:
		return DBTypeMySQL
	default:
		panic(fmt.Sprintf("unknown DB type: %d", x))
	}
}

func FromProto(response *ftlv1.GetDeploymentContextResponse) (DeploymentContext, error) {
	databases := map[string]Database{}
	for name, entry := range response.Databases {
		db, err := NewDatabase(DBTypeFromProto(entry.Type), entry.Dsn)
		if err != nil {
			return DeploymentContext{}, fmt.Errorf("could not create database %q with DSN %q: %w", name, entry.Dsn, err)
		}
		databases[entry.Name] = db
	}
	return NewBuilder(response.Module).AddConfigs(response.Configs).AddSecrets(response.Secrets).AddDatabases(databases).Build(), nil
}
