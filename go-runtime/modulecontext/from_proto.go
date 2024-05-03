package modulecontext

import (
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
)

func FromProto(response *ftlv1.ModuleContextResponse) (*ModuleContext, error) {
	moduleCtx := New(response.Module)
	for name, data := range response.Configs {
		moduleCtx.configs[name] = data
	}
	for name, data := range response.Secrets {
		moduleCtx.secrets[name] = data
	}
	for _, entry := range response.Databases {
		if err := moduleCtx.AddDatabase(entry.Name, DBType(entry.Type), entry.Dsn); err != nil {
			return nil, err
		}
	}
	return moduleCtx, nil
}
