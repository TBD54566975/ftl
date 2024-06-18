package buildengine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/common/moduleconfig"
	"github.com/TBD54566975/ftl/go-runtime/compile"
	"github.com/TBD54566975/ftl/internal/errors"
	"github.com/TBD54566975/ftl/internal/flock"
	"github.com/TBD54566975/ftl/internal/log"
)

const BuildLockTimeout = time.Minute

// Build a module in the given directory given the schema and module config.
//
// A lock file is used to ensure that only one build is running at a time.
func Build(ctx context.Context, projectRootDir string, sch *schema.Schema, module Module, filesTransaction ModifyFilesTransaction) error {
	return buildModule(ctx, projectRootDir, sch, module, filesTransaction)
}

func buildModule(ctx context.Context, projectRootDir string, sch *schema.Schema, module Module, filesTransaction ModifyFilesTransaction) error {
	release, err := flock.Acquire(ctx, filepath.Join(module.Config.Dir, ".ftl.lock"), BuildLockTimeout)
	if err != nil {
		return err
	}
	defer release() //nolint:errcheck

	logger := log.FromContext(ctx).Scope(module.Config.Module)
	ctx = log.ContextWithLogger(ctx, logger)

	// clear the deploy directory before extracting schema
	if err := os.RemoveAll(module.Config.AbsDeployDir()); err != nil {
		return fmt.Errorf("failed to clear errors: %w", err)
	}

	startTime := time.Now()

	logger.Infof("Building module")
	switch module.Config.Language {
	case "go":
		err = buildGoModule(ctx, projectRootDir, sch, module, filesTransaction)
	case "kotlin":
		err = buildKotlinModule(ctx, sch, module)
	default:
		return fmt.Errorf("unknown language %q", module.Config.Language)
	}

	var errs []error
	if err != nil {
		errs = append(errs, err)
	}
	// read runtime-specific build errors from the build directory
	errorList, err := loadProtoErrors(module.Config)
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

	logger.Infof("Module built in (%.2fs)", time.Since(startTime).Seconds())
	return generateStubs(ctx, projectRootDir, sch, module, filesTransaction)
}

func loadProtoErrors(config moduleconfig.ModuleConfig) (*schema.ErrorList, error) {
	f := filepath.Join(config.AbsDeployDir(), config.Errors)
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

func generateStubs(ctx context.Context, projectRootDir string, sch *schema.Schema, module Module, filesTransaction ModifyFilesTransaction) error {
	logger := log.FromContext(ctx)

	if module.Config.Language == "go" {
		module, err := schema.ModuleFromProtoFile(module.Config.AbsSchemaDir())
		if err != nil {
			return fmt.Errorf("failed to load module from proto file: %w", err)
		}

		logger.Debugf("Generating stubs")
		return compile.GenerateStubsForModule(ctx, projectRootDir, module, filesTransaction)
	}

	return nil
}
