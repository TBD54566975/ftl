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
	"syscall"

	"github.com/alecthomas/atomic"
	"github.com/alecthomas/errors"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"

	"github.com/TBD54566975/ftl/common/exec"
	ftlv1 "github.com/TBD54566975/ftl/common/gen/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/common/log"
	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/drive-go/codewriter"
)

type Config struct {
	FTLSource  string `env:"FTL_SOURCE" type:"existingdir" help:"Path to FTL source code when developing locally."`
	WorkingDir string `required:"" type:"existingdir" env:"FTL_WORKING_DIR" help:"Working directory for FTL runtime."`
	Dir        string `required:"" type:"existingdir" env:"FTL_MODULE_ROOT" help:"Directory to root of Go FTL module"`
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
		plugin: atomic.New[*plugin.Plugin[ftlv1.DriveServiceClient]](nil),
	}

	logger.Info("Starting FTL.module")

	// Build and start the sub-process.
	exe, err := d.rebuild(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	d.exe = exe

	plugin, cmdCtx, err := plugin.Spawn(ctx, d.Config.Dir, exe, ftlv1.NewDriveServiceClient)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	d.plugin.Store(plugin)

	go d.restartModuleOnExit(ctx, cmdCtx)
	return d, nil
}

var _ ftlv1.DriveServiceServer = (*driveServer)(nil)

type driveServer struct {
	Config
	exe      string
	module   string
	handlers []Handler
	plugin   *atomic.Value[*plugin.Plugin[ftlv1.DriveServiceClient]]
}

func (d *driveServer) List(ctx context.Context, req *ftlv1.ListRequest) (*ftlv1.ListResponse, error) {
	return d.plugin.Load().Client.List(ctx, req)
}

func (d *driveServer) Call(ctx context.Context, req *ftlv1.CallRequest) (*ftlv1.CallResponse, error) {
	return d.plugin.Load().Client.Call(ctx, req)
}

func (d *driveServer) FileChange(ctx context.Context, req *ftlv1.FileChangeRequest) (*ftlv1.FileChangeResponse, error) {
	err := d.plugin.Load().Cmd.Kill(syscall.SIGHUP)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	_, err = d.rebuild(ctx)
	return &ftlv1.FileChangeResponse{}, errors.WithStack(err)
}

func (*driveServer) Ping(context.Context, *ftlv1.PingRequest) (*ftlv1.PingResponse, error) {
	return &ftlv1.PingResponse{}, nil
}

// Restart the FTL hot reload module if it terminates unexpectedly.
func (d *driveServer) restartModuleOnExit(ctx, cmdCtx context.Context) {
	logger := log.FromContext(ctx)
	for {
		select {
		case <-ctx.Done():
			_ = d.plugin.Load().Cmd.Kill(syscall.SIGTERM)
			return

		case <-cmdCtx.Done():
			err := cmdCtx.Err()
			logger.Warn("FTL module exited, restarting", "err", err)
			var nextPlugin *plugin.Plugin[ftlv1.DriveServiceClient]
			nextPlugin, cmdCtx, err = plugin.Spawn(ctx, d.Config.Dir, d.exe, ftlv1.NewDriveServiceClient)
			if err != nil {
				logger.Error("Failed to restart FTL module", err)
				continue
			}
			d.plugin.Store(nextPlugin)
		}
	}
}

func (d *driveServer) rebuild(ctx context.Context) (exe string, err error) {
	logger := log.FromContext(ctx)
	err = writeMain(d.Dir, d.WorkingDir, d.module, d.Config)
	if err != nil {
		return "", errors.WithStack(err)
	}
	err = execInRoot(ctx, d.WorkingDir, "go", "mod", "tidy")
	if err != nil {
		return "", errors.WithStack(err)
	}

	exe = filepath.Join(d.WorkingDir, "ftl-module")
	logger.Info("Compiling FTL.module...")
	err = execInRoot(ctx, d.WorkingDir, "go", "build", "-trimpath", "-buildvcs=false", "-ldflags=-s -w -buildid=", "-o", exe)
	if err != nil {
		source, merr := ioutil.ReadFile(filepath.Join(d.WorkingDir, "main.go"))
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
	w.Import("github.com/TBD54566975/ftl/drive-go")
	w.Import("github.com/TBD54566975/ftl/common/plugin")
	w.Import(`github.com/TBD54566975/ftl/common/gen/xyz/block/ftl/v1`)

	w.L(`func main() {`)
	w.In(func(w *codewriter.Writer) {
		w.L(`handlers := []drivego.Handler{}`)
		for pkg, endpoints := range endpoints {
			pkgImp := w.Import(pkg)
			for _, endpoint := range endpoints {
				w.L(`handlers = append(handlers, drivego.Handle(%s.%s))`, pkgImp, endpoint.fn.Name())
			}
		}
		w.L(`plugin.Start(drivego.NewUserVerbServer(handlers...), ftlv1.RegisterDriveServiceServer)`)
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
