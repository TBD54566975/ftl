package deploymentcontext

import (
	"fmt"
	"strconv"

	deploymentpb "github.com/block/ftl/backend/protos/xyz/block/ftl/deployment/v1"
)

// ToProto converts a DeploymentContext to a proto response.
func (m DeploymentContext) ToProto() *deploymentpb.GetDeploymentContextResponse {
	databases := make([]*deploymentpb.GetDeploymentContextResponse_DSN, 0, len(m.databases))
	for name, entry := range m.databases {
		databases = append(databases, &deploymentpb.GetDeploymentContextResponse_DSN{
			Name: name,
			Type: entry.DBType.ToProto(),
			Dsn:  entry.DSN,
		})
	}
	routes := make([]*deploymentpb.GetDeploymentContextResponse_Route, 0, len(m.routes))
	for dep, entry := range m.routes {
		routes = append(routes, &deploymentpb.GetDeploymentContextResponse_Route{
			Deployment: dep,
			Uri:        entry,
		})
	}
	return &deploymentpb.GetDeploymentContextResponse{
		Module:    m.module,
		Configs:   m.configs,
		Secrets:   m.secrets,
		Databases: databases,
		Routes:    routes,
	}
}

func (x DBType) ToProto() deploymentpb.GetDeploymentContextResponse_DbType {
	switch x {
	case DBTypeUnspecified:
		return deploymentpb.GetDeploymentContextResponse_DB_TYPE_UNSPECIFIED
	case DBTypePostgres:
		return deploymentpb.GetDeploymentContextResponse_DB_TYPE_POSTGRES
	case DBTypeMySQL:
		return deploymentpb.GetDeploymentContextResponse_DB_TYPE_MYSQL
	default:
		panic(fmt.Sprintf("unknown DB type: %s", strconv.Itoa(int(x))))
	}
}
