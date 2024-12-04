package deploymentcontext

import (
	"fmt"
	"strconv"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/deployment/v1"
)

// ToProto converts a DeploymentContext to a proto response.
func (m DeploymentContext) ToProto() *ftlv1.GetDeploymentContextResponse {
	databases := make([]*ftlv1.GetDeploymentContextResponse_DSN, 0, len(m.databases))
	for name, entry := range m.databases {
		databases = append(databases, &ftlv1.GetDeploymentContextResponse_DSN{
			Name: name,
			Type: entry.DBType.ToProto(),
			Dsn:  entry.DSN,
		})
	}
	return &ftlv1.GetDeploymentContextResponse{
		Module:    m.module,
		Configs:   m.configs,
		Secrets:   m.secrets,
		Databases: databases,
	}
}

func (x DBType) ToProto() ftlv1.GetDeploymentContextResponse_DbType {
	switch x {
	case DBTypeUnspecified:
		return ftlv1.GetDeploymentContextResponse_DB_TYPE_UNSPECIFIED
	case DBTypePostgres:
		return ftlv1.GetDeploymentContextResponse_DB_TYPE_POSTGRES
	case DBTypeMySQL:
		return ftlv1.GetDeploymentContextResponse_DB_TYPE_MYSQL
	default:
		panic(fmt.Sprintf("unknown DB type: %s", strconv.Itoa(int(x))))
	}
}
