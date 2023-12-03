package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/kong"
	"github.com/radovskyb/watcher"
	"golang.org/x/mod/modfile"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/TBD54566975/ftl/backend/common/exec"
	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/model"
	"github.com/TBD54566975/ftl/backend/common/rpc"
	"github.com/TBD54566975/ftl/backend/common/sha256"
	"github.com/TBD54566975/ftl/backend/common/slices"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/compile"
	"github.com/TBD54566975/ftl/go-runtime/compile/generate"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

type watchCmd struct{}

func (w *watchCmd) Run(ctx context.Context, c *cli, client ftlv1connect.ControllerServiceClient, bctx BuildContext) error {
	err := buildRemoteModules(ctx, client, bctx)
	if err != nil {
		return err
	}

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error { return pullModules(ctx, client, bctx) })
	wg.Go(func() error { return pushModules(ctx, client, c.WatchFrequency, bctx, c.FTL) })

	if err := wg.Wait(); err != nil {
		return err
	}
	return nil
}

type deployCmd struct {
	Name string `arg:"" required:"" help:"Name of module to deploy."`
}

func (d *deployCmd) Run(ctx context.Context, c *cli, client ftlv1connect.ControllerServiceClient, bctx BuildContext) error {
	return pushModule(ctx, client, c.FTL, filepath.Join(c.Root, d.Name), bctx)
}

type cli struct {
	LogConfig      log.Config    `embed:""`
	FTL            string        `env:"FTL_ENDPOINT" help:"FTL endpoint to connect to." default:"http://localhost:8892"`
	WatchFrequency time.Duration `short:"w" default:"500ms" help:"Frequency to watch for changes to local FTL modules."`
	Root           string        `short:"r" type:"existingdir" help:"Root directory to sync FTL modules into." default:"."`
	OS             string        `short:"o" help:"OS to build for." env:"GOOS"`
	Arch           string        `short:"a" help:"Architecture to build for." env:"GOARCH"`

	Watch  watchCmd  `cmd:"" default:"" help:"Watch for and rebuild local and remote FTL modules."`
	Deploy deployCmd `cmd:"" help:"Deploy a local FTL module."`
}

type BuildContext struct {
	OS   string
	Arch string
	Root string
	ImportRoot
}

func main() {
	c := &cli{}
	kctx := kong.Parse(c)

	client := rpc.Dial(ftlv1connect.NewControllerServiceClient, c.FTL, log.Warn)
	logger := log.Configure(os.Stderr, c.LogConfig)
	ctx := log.ContextWithLogger(context.Background(), logger)

	importRoot, err := findImportRoot(c.Root)
	kctx.FatalIfErrorf(err)

	bctx := BuildContext{
		OS:         c.OS,
		Arch:       c.Arch,
		Root:       c.Root,
		ImportRoot: importRoot,
	}

	kctx.Bind(bctx)
	kctx.BindTo(ctx, (*context.Context)(nil))
	kctx.BindTo(client, (*ftlv1connect.ControllerServiceClient)(nil))
	err = kctx.Run()
	kctx.FatalIfErrorf(err)

}

type ImportRoot struct {
	Module      *modfile.File
	GoModuleDir string
	FTLBasePkg  string
	FTLBaseDir  string
}

// "prefix" is the import prefix for FTL modules.
func findImportRoot(root string) (importRoot ImportRoot, err error) {
	modDir := root
	for {
		if modDir == "/" {
			return ImportRoot{}, fmt.Errorf("no go.mod file found")
		}
		if _, err := os.Stat(filepath.Join(modDir, "go.mod")); err == nil {
			break
		}
		modDir = filepath.Dir(modDir)
	}
	modFile := filepath.Join(modDir, "go.mod")
	data, err := os.ReadFile(modFile)
	if err != nil {
		return ImportRoot{}, fmt.Errorf("%s: %w", "failed to read go.mod", err)
	}
	module, err := modfile.Parse(modFile, data, nil)
	if err != nil {
		return ImportRoot{}, fmt.Errorf("%s: %w", "failed to parse go.mod", err)
	}
	return ImportRoot{
		Module:      module,
		GoModuleDir: modDir,
		FTLBasePkg:  path.Join(module.Module.Mod.Path, strings.TrimPrefix(strings.TrimPrefix(root, modDir), "/")),
		FTLBaseDir:  root,
	}, nil
}

func pushModules(ctx context.Context, client ftlv1connect.ControllerServiceClient, watchFrequency time.Duration, bctx BuildContext, endpoint string) error {
	logger := log.FromContext(ctx)
	entries, err := os.ReadDir(bctx.Root)
	if err != nil {
		return fmt.Errorf("%s: %w", "failed to read root directory", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(bctx.Root, entry.Name())
		if _, err := os.Stat(filepath.Join(dir, "generated_ftl_module.go")); err == nil {
			continue
		}

		logger.Infof("Pushing local FTL module %q", entry.Name())
		err := pushModule(ctx, client, endpoint, dir, bctx)
		if err != nil {
			if connect.CodeOf(err) == connect.CodeAlreadyExists {
				logger.Infof("Module %q already exists, skipping", entry.Name())
				continue
			}
			logger.Warnf("Failed to push module %q, continuing: %s", entry.Name(), err)
		}
	}

	logger.Infof("Watching %s for changes", bctx.Root)
	wg, ctx := errgroup.WithContext(ctx)
	watch := watcher.New()
	defer watch.Close()
	wg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil

			case event := <-watch.Event:
				if event.IsDir() ||
					strings.Contains(event.Path, "/.") ||
					strings.Contains(event.Path, "/generated_ftl_module.go") ||
					!strings.HasPrefix(event.Path, bctx.Root) ||
					strings.Contains(event.Path, "/build/") {
					continue
				}
				dir := strings.TrimPrefix(event.Path, bctx.Root+"/")
				dir = filepath.Join(bctx.Root, strings.Split(dir, "/")[0])
				logger.Infof("Detected change to %s, pushing module", dir)

				err := pushModule(ctx, client, endpoint, dir, bctx)
				if err != nil {
					logger.Errorf(err, "failed to rebuild module")
				}

			case err := <-watch.Error:
				return fmt.Errorf("%s: %w", "watch error", err)
			}
		}
	})
	err = watch.AddRecursive(bctx.Root)
	if err != nil {
		return fmt.Errorf("%s: %w", "failed to watch root directory", err)
	}
	wg.Go(func() error { return watch.Start(watchFrequency) })
	return wg.Wait()
}

func pushModule(ctx context.Context, client ftlv1connect.ControllerServiceClient, endpoint string, dir string, bctx BuildContext) error {
	logger := log.FromContext(ctx)

	sch, err := compile.ExtractModuleSchema(dir)
	if err != nil {
		return fmt.Errorf("failed to extract schema for module %q: %w", dir, err)
	}

	if !hasVerbs(sch) {
		logger.Warnf("No verbs found in module %q, ignored", dir)
		return nil
	}

	tmpDir, err := generateBuildDir(dir, sch, bctx)
	if err != nil {
		return fmt.Errorf("%s: %w", "failed to generate build directory", err)
	}

	logger.Infof("Building module %s in %s", sch.Name, tmpDir)
	cmd := exec.Command(ctx, log.Info, tmpDir, "go", "build", "-o", "main", "-trimpath", "-ldflags=-s -w -buildid=", ".")
	cmd.Env = append(cmd.Environ(), "GOOS="+bctx.OS, "GOARCH="+bctx.Arch, "CGO_ENABLED=0")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", "failed to build module", err)
	}
	dest := filepath.Join(tmpDir, "main")

	logger.Infof("Preparing deployment")
	digest, err := sha256.SumFile(dest)
	if err != nil {
		return err
	}
	r, err := os.Open(dest)
	if err != nil {
		return err
	}
	deployment := &model.Deployment{
		Language: "go",
		Name:     model.NewDeploymentName(sch.Name),
		Schema:   sch,
		Module:   sch.Name,
		Artefacts: []*model.Artefact{
			{Path: "main", Executable: true, Digest: digest, Content: r},
		},
	}
	defer deployment.Close()

	err = uploadArtefacts(ctx, client, deployment)
	if err != nil {
		return fmt.Errorf("%s: %w", "failed to upload artefacts", err)
	}

	err = deploy(ctx, client, deployment, endpoint)
	if err != nil {
		return fmt.Errorf("%s: %w", "failed to deploy", err)
	}
	return nil
}

func deploy(ctx context.Context, client ftlv1connect.ControllerServiceClient, deployment *model.Deployment, endpoint string) error {
	logger := log.FromContext(ctx)
	module := deployment.Schema.ToProto().(*schemapb.Module) //nolint:forcetypeassert
	module.Runtime = &schemapb.ModuleRuntime{
		Language:    deployment.Language,
		CreateTime:  timestamppb.Now(),
		MinReplicas: 1,
	}
	labels, err := structpb.NewStruct(map[string]any{
		"os":        runtime.GOOS,
		"arch":      runtime.GOARCH,
		"languages": []any{"go"},
	})
	if err != nil {
		return fmt.Errorf("%s: %w", "failed to create labels", err)
	}
	cdResp, err := client.CreateDeployment(ctx, connect.NewRequest(&ftlv1.CreateDeploymentRequest{
		Schema: module,
		Labels: labels,
		Artefacts: slices.Map(deployment.Artefacts, func(t *model.Artefact) *ftlv1.DeploymentArtefact {
			return &ftlv1.DeploymentArtefact{
				Digest:     t.Digest.String(),
				Path:       t.Path,
				Executable: t.Executable,
			}
		}),
	}))
	if err != nil {
		return fmt.Errorf("%s: %w", "failed to create deployment", err)
	}
	logger.Infof("Created deployment %s (%s/deployments/%s)", cdResp.Msg.DeploymentName, endpoint, cdResp.Msg.DeploymentName)
	_, err = client.ReplaceDeploy(ctx, connect.NewRequest(&ftlv1.ReplaceDeployRequest{
		DeploymentName: cdResp.Msg.DeploymentName,
		MinReplicas:    1,
	}))
	if err != nil {
		return fmt.Errorf("failed to deploy %q: %w", cdResp.Msg.DeploymentName, err)
	}
	return nil
}

func uploadArtefacts(ctx context.Context, client ftlv1connect.ControllerServiceClient, deployment *model.Deployment) error {
	logger := log.FromContext(ctx)
	digests := slices.Map(deployment.Artefacts, func(t *model.Artefact) string { return t.Digest.String() })
	gadResp, err := client.GetArtefactDiffs(ctx, connect.NewRequest(&ftlv1.GetArtefactDiffsRequest{ClientDigests: digests}))
	if err != nil {
		return err
	}
	artefactsToUpload := slices.Filter(deployment.Artefacts, func(t *model.Artefact) bool {
		for _, missing := range gadResp.Msg.MissingDigests {
			if t.Digest.String() == missing {
				return true
			}
		}
		return false
	})
	for _, artefact := range artefactsToUpload {
		content, err := io.ReadAll(artefact.Content)
		if err != nil {
			return fmt.Errorf("failed to read artefact %q: %w", artefact.Path, err)
		}
		_, err = client.UploadArtefact(ctx, connect.NewRequest(&ftlv1.UploadArtefactRequest{Content: content}))
		if err != nil {
			return fmt.Errorf("failed to upload artefact %q: %w", artefact.Path, err)
		}
		logger.Infof("Uploaded %s:%s", artefact.Digest, artefact.Path)
	}
	return nil
}

func generateBuildDir(dir string, sch *schema.Module, bctx BuildContext) (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("%s: %w", "failed to get user cache directory", err)
	}
	dirHash := sha256.Sum([]byte(dir))
	tmpDir := filepath.Join(cacheDir, "ftl-go", "build", fmt.Sprintf("%s-%s", sch.Name, dirHash))
	if err := os.MkdirAll(tmpDir, 0750); err != nil {
		return "", fmt.Errorf("%s: %w", "failed to create build directory", err)
	}
	mainFile := filepath.Join(tmpDir, "main.go")
	if err := generate.File(mainFile, bctx.FTLBasePkg, generate.Main, sch); err != nil {
		return "", fmt.Errorf("%s: %w", "failed to generate main.go", err)
	}
	goWorkFile := filepath.Join(tmpDir, "go.work")
	if err := generate.File(goWorkFile, bctx.FTLBasePkg, generate.GenerateGoWork, []string{
		bctx.GoModuleDir,
	}); err != nil {
		return "", fmt.Errorf("%s: %w", "failed to generate go.work", err)
	}
	goModFile := filepath.Join(tmpDir, "go.mod")
	replace := map[string]string{
		bctx.Module.Module.Mod.Path: bctx.GoModuleDir,
	}
	if err := generate.File(goModFile, bctx.FTLBasePkg, generate.GenerateGoMod, generate.GoModConfig{
		Replace: replace,
	}); err != nil {
		return "", fmt.Errorf("%s: %w", "failed to generate go.mod", err)
	}
	return tmpDir, nil
}

func hasVerbs(sch *schema.Module) bool {
	for _, decl := range sch.Decls {
		if _, ok := decl.(*schema.Verb); ok {
			return true
		}
	}
	return false
}

func pullModules(ctx context.Context, client ftlv1connect.ControllerServiceClient, bctx BuildContext) error {
	resp, err := client.PullSchema(ctx, connect.NewRequest(&ftlv1.PullSchemaRequest{}))
	if err != nil {
		return fmt.Errorf("%s: %w", "failed to pull schema", err)
	}
	for resp.Receive() {
		msg := resp.Msg()
		err = generateModuleFromSchema(ctx, msg.Schema, bctx)
		if err != nil {
			return fmt.Errorf("%s: %w", "failed to sync module", err)
		}
	}
	if err := resp.Err(); err != nil {
		return fmt.Errorf("%s: %w", "failed to pull schema", err)
	}
	return nil
}

func buildRemoteModules(ctx context.Context, client ftlv1connect.ControllerServiceClient, bctx BuildContext) error {
	fullSchema, err := client.GetSchema(ctx, connect.NewRequest(&ftlv1.GetSchemaRequest{}))
	if err != nil {
		return fmt.Errorf("%s: %w", "failed to retrieve schema", err)
	}
	for _, module := range fullSchema.Msg.Schema.Modules {
		err := generateModuleFromSchema(ctx, module, bctx)
		if err != nil {
			return fmt.Errorf("%s: %w", "failed to generate module", err)
		}
	}
	return err
}

func generateModuleFromSchema(ctx context.Context, msg *schemapb.Module, bctx BuildContext) error {
	sch, err := schema.ModuleFromProto(msg)
	if err != nil {
		return fmt.Errorf("%s: %w", "failed to parse schema", err)
	}
	dir := filepath.Join(bctx.Root, sch.Name)
	if _, err := os.Stat(dir); err == nil {
		if _, err = os.Stat(filepath.Join(dir, "generated_ftl_module.go")); errors.Is(err, os.ErrNotExist) {
			return nil
		}
	}
	if err := generateModule(ctx, dir, sch, bctx); err != nil {
		return fmt.Errorf("%s: %w", "failed to generate module", err)
	}
	return nil
}

func generateModule(ctx context.Context, dir string, sch *schema.Module, bctx BuildContext) error {
	logger := log.FromContext(ctx)
	logger.Infof("Generating stubs for FTL module %s", sch.Name)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("%s: %w", "failed to create module directory", err)
	}
	w, err := os.Create(filepath.Join(dir, "generated_ftl_module.go~"))
	if err != nil {
		return fmt.Errorf("%s: %w", "failed to create stub file", err)
	}
	defer w.Close() //nolint:gosec
	defer os.Remove(w.Name())
	if err := generate.ExternalModule(w, sch, bctx.FTLBasePkg); err != nil {
		return fmt.Errorf("%s: %w", "failed to generate stubs", err)
	}
	if err := os.Rename(w.Name(), strings.TrimRight(w.Name(), "~")); err != nil {
		return fmt.Errorf("%s: %w", "failed to rename stub file", err)
	}
	return nil
}
