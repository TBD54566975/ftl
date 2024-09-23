package buildengine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/protobuf/proto"

	languagepb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/language"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/builderrors"
	"github.com/TBD54566975/ftl/internal/errors"
	"github.com/TBD54566975/ftl/internal/flock"
	"github.com/TBD54566975/ftl/internal/languageplugin"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
)

const BuildLockTimeout = time.Minute

// Build a module in the given directory given the schema and module config.
//
// A lock file is used to ensure that only one build is running at a time.
func build(ctx context.Context, plugin *languageplugin.LanguagePlugin, projectRootDir string, sch *schema.Schema, module Module, buildEnv []string, devMode bool) (*schema.Module, error) {
	release, err := flock.Acquire(ctx, filepath.Join(module.Config.Dir, ".ftl.lock"), BuildLockTimeout)
	if err != nil {
		return nil, err
	}
	defer release() //nolint:errcheck
	logger := log.FromContext(ctx).Scope(module.Config.Module)
	ctx = log.ContextWithLogger(ctx, logger)

	// clear the deploy directory before extracting schema
	if err := os.RemoveAll(module.Config.Abs().DeployDir); err != nil {
		return nil, fmt.Errorf("failed to clear errors: %w", err)
	}

	logger.Infof("Building module")

	startTime := time.Now()

	// switch module.Config.Language {
	// case "go":
	// 	err = buildGoModule(ctx, projectRootDir, sch, module, filesTransaction, buildEnv, devMode)
	// case "java", "kotlin":
	// 	err = buildJavaModule(ctx, module)
	// case "rust":
	// 	err = buildRustModule(ctx, sch, module)
	// default:
	// 	return fmt.Errorf("unknown language %q", module.Config.Language)

	result, err := plugin.Build(ctx, sch, projectRootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to build module: %w", err)
	}

	var errs []error
	// read runtime-specific build errors from the build directory
	// errorList, err := loadProtoErrors(module.Config.Abs())
	// if err != nil {
	// 	return fmt.Errorf("failed to read build errors for module: %w", err)
	// }

	for _, e := range result.Errors {
		if e.Level == builderrors.WARN {
			logger.Log(log.Entry{Level: log.Warn, Message: e.Error(), Error: e})
			continue
		}
		errs = append(errs, e)
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	logger.Infof("Module built (%.2fs)", time.Since(startTime).Seconds())

	// write schema proto to deploy directory
	schemaBytes, err := proto.Marshal(result.Schema.ToProto())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}
	if err := os.WriteFile(module.Config.Abs().Schema(), schemaBytes, 0600); err != nil {
		return nil, fmt.Errorf("failed to write schema: %w", err)
	}
	return result.Schema, nil
}

func loadProtoErrors(config moduleconfig.AbsModuleConfig) ([]*builderrors.Error, error) {
	if _, err := os.Stat(config.Errors); errors.Is(err, os.ErrNotExist) {
		return make([]*builderrors.Error, 0), nil
	}
	content, err := os.ReadFile(config.Errors)
	if err != nil {
		return nil, err
	}
	errorspb := &languagepb.ErrorList{}
	err = proto.Unmarshal(content, errorspb)
	if err != nil {
		return nil, err
	}
	return languagepb.ErrorsFromProto(errorspb), nil
}
