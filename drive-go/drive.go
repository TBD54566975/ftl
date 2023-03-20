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
	"golang.org/x/tools/go/packages"

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
		Config:   config,
		router:   router,
		module:   config.Module,
		goModule: goModule,
	}

	logger.Info("Starting FTL.module")

	// Build and start the sub-process.
	exe, err := s.rebuild(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	plugin, cmdCtx, err := plugin.Spawn(ctx, s.Config.Dir, exe, ftlv1.NewVerbServiceClient,
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
	router       ftlv1.VerbServiceClient
	module       string
	schema       atomic.Value[schema.Schema]
	goModule     string
	handlers     []Handler
	plugin       atomic.Value[*plugin.Plugin[ftlv1.VerbServiceClient]]
	develService ftlv1.DevelServiceClient

	lastRebuildMu sync.Mutex
	lastRebuild   time.Time
}

func (s *Server) Schema(context.Context, *ftlv1.SchemaRequest) (*ftlv1.SchemaResponse, error) {
	return &ftlv1.SchemaResponse{Schema: s.schema.Load().String()}, nil
}

func (s *Server) List(ctx context.Context, req *ftlv1.ListRequest) (*ftlv1.ListResponse, error) {
	if !metadata.IsDirectRouted(ctx) {
		return s.router.List(ctx, req)
	}
	out := &ftlv1.ListResponse{}
	for _, module := range s.schema.Load().Modules {
		for _, verb := range module.Decls {
			if verb, ok := verb.(schema.Verb); ok {
				out.Verbs = append(out.Verbs, module.Name+"."+verb.Name)
			}
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
			logger.Warn("FTL module exited, restarting", "err", err)

			exe, err := s.rebuild(ctx)
			if err != nil {
				logger.Error("Failed to rebuild FTL module", err)
				return
			}
			var nextPlugin *plugin.Plugin[ftlv1.VerbServiceClient]
			nextPlugin, cmdCtx, err = plugin.Spawn(ctx, s.Config.Dir, exe, ftlv1.NewVerbServiceClient)
			if err != nil {
				logger.Error("Failed to restart FTL module", err)
				continue
			}
			s.plugin.Store(nextPlugin)
		}
	}
}

func (s *Server) rebuild(ctx context.Context) (exe string, err error) {
	s.lastRebuildMu.Lock()
	defer s.lastRebuildMu.Unlock()

	exe = filepath.Join(s.WorkingDir, "ftl-module")

	if time.Now().Sub(s.lastRebuild) < time.Millisecond*250 {
		return exe, nil
	}
	defer func() { s.lastRebuild = time.Now() }()

	logger := log.FromContext(ctx)
	err = writeMain(s.Dir, s.WorkingDir, s.goModule, s.Config)
	if err != nil {
		return "", errors.WithStack(err)
	}
	err = execInRoot(ctx, s.WorkingDir, "go", "mod", "tidy")
	if err != nil {
		return "", errors.WithStack(err)
	}

	logger.Info("Extracting schema")
	module, err := sdkgo.ExtractModule(s.Dir)
	if err != nil {
		return "", errors.WithStack(err)
	}
	s.schema.Store(schema.Schema{Modules: []schema.Module{module}})

	logger.Info("Compiling FTL.module...")
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

func execInRoot(ctx context.Context, root string, command string, args ...string) error {
	cmd := exec.Command(ctx, root, command, args...)
	return errors.WithStack(cmd.Run())
}

func writeMain(root, workDir, module string, config Config) error {
	w, err := generate(config)
	if err != nil {
		return errors.WithStack(err)
	}

	main, err := os.Create(filepath.Join(workDir, "main.go"))
	if err != nil {
		return errors.WithStack(err)
	}
	defer main.Close() //nolint:gosec
	_, err = main.WriteString(w.String())
	if err != nil {
		return errors.WithStack(err)
	}

	goMod, err := os.Create(filepath.Join(workDir, "go.mod"))
	if err != nil {
		return errors.WithStack(err)
	}
	defer goMod.Close() //nolint:gosec
	fmt.Fprintf(goMod, "module main\n")
	fmt.Fprintf(goMod, "require %s v0.0.0\n", module)
	fmt.Fprintf(goMod, "replace %s => %s\n", module, root)
	if config.FTLSource != "" {
		fmt.Fprintf(goMod, "require github.com/TBD54566975/ftl v0.0.0\n")
		fmt.Fprintf(goMod, "replace github.com/TBD54566975/ftl => %s\n", config.FTLSource)
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
		w.L(`ctx = sdkgo.ContextWithModule(ctx, os.Getenv("FTL_MODULE"))`)
		w.L(`plugin.Start(ctx, drivego.NewUserVerbServer(`)
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
