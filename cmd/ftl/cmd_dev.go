package main

import (
	"bufio"
	"context"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/bmatcuk/doublestar/v4"

	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/moduleconfig"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

type moduleFolderInfo struct {
	ModuleName  string
	NumFiles    int
	LastModTime time.Time
}

type devCmd struct {
	BaseDir        string        `arg:"" help:"Directory to watch for FTL modules" type:"existingdir" default:"."`
	Watch          time.Duration `help:"Watch template directory at this frequency and regenerate on change." default:"500ms"`
	FailureDelay   time.Duration `help:"Delay before retrying a failed deploy." default:"5s"`
	ReconnectDelay time.Duration `help:"Delay before attempting to reconnect to FTL." default:"5s"`
}

type moduleMap map[string]*moduleFolderInfo

func (m *moduleMap) ForceRebuild(dir string) {
	(*m)[dir].NumFiles = 0
	(*m)[dir].LastModTime = time.Now()
}

func (m *moduleMap) AddModule(dir string, module string) {
	(*m)[dir] = &moduleFolderInfo{
		ModuleName:  module,
		LastModTime: time.Now(),
	}
}

func (m *moduleMap) RemoveModule(dir string) {
	delete(*m, dir)
}

func (m *moduleMap) ForceRebuildFromDependent(module string) {
	for dir, moduleInfo := range *m {
		if moduleInfo.ModuleName != module {
			(*m).ForceRebuild(dir)
		}
	}
}

func (d *devCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	logger := log.FromContext(ctx)
	logger.Infof("Watching %s for FTL modules", d.BaseDir)

	schemaChanges := make(chan string)
	modules := make(moduleMap)

	// Start a goroutine to watch for schema changes
	go d.watchForSchemaChanges(ctx, client, schemaChanges)

	lastScanTime := time.Now()
	for {
		delay := d.Watch
		iterationStartTime := time.Now()

		tomls, err := d.getTomls(ctx)
		if err != nil {
			return err
		}

		err = d.addOrRemoveModules(tomls, modules)
		if err != nil {
			return err
		}

		for dir := range modules {
			currentModule := modules[dir]
			err := d.updateFileInfo(ctx, dir, modules)
			if err != nil {
				return err
			}

			if currentModule.NumFiles != modules[dir].NumFiles || modules[dir].LastModTime.After(lastScanTime) {
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
				}
			}
		}

		lastScanTime = iterationStartTime
		select {
		case moduleName := <-schemaChanges:
			logger.Infof("Schema change detected for module %s, rebuilding other modules.", moduleName)
			modules.ForceRebuildFromDependent(moduleName)
		case <-time.After(delay):
		case <-ctx.Done():
			return nil
		}
	}
}

func (d *devCmd) watchForSchemaChanges(ctx context.Context, client ftlv1connect.ControllerServiceClient, schemaChanges chan string) {
	logger := log.FromContext(ctx)
	for {
		stream, err := client.PullSchema(ctx, connect.NewRequest(&ftlv1.PullSchemaRequest{}))
		if err != nil {
			logger.Errorf(err, "Error connecting to FTL. Will retry...")
			time.Sleep(d.ReconnectDelay)
			continue
		}

		for stream.Receive() {
			msg := stream.Msg()
			if msg.ChangeType == ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGED {
				schemaChanges <- msg.ModuleName
			}
		}

		stream.Close()
		logger.Infof("Stream disconnected, attempting to reconnect...")
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

func (d *devCmd) updateFileInfo(ctx context.Context, dir string, modules moduleMap) error {
	config, err := moduleconfig.LoadConfig(dir)
	if err != nil {
		return err
	}

	ignores := initGitIgnore(ctx, dir)

	var changed string
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
				fileInfo, err := entry.Info()
				if err != nil {
					return err
				}

				module := modules[dir]
				module.NumFiles++
				if fileInfo.ModTime().After(module.LastModTime) {
					changed = srcPath
					module.LastModTime = fileInfo.ModTime()
				}
				modules[dir] = module
			}
		}

		return nil
	})

	if changed != "" {
		log.FromContext(ctx).Tracef("Detected change in %s", changed)
	}

	return err
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
