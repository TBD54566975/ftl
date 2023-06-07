package drivego

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/alecthomas/atomic"
	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"
	"github.com/fsnotify/fsnotify"
	"golang.org/x/mod/modfile"
	"golang.org/x/sync/errgroup"
	"golang.org/x/tools/go/packages"

	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/drive-go/codewriter"
	"github.com/TBD54566975/ftl/internal/eventsource"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	pschema "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/schema"
	sdkgo "github.com/TBD54566975/ftl/sdk-go"
)

// Base import for all Go FTL modules.
const syntheticGoPath = "github.com/TBD54566975/ftl/examples"

type Config struct {
	Endpoint   *url.URL `required:"" help:"FTL endpoint to connect to." env:"FTL_ENDPOINT"`
	Module     string   `required:"" env:"FTL_MODULE" help:"The FTL module as configured."`
	FTLSource  string   `required:"" type:"existingdir" env:"FTL_SOURCE" help:"Path to FTL source code when developing locally."`
	WorkingDir string   `required:"" type:"existingdir" env:"FTL_WORKING_DIR" help:"Working directory for FTL runtime."`
	Dir        string   `required:"" type:"existingdir" env:"FTL_MODULE_ROOT" help:"Directory to root of Go FTL module"`
}

// Run creates and starts a new DevelService for a directory of Go Verbs.
func Run(ctx context.Context, config Config) (context.Context, *Server, error) {
	logger := log.FromContext(ctx)
	goModFile := filepath.Join(config.Dir, "go.mod")
	_, err := os.Stat(goModFile)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "go.mod not found in %s", config.Dir)
	}
	goModule, err := findGoModule(goModFile)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	s := &Server{
		Config:         config,
		agent:          rpc.Dial(ftlv1connect.NewDevelServiceClient, config.Endpoint.String(), log.Error),
		router:         rpc.Dial(ftlv1connect.NewVerbServiceClient, config.Endpoint.String(), log.Error),
		module:         config.Module,
		wg:             &errgroup.Group{},
		goModule:       goModule,
		triggerRebuild: make(chan struct{}, 64),
		moduleSchema:   eventsource.New[*schema.Module](),
	}

	logger.Infof("Starting")

	s.wg.Go(func() error { return s.watchForChanges(ctx) })

	// Build and start the sub-process.
	exe, err := s.rebuild(ctx)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	plugin, cmdCtx, err := plugin.Spawn(ctx, "", s.Config.Dir, exe, ftlv1connect.NewVerbServiceClient,
		plugin.WithEnvars("FTL_MODULE="+s.module))
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	s.plugin.Store(plugin)

	s.wg.Go(func() error { return s.restartModuleOnExit(ctx, cmdCtx) })

	logger.Infof("Online")
	return ctx, s, nil
}

var _ ftlv1connect.DevelServiceHandler = (*Server)(nil)
var _ ftlv1connect.VerbServiceHandler = (*Server)(nil)

type Server struct {
	Config
	agent          ftlv1connect.DevelServiceClient
	router         ftlv1connect.VerbServiceClient
	module         string
	goModule       *modfile.File
	wg             *errgroup.Group
	triggerRebuild chan struct{}

	fullSchema   atomic.Value[schema.Schema]
	moduleSchema *eventsource.EventSource[*schema.Module]

	plugin atomic.Value[*plugin.Plugin[ftlv1connect.VerbServiceClient]]

	rebuildMu   sync.Mutex
	lastRebuild time.Time
}

// Wait for the server to exit.
func (s *Server) Wait() error {
	return errors.WithStack(s.wg.Wait())
}

func (s *Server) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	if rpc.IsDirectRouted(ctx) || req.Msg.Verb.Module == s.module {
		return s.plugin.Load().Client.Call(ctx, req)
	}
	resp, err := s.router.Call(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return resp, nil
}

func (s *Server) List(ctx context.Context, req *connect.Request[ftlv1.ListRequest]) (*connect.Response[ftlv1.ListResponse], error) {
	if !rpc.IsDirectRouted(ctx) {
		resp, err := s.router.List(ctx, req)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return resp, nil
	}
	out := &ftlv1.ListResponse{}
	module := s.moduleSchema.Load()
	for _, verb := range module.Decls {
		if verb, ok := verb.(*schema.Verb); ok {
			out.Verbs = append(out.Verbs, &pschema.VerbRef{Module: module.Name, Name: verb.Name})
		}
	}
	return connect.NewResponse(out), nil
}

func (s *Server) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *Server) PullSchema(ctx context.Context, req *connect.Request[ftlv1.PullSchemaRequest], stream *connect.ServerStream[ftlv1.PullSchemaResponse]) error {
	// Send initial and subsequent schema updates.
	err := s.sendSchema(stream, s.moduleSchema.Load())
	if err != nil {
		return errors.WithStack(err)
	}
	changes := s.moduleSchema.Subscribe(make(chan *schema.Module, 64))
	defer s.moduleSchema.Unsubscribe(changes)

	for {
		select {
		case <-ctx.Done():
			return nil

		case module := <-changes:
			if err := s.sendSchema(stream, module); err != nil {
				return errors.WithStack(err)
			}
		}
	}
}

func (s *Server) PushSchema(ctx context.Context, stream *connect.ClientStream[ftlv1.PushSchemaRequest]) (*connect.Response[ftlv1.PushSchemaResponse], error) {
	logger := log.FromContext(ctx)
	for stream.Receive() {
		received := stream.Msg()
		module, err := schema.ModuleFromProto(received.Schema)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid schema"))
		}

		logger.Debugf("Received schema update from %s", module.Name)

		fullSchema := s.fullSchema.Load()
		oldHash := fullSchema.Hash()
		fullSchema.Upsert(module)
		newHash := fullSchema.Hash()
		if oldHash != newHash {
			s.fullSchema.Store(fullSchema)
			s.triggerRebuild <- struct{}{}
		}
	}
	return connect.NewResponse(&ftlv1.PushSchemaResponse{}), nil
}

// func (s *Server) _FileChange(ctx context.Context, req *ftlv1.FileChangeRequest) (*ftlv1.FileChangeResponse, error) {
// 	_, err := s.rebuild(ctx)
// 	if err != nil {
// 		return nil, errors.WithStack(err)
// 	}
// 	err = s.plugin.Load().Cmd.Kill(syscall.SIGHUP)
// 	if err != nil {
// 		return nil, errors.WithStack(err)
// 	}
// 	return &ftlv1.FileChangeResponse{}, errors.WithStack(err)
// }

// Restart the FTL hot reload module if it terminates unexpectedly.
func (s *Server) restartModuleOnExit(ctx, cmdCtx context.Context) error {
	logger := log.FromContext(ctx)
	for {
		select {
		case <-ctx.Done():
			_ = s.plugin.Load().Cmd.Kill(syscall.SIGTERM)
			return nil

		case <-cmdCtx.Done():
			err := cmdCtx.Err()
			logger.Warnf("Module exited, restarting: %s", err)

			exe, err := s.rebuild(ctx)
			if err != nil {
				logger.Errorf(err, "Failed to rebuild FTL module")
				continue
			}
			var nextPlugin *plugin.Plugin[ftlv1connect.VerbServiceClient]
			nextPlugin, cmdCtx, err = plugin.Spawn(ctx, s.Config.Module, s.Config.Dir, exe, ftlv1connect.NewVerbServiceClient)
			if err != nil {
				logger.Errorf(err, "Failed to restart FTL module")
				continue
			}
			s.plugin.Store(nextPlugin)

		case <-s.triggerRebuild:
			// This is not efficient.
			_ = s.plugin.Load().Cmd.Kill(syscall.SIGTERM)
		}
	}
}

func (s *Server) rebuild(ctx context.Context) (exe string, err error) {
	s.rebuildMu.Lock()
	defer s.rebuildMu.Unlock()

	exe = filepath.Join(s.WorkingDir, s.module)

	if time.Since(s.lastRebuild) < time.Millisecond*250 {
		return exe, nil
	}
	defer func() { s.lastRebuild = time.Now() }()

	logger := log.FromContext(ctx)

	logger.Infof("Extracting schema")
	module, err := sdkgo.ExtractModule(s.Dir)
	if err != nil {
		return "", errors.WithStack(err)
	}
	s.moduleSchema.Store(module)

	err = s.writeGoLayout()
	if err != nil {
		return "", errors.WithStack(err)
	}
	err = execInRoot(ctx, s.WorkingDir, "go", "mod", "tidy")
	if err != nil {
		return "", errors.WithStack(err)
	}

	logger.Infof("Compiling FTL module...")
	err = execInRoot(ctx, s.WorkingDir, "go", "build", "-trimpath", "-buildvcs=false", "-ldflags=-s -w -buildid=", "-o", exe)
	if err != nil {
		source, merr := ioutil.ReadFile(filepath.Join(s.WorkingDir, "main.go"))
		if merr != nil {
			return "", errors.WithStack(err)
		}
		return "", errors.Wrap(err, string(source))
	}

	return exe, nil
}

func (s *Server) writeGoLayout() error {
	if err := s.writeModules(); err != nil {
		return errors.WithStack(err)
	}
	if err := s.writeMain(); err != nil {
		return errors.WithStack(err)
	}
	if err := s.writeGoMod(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *Server) writeModules() error {
	_ = os.RemoveAll(filepath.Join(s.WorkingDir, "_modules"))
	schema := s.fullSchema.Load()
	for _, module := range schema.Modules {
		err := s.writeModule(module)
		if err != nil {
			return errors.Wrapf(err, "%s: failed to write module", module.Name)
		}
	}
	return nil
}

func (s *Server) writeModule(module *schema.Module) error {
	modulePath := s.modulePath(module)
	err := os.MkdirAll(modulePath, 0750)
	if err != nil {
		return errors.Wrapf(err, "%s: failed to create module directory", module.Name)
	}
	w, err := os.Create(filepath.Join(modulePath, "go.mod"))
	if err != nil {
		return errors.Wrapf(err, "%s: failed to create go.mod file", module.Name)
	}
	defer w.Close() //nolint:gosec

	w, err = os.Create(filepath.Join(modulePath, "schema.go"))
	if err != nil {
		return errors.Wrapf(err, "%s: failed to create schema file", module.Name)
	}
	defer w.Close() //nolint:gosec
	err = sdkgo.Generate(module, w)
	if err != nil {
		return errors.Wrapf(err, "%s: failed to generate schema", module.Name)
	}
	return nil
}

func (s *Server) modulePath(module *schema.Module) string {
	return filepath.Join(s.WorkingDir, "_modules", module.Name)
}

func (s *Server) writeMain() error {
	w, err := generate(s.Config)
	if err != nil {
		return errors.WithStack(err)
	}

	main, err := os.Create(filepath.Join(s.WorkingDir, "main.go"))
	if err != nil {
		return errors.WithStack(err)
	}
	defer main.Close() //nolint:gosec
	_, err = main.WriteString(w.String())
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Write the go.mod file for the FTL module.
func (s *Server) writeGoMod() error {
	managedModules := []*schema.Module{}
	managedModules = append(managedModules, s.fullSchema.Load().Modules...)
	managedModulePaths := map[string]bool{}
	if s.FTLSource != "" {
		managedModulePaths["github.com/TBD54566975/ftl"] = true
	}
	for _, module := range managedModules {
		managedModulePaths[path.Join(syntheticGoPath, module.Name)] = true
	}

	goMod, err := os.Create(filepath.Join(s.WorkingDir, "go.mod"))
	if err != nil {
		return errors.WithStack(err)
	}
	defer goMod.Close() //nolint:gosec
	fmt.Fprintf(goMod, "module main\n")
	fmt.Fprintf(goMod, "require %s v0.0.0\n", s.goModule.Module.Mod.Path)
	fmt.Fprintf(goMod, "replace %s => %s\n", s.goModule.Module.Mod.Path, s.Dir)
	// Apply any custom replace directives that aren't replacing FTL modules.
	for _, replace := range s.goModule.Replace {
		if managedModulePaths[replace.Old.Path] {
			continue
		}
		newPath := replace.New.Path
		if strings.HasPrefix(newPath, ".") {
			newPath, err = filepath.Abs(filepath.Join(s.Dir, newPath))
			if err != nil {
				return errors.WithStack(err)
			}
		}
		fmt.Fprintf(goMod, "replace %s => %s\n", replace.Old.Path, newPath)
	}
	for _, module := range managedModules {
		modulePath := s.modulePath(module)
		goImportPath := path.Join(syntheticGoPath, module.Name)
		fmt.Fprintf(goMod, "require %s v0.0.0\n", goImportPath)
		fmt.Fprintf(goMod, "replace %s => %s\n", goImportPath, modulePath)
	}
	if s.FTLSource != "" {
		fmt.Fprintf(goMod, "require github.com/TBD54566975/ftl v0.0.0\n")
		fmt.Fprintf(goMod, "replace github.com/TBD54566975/ftl => %s\n", s.FTLSource)
	}
	return nil
}

func (s *Server) watchForChanges(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.WithStack(err)
	}
	if err = watcher.Add(s.Dir); err != nil {
		return errors.WithStack(err)
	}

	logger := log.FromContext(ctx)
	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Chmod == fsnotify.Chmod {
				continue
			}
			path := event.Name
			logger.Debugf("File changed, notifying drives: %s", path)

			_, err := s.rebuild(ctx)
			if err != nil {
				return errors.Wrapf(err, "failed to rebuild after file change")
			}
			err = s.plugin.Load().Cmd.Kill(syscall.SIGHUP)
			if err != nil {
				return errors.Wrapf(err, "failed to send SIGHUP to plugin")
			}

		case err := <-watcher.Errors:
			logger.Warnf("File watcher error: %s", err)
			return errors.Wrapf(err, "file watcher error")

		case <-ctx.Done():
			return nil
		}
	}
}

func (s *Server) sendSchema(stream *connect.ServerStream[ftlv1.PullSchemaResponse], module *schema.Module) error {
	err := stream.Send(&ftlv1.PullSchemaResponse{ //nolint:forcetypeassert
		Schema: module.ToProto().(*pschema.Module),
	})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func generate(config Config) (*codewriter.Writer, error) {
	fset := token.NewFileSet()
	pkgs, err := packages.Load(&packages.Config{
		Dir:  config.Dir,
		Fset: fset,
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
	}, "./...")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	endpoints := map[string][]endpoint{}
	for _, pkg := range pkgs {
		pkgEndpoints, err := extractEndpoints(pkg)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		for pkg, pkgEndpoints := range pkgEndpoints {
			endpoints[pkg] = append(endpoints[pkg], pkgEndpoints...)
		}
	}

	w := codewriter.New("main")
	w.Import("context")
	w.Import("github.com/TBD54566975/ftl/drive-go")
	w.Import("github.com/TBD54566975/ftl/common/plugin")
	w.Import(`github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect`)

	w.L(`func main() {`)
	w.In(func(w *codewriter.Writer) {
		w.L(`verbConstructor := drivego.NewUserVerbServer(%q,`, config.Module)
		for pkg, endpoints := range endpoints {
			pkgImp := w.Import(pkg)
			for _, endpoint := range endpoints {
				w.L(`  drivego.Handle(%s.%s),`, pkgImp, endpoint.fn.Name())
			}
		}
		w.L(`)`)
		w.L(`plugin.Start(context.Background(), %q, verbConstructor, ftlv1connect.VerbServiceName, ftlv1connect.NewVerbServiceHandler)`, config.Module)
	})

	w.L(`}`)

	return w, nil
}

type endpoint struct {
	pos       token.Position
	pkg       *packages.Package
	decl      *ast.FuncDecl
	fn        *types.Func
	signature *types.Signature
}

func extractEndpoints(pkg *packages.Package) (endpoints map[string][]endpoint, retErr error) { //nolint:unparam
	endpoints = map[string][]endpoint{}
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(node ast.Node) bool {
			fn, ok := node.(*ast.FuncDecl)
			if !ok || fn.Doc == nil {
				return true
			}
			for _, line := range fn.Doc.List {
				if strings.HasPrefix(line.Text, "//ftl:verb") {
					pos := pkg.Fset.Position(fn.Pos())
					fnt := pkg.TypesInfo.Defs[fn.Name].(*types.Func) //nolint:forcetypeassert
					sig := fnt.Type().(*types.Signature)             //nolint:forcetypeassert
					key := pkg.PkgPath
					endpoints[key] = append(endpoints[key], endpoint{
						pkg:       pkg,
						pos:       pos,
						decl:      fn,
						fn:        fnt,
						signature: sig,
					})
				}
			}
			return true
		})
		if retErr != nil {
			return
		}
	}
	return
}

func findGoModule(file string) (*modfile.File, error) {
	goModContent, err := os.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read go.mod file")
	}
	goModFile, err := modfile.Parse(file, goModContent, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to extract Go module from go.mod file")
	}
	return goModFile, nil
}

func execInRoot(ctx context.Context, root string, command string, args ...string) error {
	cmd := exec.Command(ctx, root, command, args...)
	return errors.WithStack(cmd.Run())
}
