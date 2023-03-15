// Package local manages locally running FTL drives.
package local

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
	"google.golang.org/protobuf/proto"

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
	Module   string `toml:"module"`
}

type driveContext struct {
	*plugin.Plugin[ftlv1.VerbServiceClient]
	develService ftlv1.DevelServiceClient
	root         string
	workingDir   string
	config       ModuleConfig
}

type Local struct {
	lock    sync.RWMutex
	watcher *fsnotify.Watcher
	drives  map[string]driveContext
	wg      *errgroup.Group
}

var _ ftlv1.AgentServiceServer = (*Local)(nil)
var _ http.Handler = (*Local)(nil)

// New creates a new Local drive coordinator.
func New(ctx context.Context) (*Local, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	e := &Local{
		watcher: watcher,
		drives:  map[string]driveContext{},
		wg:      &errgroup.Group{},
	}
	e.wg.Go(func() error {
		return e.watch(ctx)
	})
	return e, nil
}

// Drives returns the list of active drives.
func (l *Local) Drives() []string {
	l.lock.RLock()
	defer l.lock.Unlock()
	return maps.Keys(l.drives)
}

// Manage starts a new Drive to manage a directory of functions.
//
// The Drive executable must have the name ftl-drive-$LANG. The Local
// will pass the following envars through to the Drive:
//
//	FTL_DRIVE_SOCKET - Path to a Unix socket that the Drive must serve the gRPC service xyz.block.ftl.v1.DriveService on.
//	FTL_MODULE_ROOT - Path to a directory containing FTL module source and an ftl.toml file.
//	FTL_WORKING_DIR - Path to a directory that the Drive can use for temporary files.
func (l *Local) Manage(ctx context.Context, dir string) (err error) {
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

	var develClient ftlv1.DevelServiceClient
	verbServicePlugin, cmdCtx, err := plugin.Spawn(ctx, dir, exe, ftlv1.NewVerbServiceClient,
		plugin.WithEnvars("FTL_MODULE_ROOT="+dir),
		plugin.WithExtraClient(&develClient, ftlv1.NewDevelServiceClient),
	)
	if err != nil {
		return errors.WithStack(err)
	}
	l.wg.Go(func() error {
		<-cmdCtx.Done()
		return errors.WithStack(cmdCtx.Err())
	})

	// Ensure we stop the sub-process if anything errors.
	defer func() {
		if err != nil {
			_ = verbServicePlugin.Cmd.Kill(syscall.SIGKILL)
		}
	}()

	l.lock.Lock()
	defer l.lock.Unlock()

	err = l.watcher.Add(dir)
	if err != nil {
		return errors.WithStack(err)
	}

	l.drives[dir] = driveContext{
		Plugin:     verbServicePlugin,
		root:       dir,
		workingDir: workingDir,
		config:     config,
	}
	return nil
}

func (l *Local) Wait() error {
	return errors.WithStack(l.wg.Wait())
}

type Error struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, msg string, args ...any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{fmt.Sprintf(msg, args...)})
}

// ServeHTTP because we want the local agent to also be able to serve Verbs directly via HTTP.
func (l *Local) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	verb := strings.TrimPrefix(r.URL.Path, "/")

	// Root of server lists all the verbs.
	if verb == "" {
		resp, err := l.List(r.Context(), &ftlv1.ListRequest{})
		if err != nil {
			writeError(w, http.StatusBadGateway, "failed to list Verbs %q: %s", verb, err)
			return
		}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	req, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read request: %s", err)
		return
	}

	if len(req) == 0 {
		req = []byte("{}")
	}

	resp, err := l.Call(r.Context(), &ftlv1.CallRequest{Verb: verb, Body: req})
	if err != nil {
		writeError(w, http.StatusBadGateway, "failed to call Verb %q: %s", verb, err)
		return
	}
	switch resp := resp.Response.(type) {
	case *ftlv1.CallResponse_Error_:
		writeError(w, http.StatusInternalServerError, "verb failed: %s", resp.Error.Message)

	case *ftlv1.CallResponse_Body:
		_, _ = w.Write(resp.Body)
	}
}

func (l *Local) Serve(ctx context.Context, req *ftlv1.ServeRequest) (*ftlv1.ServeResponse, error) {
	err := l.Manage(ctx, req.Path)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &ftlv1.ServeResponse{}, nil
}

func (l *Local) Call(ctx context.Context, req *ftlv1.CallRequest) (*ftlv1.CallResponse, error) {
	drive, err := l.findDrive(req.Verb)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return drive.Client.Call(ctx, req)
}

func (l *Local) List(ctx context.Context, req *ftlv1.ListRequest) (*ftlv1.ListResponse, error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	out := &ftlv1.ListResponse{}
	for _, drive := range l.drives {
		resp, err := drive.Client.List(ctx, proto.Clone(req).(*ftlv1.ListRequest))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list module %q", drive.config.Module)
		}
		out.Verbs = append(out.Verbs, resp.Verbs...)
	}
	return out, nil
}

func (l *Local) Ping(ctx context.Context, req *ftlv1.PingRequest) (*ftlv1.PingResponse, error) {
	return &ftlv1.PingResponse{}, nil
}

func (l *Local) findDrive(verb string) (driveContext, error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	modules := make([]string, 0, len(l.drives))
	for _, drive := range l.drives {
		modules = append(modules, drive.config.Module)
		if strings.HasPrefix(verb, drive.config.Module) {
			return drive, nil
		}
	}
	return driveContext{}, errors.Errorf("could not find module serving Verb %q among %s", verb, strings.Join(modules, ", "))
}

// Watch FTL modules for changes and notify the Drives.
func (l *Local) watch(ctx context.Context) error {
	logger := log.FromContext(ctx)
	for {
		select {
		case event := <-l.watcher.Events:
			if event.Op&fsnotify.Chmod == fsnotify.Chmod {
				continue
			}
			path := event.Name
			l.lock.Lock()
			for root, drive := range l.drives {
				if strings.HasPrefix(path, root) {
					_, err := drive.develService.FileChange(ctx, &ftlv1.FileChangeRequest{Path: path})
					if err != nil {
						l.lock.Unlock()
						return errors.WithStack(err)
					}
				}
			}
			l.lock.Unlock()

		case err := <-l.watcher.Errors:
			logger.Warn("File watcher error", "err", err)
			return err

		case <-ctx.Done():
			return nil
		}
	}
}
