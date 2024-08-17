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
	"github.com/TBD54566975/ftl/internal/errors"
	"github.com/TBD54566975/ftl/internal/flock"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
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
	if err := os.RemoveAll(module.Config.Abs().DeployDir); err != nil {
		return fmt.Errorf("failed to clear errors: %w", err)
	}

	logger.Infof("Building module")

	startTime := time.Now()

	switch module.Config.Language {
	case "go":
		err = buildGoModule(ctx, projectRootDir, sch, module, filesTransaction)
	case "java", "kotlin":
		err = buildJavaModule(ctx, module)
	case "rust":
		err = buildRustModule(ctx, sch, module)
	default:
		return fmt.Errorf("unknown language %q", module.Config.Language)
	}

	var errs []error
	if err != nil {
		errs = append(errs, err)
	}
	// read runtime-specific build errors from the build directory
	errorList, err := loadProtoErrors(module.Config.Abs())
	if err != nil {
		return fmt.Errorf("failed to read build errors for module: %w", err)
	}
	schema.SortErrorsByPosition(errorList.Errors)
	for _, e := range errorList.Errors {
		if e.Level == schema.WARN {
			logger.Log(log.Entry{Level: log.Warn, Message: e.Error(), Error: e})
			continue
		}
		errs = append(errs, e)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	logger.Infof("Module built (%.2fs)", time.Since(startTime).Seconds())

	return nil
}

func loadProtoErrors(config moduleconfig.AbsModuleConfig) (*schema.ErrorList, error) {
	if _, err := os.Stat(config.Errors); errors.Is(err, os.ErrNotExist) {
		return &schema.ErrorList{Errors: make([]*schema.Error, 0)}, nil
	}

	content, err := os.ReadFile(config.Errors)
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
