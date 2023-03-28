// Package agent runs on developer machines, facilitating hot reloading and routing.
package agent

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
	"github.com/alecthomas/atomic"
	"github.com/alecthomas/errors"
	"github.com/fsnotify/fsnotify"
	option "github.com/jordan-bonecutter/goption"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/TBD54566975/ftl/common/exec"
	"github.com/TBD54566975/ftl/common/log"
	"github.com/TBD54566975/ftl/common/metadata"
	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/common/pubsub"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	pschema "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/schema"
	sdkgo "github.com/TBD54566975/ftl/sdk-go"
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
	schema       atomic.Value[option.Option[*schema.Module]]
	root         string
	workingDir   string
	config       ModuleConfig
}

type Agent struct {
	lock          sync.RWMutex
	watcher       *fsnotify.Watcher
	drives        map[string]*driveContext
	schemaChanges *pubsub.Topic[*schema.Module]
	wg            *errgroup.Group
}

var _ http.Handler = (*Agent)(nil)
var _ ftlv1.VerbServiceServer = (*Agent)(nil)
var _ ftlv1.DevelServiceServer = (*Agent)(nil)

// New creates a new local agent.
func New(ctx context.Context) (*Agent, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	e := &Agent{
		watcher:       watcher,
		drives:        map[string]*driveContext{},
		wg:            &errgroup.Group{},
		schemaChanges: pubsub.New[*schema.Module](),
	}
	e.wg.Go(func() error { return e.watch(ctx) })
	return e, nil
}

// Drives returns the list of active drives.
func (l *Agent) Drives() []string {
	l.lock.RLock()
	defer l.lock.Unlock()
	return maps.Keys(l.drives)
}

// Manage starts a new Drive to manage a directory of functions.
//
// The Drive executable must have the name ftl-drive-$LANG. The Agent
// will pass the following envars through to the Drive:
//
//	FTL_MODULE_ROOT - Path to a directory containing FTL module source and an ftl.toml file.
//	FTL_WORKING_DIR - Path to a directory that the Drive can use for temporary files.
//	FTL_MODULE - The name of the module.
func (l *Agent) Manage(ctx context.Context, dir string) (err error) {
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
		return errors.Wrapf(err, "could not find ftl-drive-%s", config.Language)
	}

	// Setup the working directory for the module.
	workingDir := filepath.Join(dir, ".ftl")
	err = os.MkdirAll(workingDir, 0750)
	if err != nil {
		return errors.WithStack(err)
	}

	var develClient ftlv1.DevelServiceClient
	verbServicePlugin, cmdCtx, err := plugin.Spawn(ctx, config.Module, dir, exe, ftlv1.NewVerbServiceClient,
		plugin.WithEnvars("FTL_MODULE_ROOT="+dir),
		plugin.WithEnvars("FTL_MODULE="+config.Module),
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

	dctx := &driveContext{
		Plugin:       verbServicePlugin,
		develService: develClient,
		root:         dir,
		workingDir:   workingDir,
		config:       config,
	}
	l.wg.Go(func() error { return l.syncSchemaFromDrive(cmdCtx, dctx) })
	l.drives[dir] = dctx
	return nil
}

func (l *Agent) Wait() error {
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
func (l *Agent) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cleanedPath := strings.TrimPrefix(r.URL.Path, "/")

	// Root of server lists all the verbs.
	if cleanedPath == "" {
		resp, err := l.List(r.Context(), &ftlv1.ListRequest{})
		if err != nil {
			writeError(w, http.StatusBadGateway, "failed to list Verbs: %s", err)
			return
		}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	verb, err := sdkgo.ParseVerbRef(cleanedPath)
	if err != nil {
		writeError(w, http.StatusBadRequest, "%s", err.Error())
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

	resp, err := l.Call(r.Context(), &ftlv1.CallRequest{Verb: verb.ToProto(), Body: req})
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

func (l *Agent) Call(ctx context.Context, req *ftlv1.CallRequest) (*ftlv1.CallResponse, error) {
	logger := log.FromContext(ctx)
	logger.Infof("Calling %s", req.Verb)
	ctx = metadata.WithDirectRouting(ctx)
	drive, err := l.findDrive(req.Verb.ToFTL())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return drive.Client.Call(ctx, req)
}

// Send implements ftlv1.VerbServiceServer
func (*Agent) Send(context.Context, *ftlv1.SendRequest) (*ftlv1.SendResponse, error) {
	panic("unimplemented")
}

func (l *Agent) List(ctx context.Context, req *ftlv1.ListRequest) (*ftlv1.ListResponse, error) {
	ctx = metadata.WithDirectRouting(ctx)
	out := &ftlv1.ListResponse{}
	for _, drive := range l.allDrives() {
		resp, err := drive.Client.List(ctx, proto.Clone(req).(*ftlv1.ListRequest))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list module %q", drive.config.Module)
		}
		out.Verbs = append(out.Verbs, resp.Verbs...)
	}
	slices.SortFunc(out.Verbs, func(a, b *pschema.VerbRef) bool {
		return a.Module < b.Module || (a.Module == b.Module && a.Name < b.Name)
	})
	return out, nil
}

func (l *Agent) Ping(ctx context.Context, req *ftlv1.PingRequest) (*ftlv1.PingResponse, error) {
	return &ftlv1.PingResponse{}, nil
}

func (*Agent) FileChange(context.Context, *ftlv1.FileChangeRequest) (*ftlv1.FileChangeResponse, error) {
	return nil, errors.WithStack(status.Error(codes.Unimplemented, "not implemented"))
}

func (l *Agent) SyncSchema(stream ftlv1.DevelService_SyncSchemaServer) error {
	drives := l.allDrives()
	slices.SortStableFunc(drives, func(a, b *driveContext) bool {
		return a.config.Module < b.config.Module
	})
	for i, drive := range drives {
		module, ok := drive.schema.Load().Get()
		if !ok {
			continue
		}
		err := stream.Send(&ftlv1.SyncSchemaResponse{ //nolint:forcetypeassert
			Schema: module.ToProto().(*pschema.Module),
			More:   i < len(drives)-1,
		})
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (l *Agent) allDrives() []*driveContext {
	l.lock.Lock()
	defer l.lock.Unlock()
	out := make([]*driveContext, 0, len(l.drives))
	for _, drive := range l.drives {
		out = append(out, drive)
	}
	return out
}

func (l *Agent) findDrive(verb string) (*driveContext, error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	modules := make([]string, 0, len(l.drives))
	for _, drive := range l.drives {
		modules = append(modules, drive.config.Module)
		if strings.HasPrefix(verb, drive.config.Module) {
			return drive, nil
		}
	}
	return nil, errors.Errorf("could not find module serving Verb %q among %s", verb, strings.Join(modules, ", "))
}

func (l *Agent) syncSchemaFromDrive(ctx context.Context, drive *driveContext) error {
	stream, err := drive.develService.SyncSchema(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	wg, _ := errgroup.WithContext(ctx)

	// Receive schema changes from the drive.
	wg.Go(func() error {
		for {
			resp, err := stream.Recv()
			if err != nil {
				return errors.WithStack(err)
			}
			module := schema.ProtoToModule(resp.Schema)
			l.schemaChanges.Publish(module)
			drive.schema.Store(option.Some(module))
		}
	})

	// Send snapshot.
	drives := l.allDrives()
	for _, drive := range drives {
		module, ok := drive.schema.Load().Get()
		if !ok {
			continue
		}
		if err := l.sendSchema(stream, module); err != nil {
			return errors.WithStack(err)
		}
	}
	// Send changes to the drive.
	changes := l.schemaChanges.Subscribe(make(chan *schema.Module, 64))
	wg.Go(func() error {
		for module := range changes {
			// Don't send updates back to itself.
			if module.Name == drive.config.Module {
				continue
			}
			if err := l.sendSchema(stream, module); err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	})

	return errors.WithStack(wg.Wait())
}

func (l *Agent) sendSchema(stream ftlv1.DevelService_SyncSchemaClient, module *schema.Module) error {
	err := stream.Send(&ftlv1.SyncSchemaRequest{ //nolint:forcetypeassert
		Schema: module.ToProto().(*pschema.Module),
	})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Watch FTL modules for changes and notify the Drives.
func (l *Agent) watch(ctx context.Context) error {
	logger := log.FromContext(ctx)
	for {
		select {
		case event := <-l.watcher.Events:
			if event.Op&fsnotify.Chmod == fsnotify.Chmod {
				continue
			}
			path := event.Name
			logger.Debugf("File changed, notifying drives: %s", path)
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
			logger.Warnf("File watcher error: %s", err)
			return err

		case <-ctx.Done():
			return nil
		}
	}
}
