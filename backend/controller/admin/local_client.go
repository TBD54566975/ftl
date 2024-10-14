package admin

import (
	"context"
	"fmt"

	"github.com/alecthomas/types/either"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/internal/bind"
	"github.com/TBD54566975/ftl/internal/buildengine/languageplugin"
	cf "github.com/TBD54566975/ftl/internal/configuration"
	"github.com/TBD54566975/ftl/internal/configuration/manager"
	"github.com/TBD54566975/ftl/internal/errors"
	"github.com/TBD54566975/ftl/internal/projectconfig"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/watch"
)

// localClient reads and writes to local projectconfig files without making any network
// calls. It allows us to interface with local ftl-project.toml files without needing to
// start a controller.
type localClient struct {
	*AdminService
}

type diskSchemaRetriever struct {
	// Omit to use the project root as the deploy root (used in tests)
	deployRoot optional.Option[string]
}

// NewLocalClient creates a admin client that reads and writes from the provided config and secret managers
func NewLocalClient(cm *manager.Manager[cf.Configuration], sm *manager.Manager[cf.Secrets], bindAllocator *bind.BindAllocator) Client {
	return &localClient{NewAdminService(cm, sm, &diskSchemaRetriever{}, optional.Some(bindAllocator))}
}

func (s *diskSchemaRetriever) GetActiveSchema(ctx context.Context, bAllocator optional.Option[*bind.BindAllocator]) (*schema.Schema, error) {
	bindAllocator, ok := bAllocator.Get()
	if !ok {
		return nil, fmt.Errorf("no bind allocator available")
	}
	path, ok := projectconfig.DefaultConfigPath().Get()
	if !ok {
		return nil, fmt.Errorf("no project config path available")
	}
	projConfig, err := projectconfig.Load(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("could not load project config: %w", err)
	}
	modules, err := watch.DiscoverModules(ctx, projConfig.AbsModuleDirs())
	if err != nil {
		return nil, fmt.Errorf("could not discover modules: %w", err)
	}

	moduleSchemas := make(chan either.Either[*schema.Module, error], 32)
	defer close(moduleSchemas)

	for _, m := range modules {
		go func() {
			// Loading a plugin can be expensive. Is there a better way?
			plugin, err := languageplugin.New(ctx, bindAllocator, m.Language)
			if err != nil {
				moduleSchemas <- either.RightOf[*schema.Module](fmt.Errorf("could not load plugin for %s: %w", m.Module, err))
			}
			defer plugin.Kill() // nolint:errcheck

			customDefaults, err := plugin.ModuleConfigDefaults(ctx, m.Dir)
			if err != nil {
				moduleSchemas <- either.RightOf[*schema.Module](fmt.Errorf("could not get module config defaults for %s: %w", m.Module, err))
			}

			config, err := m.FillDefaultsAndValidate(customDefaults)
			if err != nil {
				moduleSchemas <- either.RightOf[*schema.Module](fmt.Errorf("could not validate module config for %s: %w", m.Module, err))
			}
			module, err := schema.ModuleFromProtoFile(config.Abs().Schema())
			if err != nil {
				moduleSchemas <- either.RightOf[*schema.Module](fmt.Errorf("could not load module schema: %w", err))
				return
			}
			moduleSchemas <- either.LeftOf[error](module)
		}()
	}
	sch := &schema.Schema{}
	errs := []error{}
	for range len(modules) {
		result := <-moduleSchemas
		switch result := result.(type) {
		case either.Left[*schema.Module, error]:
			sch.Upsert(result.Get())
		case either.Right[*schema.Module, error]:
			errs = append(errs, result.Get())
		default:
			panic(fmt.Sprintf("unexpected type %T", result))
		}
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return sch, nil
}
