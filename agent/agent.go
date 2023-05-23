// Package agent runs on developer machines, facilitating hot reloading and routing.
package agent

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/alecthomas/atomic"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types"
	"github.com/bufbuild/connect-go"
	grpcreflect "github.com/bufbuild/connect-grpcreflect-go"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"

	"github.com/TBD54566975/ftl/common/exec"
	"github.com/TBD54566975/ftl/common/log"
	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/common/pubsub"
	"github.com/TBD54566975/ftl/common/rpc"
	"github.com/TBD54566975/ftl/common/server"
	"github.com/TBD54566975/ftl/console"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	pschema "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/schema"
)

type driveContext struct {
	*plugin.Plugin[ftlv1connect.VerbServiceClient]
	develService ftlv1connect.DevelServiceClient
	schema       atomic.Value[types.Option[*schema.Module]]
	root         string
	workingDir   string
	config       ModuleConfig
}

type Agent struct {
	lock   sync.RWMutex
	drives map[string]*driveContext
	// Each time any module managed by the agent changes, a new schema is
	// published on this topic.
	schemaChanges *pubsub.Topic[*schema.Module]
	wg            *errgroup.Group
	listen        *url.URL
}

var _ ftlv1connect.VerbServiceHandler = (*Agent)(nil)
var _ ftlv1connect.DevelServiceHandler = (*Agent)(nil)

// New creates a new local agent.
func New(ctx context.Context, listen *url.URL) (*Agent, error) {
	e := &Agent{
		drives:        map[string]*driveContext{},
		wg:            &errgroup.Group{},
		schemaChanges: pubsub.New[*schema.Module](),
		listen:        listen,
	}
	return e, nil
}

// Serve starts the agent server.
func (a *Agent) Serve(ctx context.Context) error {
	reflector := grpcreflect.NewStaticReflector(ftlv1connect.DevelServiceName, ftlv1connect.VerbServiceName)
	c, err := console.Server(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	return server.Serve(ctx, a.listen,
		server.GRPC(ftlv1connect.NewDevelServiceHandler, a),
		server.GRPC(ftlv1connect.NewVerbServiceHandler, a),
		server.Route(grpcreflect.NewHandlerV1(reflector)),
		server.Route(grpcreflect.NewHandlerV1Alpha(reflector)),
		server.Route("/", c),
	)
}

// Drives returns the list of active drives.
func (a *Agent) Drives() []string {
	a.lock.RLock()
	defer a.lock.Unlock()
	return maps.Keys(a.drives)
}

// Manage starts a new Drive to manage a directory of functions.
//
// The Drive executable must have the name ftl-drive-$LANG. The Agent
// will pass the following envars through to the Drive:
//
//	FTL_MODULE_ROOT - Path to a directory containing FTL module source and an ftl.toml file.
//	FTL_WORKING_DIR - Path to a directory that the Drive can use for temporary files.
//	FTL_MODULE - The name of the module.
func (a *Agent) Manage(ctx context.Context, dir string) (err error) {
	dir, err = filepath.Abs(dir)
	if err != nil {
		return errors.WithStack(err)
	}

	// Load the config.
	config, err := LoadConfig(dir)
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

	var develClient ftlv1connect.DevelServiceClient
	verbServicePlugin, cmdCtx, err := plugin.Spawn(ctx, config.Module, dir, exe, ftlv1connect.NewVerbServiceClient,
		plugin.WithEnvars(
			"FTL_MODULE_ROOT="+dir,
			"FTL_MODULE="+config.Module,
			// Used by sub-processes to call back into FTL.
			"FTL_ENDPOINT="+a.listen.String(),
		),
		plugin.WithExtraClient(&develClient, ftlv1connect.NewDevelServiceClient),
	)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = develClient.Ping(ctx, connect.NewRequest(&ftlv1.PingRequest{}))
	if err != nil {
		return errors.WithStack(err)
	}

	a.wg.Go(func() error {
		<-cmdCtx.Done()
		return errors.WithStack(cmdCtx.Err())
	})

	// Ensure we stop the sub-process if anything errors.
	defer func() {
		if err != nil {
			_ = verbServicePlugin.Cmd.Kill(syscall.SIGKILL)
		}
	}()

	a.lock.Lock()
	defer a.lock.Unlock()

	dctx := &driveContext{
		Plugin:       verbServicePlugin,
		develService: develClient,
		root:         dir,
		workingDir:   workingDir,
		config:       config,
	}
	a.wg.Go(func() error { return a.syncSchemaFromDrive(cmdCtx, dctx) })
	a.wg.Go(func() error { return a.syncSchemaToDrive(cmdCtx, dctx) })
	a.drives[dir] = dctx
	return nil
}

func (a *Agent) Wait() error {
	return errors.WithStack(a.wg.Wait())
}

func (a *Agent) PullSchema(ctx context.Context, req *connect.Request[ftlv1.PullSchemaRequest], stream *connect.ServerStream[ftlv1.PullSchemaResponse]) error {
	return a.watchModuleChanges(ctx, "", func(m *schema.Module, more bool) error {
		return errors.WithStack(stream.Send(&ftlv1.PullSchemaResponse{
			Schema: m.ToProto().(*pschema.Module), //nolint:forcetypeassert
			More:   more,
		}))
	})
}

func (a *Agent) PushSchema(ctx context.Context, req *connect.ClientStream[ftlv1.PushSchemaRequest]) (*connect.Response[ftlv1.PushSchemaResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("FTL agent does not support PushSchema"))
}

func (a *Agent) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	logger := log.FromContext(ctx)
	logger.Infof("Calling %s", req.Msg.Verb)
	ctx = rpc.WithDirectRouting(ctx)
	drive, err := a.findDrive(req.Msg.Verb.ToFTL())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return drive.Client.Call(ctx, req)
}

func (a *Agent) List(ctx context.Context, req *connect.Request[ftlv1.ListRequest]) (*connect.Response[ftlv1.ListResponse], error) {
	ctx = rpc.WithDirectRouting(ctx)
	out := &ftlv1.ListResponse{}

	for _, drive := range a.allDrives() {
		resp, err := drive.Client.List(ctx, req)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list module %q", drive.config.Module)
		}
		out.Verbs = append(out.Verbs, resp.Msg.Verbs...)
	}

	slices.SortFunc(out.Verbs, func(a, b *pschema.VerbRef) bool {
		return a.Module < b.Module || (a.Module == b.Module && a.Name < b.Name)
	})

	return connect.NewResponse(out), nil
}

func (*Agent) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (a *Agent) allDrives() []*driveContext {
	a.lock.Lock()
	defer a.lock.Unlock()
	out := make([]*driveContext, 0, len(a.drives))
	for _, drive := range a.drives {
		out = append(out, drive)
	}
	return out
}

func (a *Agent) findDrive(verb string) (*driveContext, error) {
	a.lock.Lock()
	defer a.lock.Unlock()
	modules := make([]string, 0, len(a.drives))
	for _, drive := range a.drives {
		modules = append(modules, drive.config.Module)
		if strings.HasPrefix(verb, drive.config.Module) {
			return drive, nil
		}
	}
	return nil, errors.Errorf("could not find module serving Verb %q among %s", verb, strings.Join(modules, ", "))
}

func (a *Agent) syncSchemaToDrive(ctx context.Context, drive *driveContext) error {
	stream := drive.develService.PushSchema(ctx)
	return a.watchModuleChanges(ctx, drive.config.Module, func(module *schema.Module, more bool) error {
		return errors.WithStack(stream.Send(&ftlv1.PushSchemaRequest{
			Schema: module.ToProto().(*pschema.Module), //nolint:forcetypeassert
		}))
	})
}

func (a *Agent) watchModuleChanges(ctx context.Context, skipModule string, sendChange func(module *schema.Module, more bool) error) error {
	// Send snapshot.
	drives := a.allDrives()
	for i, drive := range drives {
		module, ok := drive.schema.Load().Get()
		if !ok {
			continue
		}
		if err := sendChange(module, i < len(drives)-1); err != nil {
			return errors.WithStack(err)
		}
	}

	// Send changes to the drive.
	changes := a.schemaChanges.Subscribe(make(chan *schema.Module, 64))
	defer a.schemaChanges.Unsubscribe(changes)

	for {
		select {
		case <-ctx.Done():
			return nil

		case module := <-changes:
			// Don't send updates back to itself.
			if module.Name == skipModule {
				continue
			}
			if err := sendChange(module, false); err != nil {
				return errors.WithStack(err)
			}
		}
	}
}

func (a *Agent) syncSchemaFromDrive(ctx context.Context, drive *driveContext) error {
	stream, err := drive.develService.PullSchema(ctx, connect.NewRequest(&ftlv1.PullSchemaRequest{}))
	if err != nil {
		return errors.WithStack(err)
	}

	for stream.Receive() {
		resp := stream.Msg()
		module := schema.ModuleFromProto(resp.Schema)
		a.schemaChanges.Publish(module)
		drive.schema.Store(types.Some(module))
	}
	fmt.Println("DONE", stream.Err())
	return errors.WithStack(stream.Err())
}
