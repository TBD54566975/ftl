package drivego

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/alecthomas/errors"
	"github.com/fsnotify/fsnotify"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"

	"github.com/TBD54566975/ftl/common/log"
	"github.com/TBD54566975/ftl/drive-go/codewriter"
)

var fset = token.NewFileSet()

type Config struct {
	Live      bool   `negatable:"" default:"true" help:"Enable live reloading."`
	FTLSource string `type:"existingdir" help:"Path to FTL source code when developing locally."`
	Dir       string `arg:"" type:"existingdir" help:"Directory of FTL functions."`
}

// Serve a directory of FTL functions.
func Serve(ctx context.Context, config Config) error {
	goModFile, err := findGoModuleFile(config.Dir)
	if err != nil {
		return errors.WithStack(err)
	}
	module, err := findGoModule(goModFile)
	if err != nil {
		return errors.WithStack(err)
	}

	root := filepath.Dir(goModFile)

	scratchDir := filepath.Join(root, ".ftl-drive-go")
	err = os.MkdirAll(scratchDir, 0750)
	if err != nil {
		return errors.WithStack(err)
	}

	err = writeMain(root, scratchDir, module, config.FTLSource, config)
	if err != nil {
		return errors.WithStack(err)
	}
	err = execInRoot(scratchDir, "go", "mod", "tidy")
	if err != nil {
		return errors.Wrap(err, "go mod tidy")
	}

	if config.Live {
		err = watchLoop(ctx, scratchDir, root)
		if err != nil {
			return errors.Wrap(err, "watch loop")
		}
	} else {
		err = execInRoot(scratchDir, "go", "run", ".")
		if err != nil {
			return errors.Wrap(err, "go run")
		}
	}
	return nil
}

func watchLoop(ctx context.Context, scratchDir, root string) error {
	logger := log.FromContext(ctx)
	logger.Info("Live reloading")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.WithStack(err)
	}
	err = watcher.Add(root)
	if err != nil {
		return errors.WithStack(err)
	}

	for {
		ctx, cancel := context.WithCancel(ctx) //nolint:govet
		logger.Info("Restarting FTL.drive-go...")
		cmd := exec.CommandContext(ctx, "go", "run", ".")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = scratchDir
		err = cmd.Start()
		if err != nil {
			cancel()
			return errors.WithStack(err)
		}

	skip:
		select {
		case <-ctx.Done():
			cancel()
			_ = cmd.Wait()
			return nil

		case ev := <-watcher.Events:
			if ev.Op == fsnotify.Chmod {
				goto skip
			}
			cancel()
			_ = cmd.Wait()
		}
	}
}

func execInRoot(root string, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return errors.WithStack(cmd.Run())
}

func writeMain(root, scratchDir, module, ftlSource string, config Config) error {
	w, err := generate(config)
	if err != nil {
		return errors.WithStack(err)
	}

	main, err := os.Create(filepath.Join(scratchDir, "main.go"))
	if err != nil {
		return errors.WithStack(err)
	}
	defer main.Close() //nolint:gosec
	_, err = main.WriteString(w.String())
	if err != nil {
		return errors.WithStack(err)
	}

	goMod, err := os.Create(filepath.Join(scratchDir, "go.mod"))
	if err != nil {
		return errors.WithStack(err)
	}
	defer goMod.Close() //nolint:gosec
	fmt.Fprintf(goMod, "module main\n")
	fmt.Fprintf(goMod, "require %s v0.0.0\n", module)
	fmt.Fprintf(goMod, "replace %s => %s\n", module, root)
	modules := map[string]string{module: root}
	if ftlSource != "" {
		fmt.Fprintf(goMod, "require github.com/TBD54566975/ftl v0.0.0\n")
		fmt.Fprintf(goMod, "replace github.com/TBD54566975/ftl => %s\n", ftlSource)
		modules["github.com/TBD54566975/ftl"] = ftlSource
	}
	return nil
}

func generate(config Config) (*codewriter.Writer, error) {
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
	w.Import("net/http")
	w.Import("github.com/TBD54566975/ftl/sdk-go")
	w.Import("github.com/TBD54566975/ftl/common/log")

	w.L(`func main() {`)
	w.In(func(w *codewriter.Writer) {
		w.L(`logger := log.New(log.Config{}, os.Stderr)`)
		w.L(`ctx := log.ContextWithLogger(context.Background(), logger)`)
		w.L(`mux := http.NewServeMux()`)
		w.L(`logger.Info("Starting FTL server on 127.0.0.1:8080")`)
		for pkg, endpoints := range endpoints {
			pkgImp := w.Import(pkg)
			for _, endpoint := range endpoints {
				w.L(`logger.Info("  Registering endpoint /%s.%s")`, pkg, endpoint.fn.Name())
				w.L(`mux.Handle("/%s.%s", sdkgo.Handler(%s.%s))`, pkg, endpoint.fn.Name(), pkgImp, endpoint.fn.Name())
			}
		}
		w.L(`mux.Handle("/", http.NotFoundHandler())`)
		w.L(`sdkgo.Serve(ctx, mux)`)
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
