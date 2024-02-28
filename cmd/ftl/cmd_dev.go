package main

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/sync/errgroup"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/buildengine"
	"github.com/TBD54566975/ftl/common/moduleconfig"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type moduleFolderInfo struct {
	moduleName   string
	schema       *schema.Module
	forceRebuild bool
}

type devCmd struct {
	BaseDir         string        `arg:"" help:"Directory to watch for FTL modules" type:"existingdir" default:"."`
	Watch           time.Duration `help:"Watch template directory at this frequency and regenerate on change." default:"500ms"`
	FailureDelay    time.Duration `help:"Delay before retrying a failed deploy." default:"500ms"`
	ReconnectDelay  time.Duration `help:"Delay before attempting to reconnect to FTL." default:"1s"`
	ExitAfterDeploy bool          `help:"Exit after all modules are deployed successfully." default:"false"`
	NoServe         bool          `help:"Do not start the FTL server." default:"false"`
	ServeCmd        serveCmd      `embed:"" prefix:"serve-"`
}

type moduleMap map[string]*moduleFolderInfo

func (m *moduleMap) ForceRebuild(dir string) {
	(*m)[dir].forceRebuild = true
}

func (m *moduleMap) AddModule(dir string, module string) {
	(*m)[dir] = &moduleFolderInfo{
		moduleName: module,
	}
}

func (m *moduleMap) RemoveModule(dir string) {
	delete(*m, dir)
}

func (m *moduleMap) SetModule(dir string, module *moduleFolderInfo) {
	(*m)[dir] = module
}

func (m *moduleMap) RebuildDependentModules(ctx context.Context, sch *schema.Module) {
	logger := log.FromContext(ctx)
	var changedModuleDir string
	for dir, moduleInfo := range *m {
		if moduleInfo.moduleName == sch.Name {
			changedModuleDir = dir
		}
	}

	// no module found, nothing to do
	if (*m)[changedModuleDir] == nil {
		return
	}

	oldSchema := (*m)[changedModuleDir].schema
	(*m)[changedModuleDir].schema = sch

	// no change in schema, nothing to do
	if oldSchema == nil || oldSchema.String() == sch.String() {
		return
	}

	for dir, moduleInfo := range *m {
		if moduleInfo.schema == nil {
			continue
		}

		for _, imp := range moduleInfo.schema.Imports() {
			if imp == sch.Name {
				logger.Warnf("Rebuilding %q due to %q schema changes", moduleInfo.moduleName, (*m)[changedModuleDir].moduleName)
				(*m).ForceRebuild(dir)
			}
		}
	}
}

func (d *devCmd) Run(ctx context.Context) error {
	logger := log.FromContext(ctx)
	client := rpc.ClientFromContext[ftlv1connect.ControllerServiceClient](ctx)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg, ctx := errgroup.WithContext(ctx)

	if !d.NoServe {
		wg.Go(func() error {
			return d.ServeCmd.Run(ctx)
		})
	}

	logger.Debugf("Watching %s for FTL modules", d.BaseDir)

	schemaChanges := make(chan *schema.Module, 64)
	modules := make(moduleMap)

	wg.Go(func() error {
		return d.watchForSchemaChanges(ctx, client, schemaChanges)
	})

	previousFailures := 0

	// Map of module directory to file hashes
	fileHashes := map[string]map[string][]byte{}

	for {
		logger.Tracef("Scanning %s for FTL module changes", d.BaseDir)
		delay := d.Watch

		tomls, err := d.getTomls()
		if err != nil {
			return err
		}

		err = d.addOrRemoveModules(tomls, modules)
		if err != nil {
			return err
		}

		allModulesDeployed := true

		failedModules := map[string]bool{}

		for dir := range modules {
			currentModule := modules[dir]
			config, err := moduleconfig.LoadModuleConfig(dir)
			if err != nil {
				return err
			}
			hashes, err := buildengine.ComputeFileHashes(config)
			if err != nil {
				return err
			}

			changeType, path, equal := buildengine.CompareFileHashes(fileHashes[dir], hashes)
			if currentModule.forceRebuild || !equal {
				if currentModule.forceRebuild {
					logger.Debugf("Forcing rebuild of module %s", dir)
					currentModule.forceRebuild = false
				} else {
					logger.Warnf("Detected change in %s%s, rebuilding...", changeType, path)
				}
				deploy := deployCmd{
					Replicas: 1,
					Dirs:     []string{dir},
				}
				err = deploy.Run(ctx)
				if err != nil {
					logger.Errorf(err, "Error deploying module %s. Will retry", dir)
					failedModules[dir] = true
					// Increase delay when there's a compile failure.
					delay = d.FailureDelay
					allModulesDeployed = false
				} else {
					modules.SetModule(dir, currentModule)
				}
			}
			fileHashes[dir] = hashes
		}
		if previousFailures != len(failedModules) || len(modules) == 0 {
			logger.Debugf("Detected %d failed modules, previously had %d", len(failedModules), previousFailures)
			for module := range failedModules {
				modules.ForceRebuild(module)
			}
			previousFailures = len(failedModules)
		}

		if allModulesDeployed && d.ExitAfterDeploy {
			logger.Infof("All modules deployed successfully.")
			cancel()
			return wg.Wait()
		}

		select {
		case module := <-schemaChanges:
			modules.RebuildDependentModules(ctx, module)

		drainLoop: // Drain all messages from the channel to avoid extra redeploys
			for {
				select {
				case module := <-schemaChanges:
					modules.RebuildDependentModules(ctx, module)
				default:
					break drainLoop
				}
			}
		case <-time.After(delay):
		case <-ctx.Done():
			return wg.Wait()
		}
	}
}

func (d *devCmd) watchForSchemaChanges(ctx context.Context, client ftlv1connect.ControllerServiceClient, schemaChanges chan *schema.Module) error {
	logger := log.FromContext(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		stream, err := client.PullSchema(ctx, connect.NewRequest(&ftlv1.PullSchemaRequest{}))
		if err != nil {
			return err
		}

		for stream.Receive() {
			select {
			case <-ctx.Done():
				logger.Warnf("Context canceled during schema change streaming, closing stream...")
				stream.Close()
				return ctx.Err()
			default:
			}

			msg := stream.Msg()
			if msg.ChangeType == ftlv1.DeploymentChangeType_DEPLOYMENT_ADDED || msg.ChangeType == ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGED {
				module, err := schema.ModuleFromProto(msg.Schema)
				if err != nil {
					return err
				}
				schemaChanges <- module
			}
		}

		if errors.Is(ctx.Err(), context.Canceled) {
			logger.Infof("Stream disconnected, attempting to reconnect...")
		}

		stream.Close()
		time.Sleep(d.ReconnectDelay)
	}
}

func (d *devCmd) getTomls() ([]string, error) {
	baseDir := d.BaseDir
	tomls := []string{}

	err := buildengine.WalkDir(baseDir, func(srcPath string, d fs.DirEntry) error {
		if filepath.Base(srcPath) == "ftl.toml" {
			tomls = append(tomls, srcPath)
			return buildengine.ErrSkip // Return errSkip to stop recursion in this branch
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return tomls, nil
}

func (d *devCmd) addOrRemoveModules(tomls []string, modules moduleMap) error {
	for _, toml := range tomls {
		dir := filepath.Dir(toml)
		if _, ok := modules[dir]; !ok {
			config, err := moduleconfig.LoadModuleConfig(dir)
			if err != nil {
				return err
			}
			modules.AddModule(dir, config.Module)
		}
	}

	for dir := range modules {
		found := false
		for _, toml := range tomls {
			if filepath.Dir(toml) == dir {
				found = true
				break
			}
		}
		if !found {
			modules.RemoveModule(dir)
		}
	}
	return nil
}
