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
	"github.com/TBD54566975/ftl/internal/errors"
	"github.com/TBD54566975/ftl/internal/flock"
	"github.com/TBD54566975/ftl/internal/log"
)

const BuildLockTimeout = time.Minute

// Build a module in the given directory given the schema and module config.
//
// A lock file is used to ensure that only one build is running at a time.
func Build(ctx context.Context, sch *schema.Schema, module Module, filesTransaction ModifyFilesTransaction) error {
	return buildModule(ctx, sch, module, filesTransaction)
}

func buildModule(ctx context.Context, sch *schema.Schema, module Module, filesTransaction ModifyFilesTransaction) error {
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
	switch module.Config.Language {
	case "go":
		err = buildGoModule(ctx, sch, module, filesTransaction)
	case "kotlin":
		err = buildKotlinModule(ctx, sch, module)
	case "rust":
		panic("unimplemented")
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
		errs = append(errs, e)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

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
