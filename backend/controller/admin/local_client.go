package admin

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/schema"
	cf "github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/common/projectconfig"
	"github.com/TBD54566975/ftl/internal/buildengine"
)

// localClient reads and writes to local projectconfig files without making any network
// calls. It allows us to interface with local ftl-project.toml files without needing to
// start a controller.
type localClient struct {
	*AdminService
}

type diskSchemaRetriever struct {
	// Omit to use the project root as the deploy root.
	deployRoot optional.Option[string]
}

// NewLocalClient creates a admin client that reads and writes from the provided config and secret managers
func NewLocalClient(cm *cf.Manager[cf.Configuration], sm *cf.Manager[cf.Secrets]) Client {
	return &localClient{NewAdminService(cm, sm, &diskSchemaRetriever{})}
}

func (s *diskSchemaRetriever) GetActiveSchema(ctx context.Context) (*schema.Schema, error) {
	path, ok := projectconfig.DefaultConfigPath().Get()
	if !ok {
		return nil, fmt.Errorf("no project config path available")
	}
	projConfig, err := projectconfig.Load(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("could not load project config: %w", err)
	}
	modules, err := buildengine.DiscoverModules(ctx, projConfig.AbsModuleDirs())
	if err != nil {
		return nil, fmt.Errorf("could not discover modules: %w", err)
	}

	sch := &schema.Schema{}
	for _, m := range modules {
		schemaPath := m.Config.Abs().Schema
		if r, ok := s.deployRoot.Get(); ok {
			schemaPath = filepath.Join(r, m.Config.Module, m.Config.DeployDir, m.Config.Schema)
		}

		module, err := schema.ModuleFromProtoFile(schemaPath)
		if err != nil {
			return nil, fmt.Errorf("could not load module schema: %w", err)
		}
		sch.Upsert(module)
	}
	return sch, nil
}
