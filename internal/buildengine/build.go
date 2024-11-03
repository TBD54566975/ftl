package buildengine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/alecthomas/types/result"
	"google.golang.org/protobuf/proto"

	"github.com/TBD54566975/ftl/internal/buildengine/languageplugin"
	"github.com/TBD54566975/ftl/internal/builderrors"
	"github.com/TBD54566975/ftl/internal/errors"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/TBD54566975/ftl/internal/projectconfig"
	"github.com/TBD54566975/ftl/internal/schema"
)

var errInvalidateDependencies = errors.New("dependencies need to be updated")

// Build a module in the given directory given the schema and module config.
//
// Plugins must use a lock file to ensure that only one build is running at a time.
//
// Returns invalidateDependenciesError if the build failed due to a change in dependencies.
func build(ctx context.Context, plugin *languageplugin.LanguagePlugin, projectConfig projectconfig.Config, bctx languageplugin.BuildContext, devMode bool) (moduleSchema *schema.Module, deploy []string, err error) {
	logger := log.FromContext(ctx).Module(bctx.Config.Module).Scope("build")
	ctx = log.ContextWithLogger(ctx, logger)

	stubsRoot := stubsLanguageDir(projectConfig.Root(), bctx.Config.Language)
	return handleBuildResult(ctx, projectConfig, bctx.Config, result.From(plugin.Build(ctx, projectConfig.Root(), stubsRoot, bctx, devMode)))
}

// handleBuildResult processes the result of a build
func handleBuildResult(ctx context.Context, projectConfig projectconfig.Config, c moduleconfig.ModuleConfig, eitherResult result.Result[languageplugin.BuildResult]) (moduleSchema *schema.Module, deploy []string, err error) {
	logger := log.FromContext(ctx)
	config := c.Abs()

	result, err := eitherResult.Result()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build module: %w", err)
	}

	if result.InvalidateDependencies {
		return nil, nil, errInvalidateDependencies
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
	sch := result.Schema
	// TODO: decide if image is passed along specially or if plugin should just return runtime info optionally and we default things here
	sch.Runtime = &schema.ModuleRuntime{
		CreateTime:  time.Now(),
		Language:    c.Language,
		MinReplicas: 1,
		Image:       result.Image,
	}
	schemaBytes, err := proto.Marshal(result.Schema.ToProto())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal schema: %w", err)
	}
	schemaPath := projectConfig.SchemaPath(config.Module)
	err = os.MkdirAll(filepath.Dir(schemaPath), 0700)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create schema directory: %w", err)
	}
	if err := os.WriteFile(schemaPath, schemaBytes, 0600); err != nil {
		return nil, nil, fmt.Errorf("failed to write schema: %w", err)
	}
	return result.Schema, result.Deploy, nil
}
