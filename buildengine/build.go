package buildengine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/common/moduleconfig"
	"github.com/TBD54566975/ftl/internal/errors"
	"github.com/TBD54566975/ftl/internal/flock"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/slices"
)

const BuildLockTimeout = time.Minute

// Build a project in the given directory given the schema and project config.
//
// For a module, this will build the module. For an external library, this will build stubs for imported modules.
//
// A lock file is used to ensure that only one build is running at a time.
func Build(ctx context.Context, sch *schema.Schema, project Project, filesTransaction ModifyFilesTransaction) error {
	switch project := project.(type) {
	case Module:
		return buildModule(ctx, sch, project, filesTransaction)
	case ExternalLibrary:
		return buildExternalLibrary(ctx, sch, project)
	default:
		panic(fmt.Sprintf("unsupported project type: %T", project))
	}
}

func buildModule(ctx context.Context, sch *schema.Schema, module Module, filesTransaction ModifyFilesTransaction) error {
	release, err := flock.Acquire(ctx, filepath.Join(module.Dir, ".ftl-build-lock"), BuildLockTimeout)
	if err != nil {
		return err
	}
	defer release() //nolint:errcheck
	logger := log.FromContext(ctx).Scope(module.Module)
	ctx = log.ContextWithLogger(ctx, logger)

	// clear the deploy directory before extracting schema
	if err := os.RemoveAll(module.AbsDeployDir()); err != nil {
		return fmt.Errorf("failed to clear errors: %w", err)
	}

	logger.Infof("Building module")
	switch module.Language {
	case "go":
		err = buildGoModule(ctx, sch, module, filesTransaction)
	case "kotlin":
		err = buildKotlinModule(ctx, sch, module)
	default:
		return fmt.Errorf("unknown language %q", module.Language)
	}

	var errs []error
	if err != nil {
		errs = append(errs, err)
	}
	// read runtime-specific build errors from the build directory
	errorList, err := loadProtoErrors(module.ModuleConfig)
	if err != nil {
		return fmt.Errorf("failed to read build errors for module: %w", err)
	}
	schema.SortErrorsByPosition(errorList.Errors)
	for _, e := range errorList.Errors {
		errs = append(errs, e)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
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

func loadProtoErrors(module moduleconfig.ModuleConfig) (*schema.ErrorList, error) {
	f := filepath.Join(module.AbsDeployDir(), module.Errors)
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
