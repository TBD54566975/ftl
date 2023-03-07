// Package torres manages FTL drives.
package torres

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/alecthomas/errors"
	"github.com/fsnotify/fsnotify"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"

	"github.com/TBD54566975/ftl/common/exec"
	ftlv1 "github.com/TBD54566975/ftl/common/gen/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/common/log"
	"github.com/TBD54566975/ftl/common/plugin"
)

// ModuleConfig is the configuration for an FTL module.
//
// Module config files are currently TOML.
type ModuleConfig struct {
	Language string `toml:"language"`
}

type driveContext struct {
	*plugin.Plugin[ftlv1.DriveServiceClient]
	root       string
	workingDir string
}

type Engineer struct {
	lock    sync.RWMutex
	watcher *fsnotify.Watcher
	drives  map[string]driveContext
	wg      *errgroup.Group
}

// New creates a new Engineer.
func New(ctx context.Context) (*Engineer, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	e := &Engineer{
		watcher: watcher,
		drives:  map[string]driveContext{},
		wg:      &errgroup.Group{},
	}
	e.wg.Go(func() error {
		return e.watch(ctx)
	})
	return e, nil
}

// Watch FTL modules for changes and notify the Drives.
func (e *Engineer) watch(ctx context.Context) error {
	logger := log.FromContext(ctx)
	for {
		select {
		case event := <-e.watcher.Events:
			if event.Op&fsnotify.Chmod == fsnotify.Chmod {
				continue
			}
			path := event.Name
			e.lock.Lock()
			for root, drive := range e.drives {
				if strings.HasPrefix(path, root) {
					_, err := drive.Client.FileChange(ctx, &ftlv1.FileChangeRequest{Path: path})
					if err != nil {
						e.lock.Unlock()
						return errors.WithStack(err)
					}
				}
			}
			e.lock.Unlock()

		case err := <-e.watcher.Errors:
			logger.Warn("File watcher error", "err", err)
			return err

		case <-ctx.Done():
			return nil
		}
	}
}

// Drives returns the list of active drives.
func (e *Engineer) Drives() []string {
	e.lock.RLock()
	defer e.lock.Unlock()
	return maps.Keys(e.drives)
}

// Manage starts a new Drive to manage a directory of functions.
//
// The Drive executable must have the name ftl-drive-<language>. The Engineer
// will pass the following envars through to the Drive:
//
//	FTL_DRIVE_SOCKET - Path to a Unix socket that the Drive must serve the gRPC service xyz.block.ftl.v1.DriveService on.
//	FTL_MODULE_ROOT - Path to a directory containing FTL module source and an ftl.toml file.
//	FTL_WORKING_DIR - Path to a directory that the Drive can use for temporary files.
func (e *Engineer) Manage(ctx context.Context, dir string) (err error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	dir, err = filepath.Abs(dir)
	if err != nil {
		return errors.WithStack(err)
	}

	// Load the config.
	path := filepath.Join(dir, "ftl.toml")
	config := ModuleConfig{}
	_, err = toml.DecodeFile(path, &config)
	if err != nil {
		return errors.WithStack(err)
	}

	// Find the Drive executable.
	exeName := "ftl-drive-" + config.Language
	exe, err := exec.LookPath(exeName)
	if err != nil {
		return errors.Wrapf(err, "could not find FTL.drive-%s", config.Language)
	}

	// Setup the working directory for the module.
	workingDir := filepath.Join(dir, ".ftl")
	err = os.MkdirAll(workingDir, 0750)
	if err != nil {
		return errors.WithStack(err)
	}

	cmdCtx, drvPlugin, err := plugin.Spawn(ctx, dir, exe, ftlv1.NewDriveServiceClient,
		plugin.WithEnvars("FTL_MODULE_ROOT="+dir))
	if err != nil {
		return errors.WithStack(err)
	}
	e.wg.Go(func() error {
		<-cmdCtx.Done()
		return errors.WithStack(cmdCtx.Err())
	})

	// Ensure we stop the sub-process if anything errors.
	defer func() {
		if err != nil {
			_ = drvPlugin.Cmd.Kill(syscall.SIGKILL)
		}
	}()

	err = e.watcher.Add(dir)
	if err != nil {
		return errors.WithStack(err)
	}

	e.drives[dir] = driveContext{
		Plugin:     drvPlugin,
		root:       dir,
		workingDir: workingDir,
	}
	return nil
}

func (e *Engineer) Wait() error {
	return errors.WithStack(e.wg.Wait())
}
