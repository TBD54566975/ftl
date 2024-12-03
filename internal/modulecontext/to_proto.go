package modulecontext

import (
	"fmt"
	"strconv"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
)

// ToProto converts a ModuleContext to a proto response.
func (m ModuleContext) ToProto() *ftlv1.GetModuleContextResponse {
	databases := make([]*ftlv1.GetModuleContextResponse_DSN, 0, len(m.databases))
	for name, entry := range m.databases {
		databases = append(databases, &ftlv1.GetModuleContextResponse_DSN{
			Name: name,
			Type: entry.DBType.ToProto(),
			Dsn:  entry.DSN,
		})
	}
	routes := make([]*ftlv1.GetModuleContextResponse_Route, 0, len(m.routes))
	for name, entry := range m.routes {
		routes = append(routes, &ftlv1.GetModuleContextResponse_Route{
			Module: name,
			Uri:    entry,
		})
	}
	return &ftlv1.GetModuleContextResponse{
		Module:    m.module,
		Configs:   m.configs,
		Secrets:   m.secrets,
		Routes:    routes,
		Databases: databases,
	}
}

func (x DBType) ToProto() ftlv1.GetModuleContextResponse_DbType {
	switch x {
	case DBTypeUnspecified:
		return ftlv1.GetModuleContextResponse_DB_TYPE_UNSPECIFIED
	case DBTypePostgres:
		return ftlv1.GetModuleContextResponse_DB_TYPE_POSTGRES
	case DBTypeMySQL:
		return ftlv1.GetModuleContextResponse_DB_TYPE_MYSQL
	default:
		panic(fmt.Sprintf("unknown DB type: %s", strconv.Itoa(int(x))))
	}
}
