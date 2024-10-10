package buildengine

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/alecthomas/types/either"
	"google.golang.org/protobuf/proto"

	"github.com/TBD54566975/ftl/internal/buildengine/languageplugin"
	"github.com/TBD54566975/ftl/internal/builderrors"
	"github.com/TBD54566975/ftl/internal/errors"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/TBD54566975/ftl/internal/schema"
)

var invalidateDependenciesError = errors.New("dependencies need to be updated")

// Build a module in the given directory given the schema and module config.
//
// A lock file is used to ensure that only one build is running at a time.
//
// Returns invalidateDependenciesError if the build failed due to a change in dependencies.
func build(ctx context.Context, plugin languageplugin.LanguagePlugin, projectRootDir string, bctx languageplugin.BuildContext, buildEnv []string, devMode bool) (moduleSchema *schema.Module, deploy []string, err error) {
	logger := log.FromContext(ctx).Module(bctx.Config.Module).Scope("build")
	ctx = log.ContextWithLogger(ctx, logger)

	logger.Infof("Building module")

	result, err := plugin.Build(ctx, projectRootDir, bctx, buildEnv, devMode)
	if err != nil {
		return handleBuildResult(ctx, bctx.Config, either.RightOf[languageplugin.BuildResult](err))
	}
	return handleBuildResult(ctx, bctx.Config, either.LeftOf[error](result))
}

// handleBuildResult processes the result of a build
func handleBuildResult(ctx context.Context, c moduleconfig.ModuleConfig, eitherResult either.Either[languageplugin.BuildResult, error]) (moduleSchema *schema.Module, deploy []string, err error) {
	logger := log.FromContext(ctx)
	config := c.Abs()

	var result languageplugin.BuildResult
	switch eitherResult := eitherResult.(type) {
	case either.Right[languageplugin.BuildResult, error]:
		return nil, nil, fmt.Errorf("failed to build module: %w", eitherResult.Get())
	case either.Left[languageplugin.BuildResult, error]:
		result = eitherResult.Get()
	}

	if result.InvalidateDependencies {
		return nil, nil, invalidateDependenciesError
	}

	var errs []error
	for _, e := range result.Errors {
		if e.Level == builderrors.WARN {
			logger.Log(log.Entry{Level: log.Warn, Message: e.Error(), Error: e})
			continue
		}
		errs = append(errs, e)
	}

	if len(errs) > 0 {
		return nil, nil, errors.Join(errs...)
	}

	logger.Infof("Module built (%.2fs)", time.Since(result.StartTime).Seconds())

	// write schema proto to deploy directory
	schemaBytes, err := proto.Marshal(result.Schema.ToProto())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal schema: %w", err)
	}
	if err := os.WriteFile(config.Schema(), schemaBytes, 0600); err != nil {
		return nil, nil, fmt.Errorf("failed to write schema: %w", err)
	}
	return result.Schema, result.Deploy, nil
}
