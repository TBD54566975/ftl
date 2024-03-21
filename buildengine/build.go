package buildengine

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/slices"
)

// Build a module in the given directory given the schema and project config.
func Build(ctx context.Context, sch *schema.Schema, project Project) error {
	switch project := project.(type) {
	case Module:
		return buildModule(ctx, sch, project)
	case ExternalLibrary:
		return buildExternalLibrary(ctx, sch, project)
	default:
		return fmt.Errorf("unsupported project type: %T", project)
	}
}

func buildModule(ctx context.Context, sch *schema.Schema, module Module) error {
	logger := log.FromContext(ctx).Scope(module.Module)
	ctx = log.ContextWithLogger(ctx, logger)

	logger.Infof("Building module")
	switch module.Language {
	case "go":
		return buildGoModule(ctx, sch, module)
	case "kotlin":
		return buildKotlinModule(ctx, sch, module)
	default:
		return fmt.Errorf("unknown language %q", module.Language)
	}
}

func buildExternalLibrary(ctx context.Context, sch *schema.Schema, lib ExternalLibrary) error {
	logger := log.FromContext(ctx).Scope(filepath.Base(lib.Dir))
	ctx = log.ContextWithLogger(ctx, logger)

	imported := slices.Map(sch.Modules, func(m *schema.Module) string {
		return m.Name
	})
	logger.Infof("Generating stubs [%s] for %v", strings.Join(imported, ", "), lib)

	switch lib.Language {
	case "go":
		if err := buildGoLibrary(ctx, sch, lib); err != nil {
			return err
		}
	case "kotlin":
		if err := buildKotlinLibrary(ctx, sch, lib); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown language %q for %s", lib.Language, lib)
	}

	logger.Infof("Generated stubs [%s] for %v", strings.Join(imported, ", "), lib)
	return nil
}
