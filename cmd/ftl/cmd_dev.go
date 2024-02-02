package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/bmatcuk/doublestar/v4"
	"golang.org/x/sync/errgroup"

	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/moduleconfig"
	"github.com/TBD54566975/ftl/backend/schema"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

type moduleFolderInfo struct {
	moduleName string
	fileHashes map[string][]byte
	schema     *schema.Module
}

type devCmd struct {
	BaseDir         string        `arg:"" help:"Directory to watch for FTL modules" type:"existingdir" default:"."`
	Watch           time.Duration `help:"Watch template directory at this frequency and regenerate on change." default:"500ms"`
	FailureDelay    time.Duration `help:"Delay before retrying a failed deploy." default:"5s"`
	ReconnectDelay  time.Duration `help:"Delay before attempting to reconnect to FTL." default:"1s"`
	ExitAfterDeploy bool          `help:"Exit after all modules are deployed successfully." default:"false"`
}

type moduleMap map[string]*moduleFolderInfo

func (m *moduleMap) ForceRebuild(dir string) {
	(*m)[dir].fileHashes = make(map[string][]byte)
}

func (m *moduleMap) AddModule(dir string, module string) {
	(*m)[dir] = &moduleFolderInfo{
		moduleName: module,
		fileHashes: make(map[string][]byte),
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

func (d *devCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	logger := log.FromContext(ctx)
	logger.Debugf("Watching %s for FTL modules", d.BaseDir)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	schemaChanges := make(chan *schema.Module, 64)
	modules := make(moduleMap)

	wg, ctx := errgroup.WithContext(ctx)

	wg.Go(func() error {
		return d.watchForSchemaChanges(ctx, client, schemaChanges)
	})

	for {
		logger.Tracef("Scanning %s for FTL module changes", d.BaseDir)
		delay := d.Watch

		tomls, err := d.getTomls(ctx)
		if err != nil {
			return err
		}

		err = d.addOrRemoveModules(tomls, modules)
		if err != nil {
			return err
		}

		allModulesDeployed := true

		for dir := range modules {
			currentModule := modules[dir]
			hashes, err := d.computeFileHashes(ctx, dir)
			if err != nil {
				return err
			}

			if !compareFileHashes(ctx, currentModule.fileHashes, hashes) {
				deploy := deployCmd{
					Replicas:  1,
					ModuleDir: dir,
				}
				err = deploy.Run(ctx, client)
				if err != nil {
					logger.Errorf(err, "Error deploying module %s. Will retry", dir)
					modules.RemoveModule(dir)
					// Increase delay when there's a compile failure.
					delay = d.FailureDelay
					allModulesDeployed = false
				} else {
					currentModule.fileHashes = hashes
					modules.SetModule(dir, currentModule)
				}
			}
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

func (d *devCmd) getTomls(ctx context.Context) ([]string, error) {
	baseDir := d.BaseDir
	ignores := initGitIgnore(ctx, baseDir)
	tomls := []string{}

	err := walkDir(baseDir, ignores, func(srcPath string, d fs.DirEntry) error {
		if filepath.Base(srcPath) == "ftl.toml" {
			tomls = append(tomls, srcPath)
			return errSkip // Return errSkip to stop recursion in this branch
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
			config, err := moduleconfig.LoadConfig(dir)
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

func (d *devCmd) computeFileHashes(ctx context.Context, dir string) (map[string][]byte, error) {
	config, err := moduleconfig.LoadConfig(dir)
	if err != nil {
		return nil, err
	}

	ignores := initGitIgnore(ctx, dir)

	fileHashes := make(map[string][]byte)
	err = walkDir(dir, ignores, func(srcPath string, entry fs.DirEntry) error {
		for _, pattern := range config.Watch {
			relativePath, err := filepath.Rel(dir, srcPath)
			if err != nil {
				return err
			}

			match, err := doublestar.PathMatch(pattern, relativePath)
			if err != nil {
				return err
			}

			if match && !entry.IsDir() {
				file, err := os.Open(srcPath)
				if err != nil {
					return err
				}

				hasher := sha256.New()
				if _, err := io.Copy(hasher, file); err != nil {
					_ = file.Close()
					return err
				}

				fileHashes[srcPath] = hasher.Sum(nil)

				if err := file.Close(); err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return fileHashes, err
}

// errSkip is returned by walkDir to skip a file or directory.
var errSkip = errors.New("skip directory")

// Depth-first walk of dir executing fn after each entry.
func walkDir(dir string, ignores []string, fn func(path string, d fs.DirEntry) error) error {
	dirInfo, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if err = fn(dir, fs.FileInfoToDirEntry(dirInfo)); err != nil {
		if errors.Is(err, errSkip) {
			return nil
		}
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var dirs []os.DirEntry

	// Process files first, then recurse into directories.
	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())

		// Check if the path matches any ignore pattern
		shouldIgnore := false
		for _, pattern := range ignores {
			match, err := doublestar.PathMatch(pattern, fullPath)
			if err != nil {
				return err
			}
			if match {
				shouldIgnore = true
				break
			}
		}

		if shouldIgnore {
			continue // Skip this entry
		}

		if entry.IsDir() {
			dirs = append(dirs, entry)
		} else {
			if err = fn(fullPath, entry); err != nil {
				if errors.Is(err, errSkip) {
					// If errSkip is found in a file, skip the remaining files in this directory
					return nil
				}
				return err
			}
		}
	}

	// Then, recurse into subdirectories
	for _, dirEntry := range dirs {
		dirPath := filepath.Join(dir, dirEntry.Name())
		ignores = append(ignores, loadGitIgnore(dirPath)...)
		if err := walkDir(dirPath, ignores, fn); err != nil {
			if errors.Is(err, errSkip) {
				return errSkip // Propagate errSkip upwards to stop this branch of recursion
			}
			return err
		}
	}
	return nil
}

func initGitIgnore(ctx context.Context, dir string) []string {
	ignore := []string{
		"**/.*",
		"**/.*/**",
	}
	home, err := os.UserHomeDir()
	if err == nil {
		ignore = append(ignore, loadGitIgnore(home)...)
	}
	gitRoot := gitRoot(ctx, dir)
	if gitRoot != "" {
		for current := dir; strings.HasPrefix(current, gitRoot); current = path.Dir(current) {
			ignore = append(ignore, loadGitIgnore(current)...)
		}
	}
	return ignore
}

func loadGitIgnore(dir string) []string {
	r, err := os.Open(path.Join(dir, ".gitignore"))
	if err != nil {
		return nil
	}
	ignore := []string{}
	lr := bufio.NewScanner(r)
	for lr.Scan() {
		line := lr.Text()
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' || line[0] == '!' { // We don't support negation.
			continue
		}
		if strings.HasSuffix(line, "/") {
			line = path.Join("**", line, "**/*")
		} else if !strings.ContainsRune(line, '/') {
			line = path.Join("**", line)
		}
		ignore = append(ignore, line)
	}
	return ignore
}

func compareFileHashes(ctx context.Context, oldFiles, newFiles map[string][]byte) bool {
	logger := log.FromContext(ctx)

	for key, hash1 := range oldFiles {
		hash2, exists := newFiles[key]
		if !exists {
			logger.Warnf("Detected file removed: %s", key)
			return false
		}
		if !bytes.Equal(hash1, hash2) {
			logger.Warnf("Detected change in file %s", key)
			return false
		}
	}

	for key := range newFiles {
		if _, exists := oldFiles[key]; !exists {
			logger.Tracef("Detected file added: %s", key)
			return false
		}
	}

	return true
}
