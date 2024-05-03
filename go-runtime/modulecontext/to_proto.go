package modulecontext

import (
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
)

// ToProto converts a ModuleContext to a proto response.
func (m *ModuleContext) ToProto(moduleName string) (*ftlv1.ModuleContextResponse, error) {
	databases := make([]*ftlv1.ModuleContextResponse_DSN, 0, len(m.databases))
	for name, entry := range m.databases {
		databases = append(databases, &ftlv1.ModuleContextResponse_DSN{
			Name: name,
			Type: ftlv1.ModuleContextResponse_DBType(entry.dbType),
			Dsn:  entry.dsn,
		})
	}
	return &ftlv1.ModuleContextResponse{
		Module:    moduleName,
		Configs:   m.configs,
		Secrets:   m.secrets,
		Databases: databases,
	}, nil
}
