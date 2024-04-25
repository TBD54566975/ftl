package modulecontext

import (
	"context"
	"fmt"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
)

// ToProto converts a ModuleContext to a proto response.
func (m *ModuleContext) ToProto(ctx context.Context) (*ftlv1.ModuleContextResponse, error) {
	config, err := m.configManager.MapForModule(ctx, m.module)
	if err != nil {
		return nil, fmt.Errorf("failed to get config map: %w", err)
	}
	secrets, err := m.secretsManager.MapForModule(ctx, m.module)
	if err != nil {
		return nil, fmt.Errorf("failed to get secrets map: %w", err)
	}
	databases := make([]*ftlv1.ModuleContextResponse_DSN, 0, len(m.dbProvider.entries))
	for name, entry := range m.dbProvider.entries {
		databases = append(databases, &ftlv1.ModuleContextResponse_DSN{
			Name: name,
			Type: ftlv1.ModuleContextResponse_DBType(entry.dbType),
			Dsn:  entry.dsn,
		})
	}
	return &ftlv1.ModuleContextResponse{
		Module:    m.module,
		Configs:   config,
		Secrets:   secrets,
		Databases: databases,
	}, nil
}
