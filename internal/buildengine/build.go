package buildengine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/protobuf/proto"

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
func build(ctx context.Context, plugin LanguagePlugin, projectRootDir string, sch *schema.Schema, c moduleconfig.ModuleConfig, buildEnv []string, devMode bool) (*schema.Module, error) {
	config := c.Abs()
	release, err := flock.Acquire(ctx, filepath.Join(config.Dir, ".ftl.lock"), BuildLockTimeout)
	if err != nil {
		return nil, fmt.Errorf("could not acquire build lock for %v: %w", config.Module, err)
	}
	defer release() //nolint:errcheck
	logger := log.FromContext(ctx).Module(config.Module).Scope("build")
	ctx = log.ContextWithLogger(ctx, logger)

	// clear the deploy directory before extracting schema
	if err := os.RemoveAll(config.DeployDir); err != nil {
		return nil, fmt.Errorf("failed to clear errors: %w", err)
	}

	logger.Infof("Building module")

	startTime := time.Now()

	result, err := plugin.Build(ctx, projectRootDir, config, sch, buildEnv, devMode)
	if err != nil {
		return nil, fmt.Errorf("failed to build module: %w", err)
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

	logger.Infof("Module built (%.2fs)", time.Since(startTime).Seconds())

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
