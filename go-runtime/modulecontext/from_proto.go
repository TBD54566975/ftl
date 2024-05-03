package modulecontext

import (
	"fmt"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
)

func FromProto(response *ftlv1.ModuleContextResponse) (ModuleContext, error) {
	databases := map[string]Database{}
	for name, entry := range response.Databases {
		db, err := NewDatabase(DBType(entry.Type), entry.Dsn)
		if err != nil {
			return ModuleContext{}, fmt.Errorf("could not create database %q with DSN %q: %w", name, entry.Dsn, err)
		}
		databases[entry.Name] = db
	}
	return New(response.Module).Update(response.Configs, response.Secrets, databases), nil
}
