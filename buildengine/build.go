package buildengine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/errors"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/slices"
)

// Build a project in the given directory given the schema and project config.
// For a module, this will build the module. For an external library, this will build stubs for imported modules.
func Build(ctx context.Context, sch *schema.Schema, project Project) error {
	switch project := project.(type) {
	case Module:
		return buildModule(ctx, sch, project)
	case ExternalLibrary:
		return buildExternalLibrary(ctx, sch, project)
	default:
		panic(fmt.Sprintf("unsupported project type: %T", project))
	}
}

func buildModule(ctx context.Context, sch *schema.Schema, module Module) error {
	logger := log.FromContext(ctx).Scope(module.Module)
	ctx = log.ContextWithLogger(ctx, logger)

	logger.Infof("Building module")
	var err error
	switch module.Language {
	case "go":
		err = buildGoModule(ctx, sch, module)
	case "kotlin":
		err = buildKotlinModule(ctx, sch, module)
	default:
		return fmt.Errorf("unknown language %q", module.Language)
	}

	if err != nil {
		// read runtime-specific build errors from the build directory
		errorList, err := loadProtoErrors(module.AbsDeployDir())
		if err != nil {
			return fmt.Errorf("failed to read build errors for module: %w", err)
		}
		for _, e := range errorList.Errors {
			errs = append(errs, e)
		}
		errs = errors.DeduplicateErrors(errs)
		schema.SortErrorsByPosition(errs)
		return errors.Join(errs...)
	}

	return err
}

func buildExternalLibrary(ctx context.Context, sch *schema.Schema, lib ExternalLibrary) error {
	logger := log.FromContext(ctx).Scope(filepath.Base(lib.Dir))
	ctx = log.ContextWithLogger(ctx, logger)

	imported := slices.Map(sch.Modules, func(m *schema.Module) string {
		return m.Name
	})
	logger.Debugf("Generating stubs [%s] for %v", strings.Join(imported, ", "), lib)

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
		return fmt.Errorf("unknown language %q for library %q", lib.Language, lib.Config().Key)
	}

	logger.Infof("Generated stubs [%s] for %v", strings.Join(imported, ", "), lib)
	return nil
}

func loadProtoErrors(buildDir string) (*schema.ErrorList, error) {
	f := filepath.Join(buildDir, "errors.pb")
	if _, err := os.Stat(f); errors.Is(err, os.ErrNotExist) {
		return &schema.ErrorList{Errors: make([]*schema.Error, 0)}, nil
	}

	content, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}
	errorspb := &schemapb.ErrorList{}
	err = proto.Unmarshal(content, errorspb)
	if err != nil {
		return nil, err
	}
	return schema.ErrorListFromProto(errorspb), nil
}
