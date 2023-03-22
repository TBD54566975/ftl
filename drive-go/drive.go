package drivego

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/alecthomas/atomic"
	"github.com/alecthomas/errors"
	"golang.org/x/mod/modfile"
	"golang.org/x/sync/errgroup"
	"golang.org/x/tools/go/packages"

	"github.com/TBD54566975/ftl/common/eventsource"
	"github.com/TBD54566975/ftl/common/exec"
	"github.com/TBD54566975/ftl/common/log"
	"github.com/TBD54566975/ftl/common/metadata"
	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/common/socket"
	"github.com/TBD54566975/ftl/drive-go/codewriter"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/schema"
	sdkgo "github.com/TBD54566975/ftl/sdk-go"
)

type Config struct {
	Endpoint   socket.Socket `required:"" help:"FTL endpoint to connect to." env:"FTL_ENDPOINT"`
	Module     string        `required:"" env:"FTL_MODULE" help:"The FTL module as configured."`
	FTLSource  string        `required:"" type:"existingdir" env:"FTL_SOURCE" help:"Path to FTL source code when developing locally."`
	WorkingDir string        `required:"" type:"existingdir" env:"FTL_WORKING_DIR" help:"Working directory for FTL runtime."`
	Dir        string        `required:"" type:"existingdir" env:"FTL_MODULE_ROOT" help:"Directory to root of Go FTL module"`
}

// New creates a new DevelService for a directory of Go Verbs.
func New(ctx context.Context, config Config) (*Server, error) {
	logger := log.FromContext(ctx)
	goModFile := filepath.Join(config.Dir, "go.mod")
	_, err := os.Stat(goModFile)
	if err != nil {
		return nil, errors.Wrapf(err, "go.mod not found in %s", config.Dir)
	}
	goModule, err := findGoModule(goModFile)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	conn, err := socket.DialGRPC(ctx, config.Endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial FTL")
	}
	router := ftlv1.NewVerbServiceClient(conn)

	s := &Server{
		Config:         config,
		router:         router,
		module:         config.Module,
		goModule:       goModule,
		triggerRebuild: make(chan struct{}, 64),
		moduleSchema:   eventsource.New[schema.Module](),
	}

	logger.Infof("Starting drive")

	// Build and start the sub-process.
	exe, err := s.rebuild(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	plugin, cmdCtx, err := plugin.Spawn(ctx, "", s.Config.Dir, exe, ftlv1.NewVerbServiceClient,
		plugin.WithEnvars("FTL_MODULE="+s.module))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	s.plugin.Store(plugin)

	go s.restartModuleOnExit(ctx, cmdCtx)
	return s, nil
}

var _ ftlv1.DevelServiceServer = (*Server)(nil)
var _ ftlv1.VerbServiceServer = (*Server)(nil)

type Server struct {
	Config
	router         ftlv1.VerbServiceClient
	module         string
	goModule       string
	triggerRebuild chan struct{}

	fullSchema   atomic.Value[schema.Schema]
	moduleSchema *eventsource.EventSource[schema.Module]

	plugin atomic.Value[*plugin.Plugin[ftlv1.VerbServiceClient]]

	rebuildMu   sync.Mutex
	lastRebuild time.Time
}

func (s *Server) SyncSchema(stream ftlv1.DevelService_SyncSchemaServer) error {
	wg, ctx := errgroup.WithContext(stream.Context())

	logger := log.FromContext(ctx)

	// Send initial and subsequent schema updates.
	err := s.sendSchema(stream, s.moduleSchema.Load())
	if err != nil {
		return errors.WithStack(err)
	}
	changes := s.moduleSchema.Subscribe(make(chan schema.Module, 64))
	defer s.moduleSchema.Unsubscribe(changes)
	wg.Go(func() error {
		for module := range changes {
			if err := s.sendSchema(stream, module); err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	})

	// Receive schema updates.
	wg.Go(func() error {
		for {
			received, err := stream.Recv()
			if err != nil {
				return errors.WithStack(err)
			}
			module, err := schema.ParseModuleString(received.Module, received.Schema)
			if err != nil {
				return errors.WithStack(err)
			}

			logger.Debugf("Received schema update from %s", module.Name)

			fullSchema := s.fullSchema.Load()
			fullSchema.Upsert(module)
			s.fullSchema.Store(fullSchema)
		}
	})

	return errors.WithStack(wg.Wait())
}

func (s *Server) List(ctx context.Context, req *ftlv1.ListRequest) (*ftlv1.ListResponse, error) {
	if !metadata.IsDirectRouted(ctx) {
		return s.router.List(ctx, req)
	}
	out := &ftlv1.ListResponse{}
	module := s.moduleSchema.Load()
	for _, verb := range module.Decls {
		if verb, ok := verb.(schema.Verb); ok {
			out.Verbs = append(out.Verbs, module.Name+"."+verb.Name)
		}
	}
	return out, nil
}

func (s *Server) Call(ctx context.Context, req *ftlv1.CallRequest) (*ftlv1.CallResponse, error) {
	if metadata.IsDirectRouted(ctx) || strings.HasPrefix(req.Verb, s.module) {
		return s.plugin.Load().Client.Call(ctx, req)
	}
	return s.router.Call(ctx, req)
}

func (s *Server) FileChange(ctx context.Context, req *ftlv1.FileChangeRequest) (*ftlv1.FileChangeResponse, error) {
	_, err := s.rebuild(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	err = s.plugin.Load().Cmd.Kill(syscall.SIGHUP)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &ftlv1.FileChangeResponse{}, errors.WithStack(err)
}

func (*Server) Ping(context.Context, *ftlv1.PingRequest) (*ftlv1.PingResponse, error) {
	return &ftlv1.PingResponse{}, nil
}

// Restart the FTL hot reload module if it terminates unexpectedly.
func (s *Server) restartModuleOnExit(ctx, cmdCtx context.Context) {
	logger := log.FromContext(ctx)
	for {
		select {
		case <-ctx.Done():
			_ = s.plugin.Load().Cmd.Kill(syscall.SIGTERM)
			return

		case <-cmdCtx.Done():
			err := cmdCtx.Err()
			logger.Warnf("Module exited, restarting: %s", err)

			exe, err := s.rebuild(ctx)
			if err != nil {
				logger.Errorf(err, "Failed to rebuild FTL module")
				return
			}
			var nextPlugin *plugin.Plugin[ftlv1.VerbServiceClient]
			nextPlugin, cmdCtx, err = plugin.Spawn(ctx, s.Config.Module, s.Config.Dir, exe, ftlv1.NewVerbServiceClient)
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
	err = s.writeGoLayout()
	if err != nil {
		return "", errors.WithStack(err)
	}
	err = execInRoot(ctx, s.WorkingDir, "go", "mod", "tidy")
	if err != nil {
		return "", errors.WithStack(err)
	}

	logger.Infof("Extracting schema")
	module, err := sdkgo.ExtractModule(s.Dir)
	if err != nil {
		return "", errors.WithStack(err)
	}
	s.moduleSchema.Store(module)

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

func (s *Server) writeModule(module schema.Module) error {
	err := os.MkdirAll(filepath.Join(s.WorkingDir, "_modules", module.Name), 0750)
	if err != nil {
		return errors.Wrapf(err, "%s: failed to create module directory", module.Name)
	}
	w, err := os.Create(filepath.Join(s.WorkingDir, "_modules", module.Name, "schema.go"))
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

func (s *Server) writeGoMod() error {
	goMod, err := os.Create(filepath.Join(s.WorkingDir, "go.mod"))
	if err != nil {
		return errors.WithStack(err)
	}
	defer goMod.Close() //nolint:gosec
	fmt.Fprintf(goMod, "module main\n")
	fmt.Fprintf(goMod, "require %s v0.0.0\n", s.goModule)
	fmt.Fprintf(goMod, "replace %s => %s\n", s.goModule, s.Dir)
	if s.FTLSource != "" {
		fmt.Fprintf(goMod, "require github.com/TBD54566975/ftl v0.0.0\n")
		fmt.Fprintf(goMod, "replace github.com/TBD54566975/ftl => %s\n", s.FTLSource)
	}
	return nil
}

func (s *Server) sendSchema(stream ftlv1.DevelService_SyncSchemaServer, module schema.Module) error {
	err := stream.Send(&ftlv1.SyncSchemaResponse{
		Module: module.Name,
		Schema: module.String(),
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
	w.Import("os")
	w.Import("context")
	w.Import("github.com/TBD54566975/ftl/drive-go")
	w.Import("github.com/TBD54566975/ftl/sdk-go")
	w.Import("github.com/TBD54566975/ftl/common/plugin")
	w.Import("github.com/TBD54566975/ftl/common/socket")
	w.Import(`github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1`)

	w.L(`func main() {`)
	w.In(func(w *codewriter.Writer) {
		w.L(`ctx, err := sdkgo.ContextWithClient(context.Background(), socket.MustParse(os.Getenv("FTL_ENDPOINT")))`)
		w.L(`if err != nil { panic(err) }`)
		w.L(`plugin.Start(ctx, %q, drivego.NewUserVerbServer(`, config.Module)
		for pkg, endpoints := range endpoints {
			pkgImp := w.Import(pkg)
			for _, endpoint := range endpoints {
				w.L(`  drivego.Handle(%s.%s),`, pkgImp, endpoint.fn.Name())
			}
		}
		w.L(`), ftlv1.RegisterVerbServiceServer)`)
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

func findGoModule(file string) (string, error) {
	goModContent, err := os.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("failed to read go.mod file: %w", err)
	}
	goModFile := modfile.ModulePath(goModContent)
	if goModFile == "" {
		return "", fmt.Errorf("failed to extract Go module from go.mod file: %w", err)
	}
	return goModFile, nil
}

func execInRoot(ctx context.Context, root string, command string, args ...string) error {
	cmd := exec.Command(ctx, root, command, args...)
	return errors.WithStack(cmd.Run())
}
