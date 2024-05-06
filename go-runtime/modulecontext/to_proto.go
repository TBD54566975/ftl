package modulecontext

import (
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
)

// ToProto converts a ModuleContext to a proto response.
func (m ModuleContext) ToProto() *ftlv1.ModuleContextResponse {
	databases := make([]*ftlv1.ModuleContextResponse_DSN, 0, len(m.databases))
	for name, entry := range m.databases {
		databases = append(databases, &ftlv1.ModuleContextResponse_DSN{
			Name: name,
			Type: ftlv1.ModuleContextResponse_DBType(entry.DBType),
			Dsn:  entry.DSN,
		})
	}
	return &ftlv1.ModuleContextResponse{
		Module:    m.module,
		Configs:   m.configs,
		Secrets:   m.secrets,
		Databases: databases,
	}
}
