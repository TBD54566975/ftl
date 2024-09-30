package buildengine

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/alecthomas/types/either"
	"google.golang.org/protobuf/proto"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/errors"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
)

// Build a module in the given directory given the schema and module config.
//
// A lock file is used to ensure that only one build is running at a time.
func build(ctx context.Context, plugin LanguagePlugin, projectRootDir string, sch *schema.Schema, c moduleconfig.ModuleConfig, buildEnv []string, devMode bool) (*schema.Module, error) {
	config := c.Abs()

	logger := log.FromContext(ctx).Module(config.Module).Scope("build")
	ctx = log.ContextWithLogger(ctx, logger)

	if err := prepareBuild(ctx, config); err != nil {
		return nil, err
	}

	result, err := plugin.Build(ctx, projectRootDir, config, sch, buildEnv, devMode)
	if err != nil {
		return handleBuildResult(ctx, config, either.RightOf[BuildResult](err))
	}
	return handleBuildResult(ctx, config, either.LeftOf[error](result))
}

// TODO: docs
func prepareBuild(ctx context.Context, config moduleconfig.AbsModuleConfig) error {
	// clear the deploy directory before extracting schema
	if err := os.RemoveAll(config.DeployDir); err != nil {
		return fmt.Errorf("failed to clear deploy directory: %w", err)
	}
	if err := os.MkdirAll(config.DeployDir, 0700); err != nil {
		return fmt.Errorf("could not create deploy directory: %w", err)
	}

	log.FromContext(ctx).Infof("Building module")
	return nil
}

// TODO: docs
func handleBuildResult(ctx context.Context, config moduleconfig.AbsModuleConfig, eitherResult either.Either[BuildResult, error]) (*schema.Module, error) {
	logger := log.FromContext(ctx)

	var result BuildResult
	switch eitherResult := eitherResult.(type) {
	case either.Right[BuildResult, error]:
		return nil, fmt.Errorf("failed to build module: %w", eitherResult.Get())
	case either.Left[BuildResult, error]:
		result = eitherResult.Get()
	}

	var errs []error
	for _, e := range result.Errors {
		if e.Level == schema.WARN {
			logger.Log(log.Entry{Level: log.Warn, Message: e.Error(), Error: e})
			continue
		}
		errs = append(errs, e)
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	logger.Infof("Module built (%.2fs)", time.Since(result.StartTime).Seconds())

	// write schema proto to deploy directory
	schemaBytes, err := proto.Marshal(result.Schema.ToProto())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}
	if err := os.WriteFile(config.Schema(), schemaBytes, 0600); err != nil {
		return nil, fmt.Errorf("failed to write schema: %w", err)
	}
	return result.Schema, nil
}
