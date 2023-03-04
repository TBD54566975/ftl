package drivego

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/alecthomas/atomic"
	"github.com/alecthomas/errors"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/TBD54566975/ftl/common/log"
	"github.com/TBD54566975/ftl/drive-go/codewriter"
	"github.com/TBD54566975/ftl/internal/exec"
	ftlv1 "github.com/TBD54566975/ftl/internal/gen/xyz/block/ftl/v1"
)

type Config struct {
	FTLSource  string `env:"FTL_SOURCE" type:"existingdir" help:"Path to FTL source code when developing locally."`
	WorkingDir string `required:"" type:"existingdir" env:"FTL_WORKING_DIR" help:"Working directory for FTL runtime."`
	Dir        string `required:"" type:"existingdir" env:"FTL_DRIVE_ROOT" help:"Directory of Go FTL functions."`
}

// New creates a new DriveService for a directory of Go Verbs.
func New(ctx context.Context, config Config) (ftlv1.DriveServiceServer, error) {
	logger := log.FromContext(ctx)
	goModFile := filepath.Join(config.Dir, "go.mod")
	_, err := os.Stat(goModFile)
	if err != nil {
		return nil, errors.Wrapf(err, "go.mod not found in %s", config.Dir)
	}
	module, err := findGoModule(goModFile)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	d := &driveServer{
		Config: config,
		module: module,
		cmd:    atomic.New[*exec.Cmd](nil),
	}

	logger.Info("Starting FTL.module")

	// Build and start the sub-process.
	d.cmd.Store(d.newModuleCmd(ctx))
	err = d.rebuild(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	finished := make(chan error, 1)
	go func() { finished <- d.cmd.Load().Run() }()

	socket := filepath.Join(d.WorkingDir, "module.sock")

	dialCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(dialCtx, "unix://"+socket,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		panic(err)
	}

	logger.Info("FTL.module online", "dir", d.Dir)

	d.client = ftlv1.NewModuleServiceClient(conn)

	_, err = d.client.Ping(ctx, &ftlv1.PingRequest{})
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}

	go d.restartModuleOnExit(ctx, finished)
	return d, nil
}

type driveServer struct {
	Config
	module string
	client ftlv1.ModuleServiceClient
	cmd    *atomic.Value[*exec.Cmd]
}

func (d *driveServer) Call(ctx context.Context, req *ftlv1.CallRequest) (*ftlv1.CallResponse, error) {
	return d.client.Call(ctx, req)
}

func (d *driveServer) FileChange(ctx context.Context, req *ftlv1.FileChangeRequest) (*ftlv1.FileChangeResponse, error) {
	err := d.cmd.Load().Kill(syscall.SIGHUP)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &ftlv1.FileChangeResponse{}, d.rebuild(ctx)
}

func (*driveServer) Ping(context.Context, *ftlv1.PingRequest) (*ftlv1.PingResponse, error) {
	return &ftlv1.PingResponse{}, nil
}

// Restart the FTL hot reload module if it terminates unexpectedly.
func (d *driveServer) restartModuleOnExit(ctx context.Context, finished chan error) {
	logger := log.FromContext(ctx)
	for {
		select {
		case <-ctx.Done():
			_ = d.cmd.Load().Kill(syscall.SIGTERM)
			return

		case err := <-finished:
			logger.Warn("FTL module exited, restarting in 1s", "err", err)
			time.Sleep(time.Second)
			cmd := d.newModuleCmd(ctx)
			go func() { finished <- d.cmd.Load().Run() }()
			d.cmd.Store(cmd)
		}
	}
}

func (d *driveServer) newModuleCmd(ctx context.Context) *exec.Cmd {
	socket := filepath.Join(d.WorkingDir, "module.sock")
	cmd := exec.Command(ctx, d.WorkingDir, filepath.Join(d.WorkingDir, "ftl-module"))
	cmd.Env = append(cmd.Env, "FTL_MODULE_SOCKET="+socket)
	return cmd
}

func (d *driveServer) rebuild(ctx context.Context) error {
	logger := log.FromContext(ctx)
	err := writeMain(d.Dir, d.WorkingDir, d.module, d.Config)
	if err != nil {
		return errors.WithStack(err)
	}
	logger.Info("Restarting FTL.drive-go...")
	err = execInRoot(ctx, d.WorkingDir, "go", "mod", "tidy")
	if err != nil {
		return errors.WithStack(err)
	}

	err = execInRoot(ctx, d.WorkingDir, "go", "build", "-trimpath", "-buildvcs=false", "-ldflags=-s -w -buildid=", "-o", "ftl-module")
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
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
	w.Import("context")
	w.Import("os")
	w.Import("github.com/TBD54566975/ftl/sdk-go")
	w.Import("github.com/TBD54566975/ftl/common/log")

	w.L(`func main() {`)
	w.In(func(w *codewriter.Writer) {
		w.L(`logger := log.New(log.Config{}, os.Stderr).With("C", "FTL.module")`)
		w.L(`ctx := log.ContextWithLogger(context.Background(), logger)`)
		w.L(`socket := os.Getenv("FTL_MODULE_SOCKET")`)
		w.L(`if socket == "" { panic("FTL_MODULE_SOCKET not set") }`)
		w.L(`logger.Info("Starting FTL server on " + socket)`)
		w.L(`handlers := []sdkgo.Handler{}`)
		for pkg, endpoints := range endpoints {
			pkgImp := w.Import(pkg)
			for _, endpoint := range endpoints {
				w.L(`logger.Info("  Registering endpoint %s.%s")`, pkg, endpoint.fn.Name())
				w.L(`handlers = append(handlers, sdkgo.Handle(%s.%s))`, pkgImp, endpoint.fn.Name())
			}
		}
		w.L(`sdkgo.Serve(ctx, socket, handlers)`)
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

// Find the go.mod file enclosing "dir".
func findGoModuleFile(root string) (string, error) {
	dir := root
	for dir != "/" {
		path := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			break
		}
		dir = filepath.Dir(dir)
	}
	return "", fmt.Errorf("no go.mod file found in %s or any parent directory", root)
}

func typeRef(pkg *packages.Package, t types.Type) (pkgRef, ref string) {
	if named, ok := t.(*types.Named); ok {
		pkgRef = named.Obj().Pkg().Path()
		ref = named.Obj().Name()
		if pkgRef == pkg.PkgPath {
			pkgRef = ""
		} else {
			ref = path.Base(pkgRef) + "." + ref
		}
		return
	}
	return "", t.String()
}
