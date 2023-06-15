package compile

import (
	"context"
	_ "embed"
	"os"
	"path"
	"path/filepath"

	"github.com/alecthomas/errors"
	"github.com/oklog/ulid/v2"

	"github.com/TBD54566975/ftl/common/model"
	"github.com/TBD54566975/ftl/common/sha256"
	"github.com/TBD54566975/ftl/go-runtime/compile/generate"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/schema"
)

const ftlModuleSourceBase = "github.com/TBD54566975/ftl/examples"

type Config struct {
	FTLSource string   `short:"S" help:"Path to FTL source code."`
	Dir       string   `short:"d" help:"Path to root directory of module." type:"existingdir" required:""`
	Modules   []string `short:"m" help:"External module paths to include." type:"existingdir"`
}

// Compile a Go FTL module into a deployable executable.
//
// "external" is a list of external modules to generate stubs for in the build
func Compile(ctx context.Context, config Config) (*model.Deployment, error) {
	logger := log.FromContext(ctx)
	logger.Infof("Compiling %s", config.Dir)
	buildDir, err := ensureBuildDir(config.Dir)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	logger.Debugf("Build directory: %s", buildDir)

	logger.Infof("Gathering schema")
	allSchema, goModConfig, err := gatherSchema(config, buildDir)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	mainModuleSchema := allSchema.Modules[0]

	logger.Infof("Generating stubs")
	err = generateBuildContext(buildDir, goModConfig, allSchema, config)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	logger.Infof("Tidying up")
	err = exec.Command(ctx, buildDir, "go", "mod", "tidy").Run()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	dest := filepath.Join(buildDir, "main")

	logger.Infof("Compiling")
	err = exec.Command(ctx, buildDir, "go", "build", "-o", dest).Run()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	digest, err := sha256.SumFile(dest)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	r, err := os.Open(dest)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &model.Deployment{
		Key:    ulid.Make(),
		Schema: mainModuleSchema,
		Module: mainModuleSchema.Name,
		Artefacts: []*model.Artefact{
			{Path: "main", Executable: true, Digest: digest, Content: r},
		},
	}, nil
}

func generateBuildContext(buildDir string, goModConfig generate.GoModConfig, allSchema *schema.Schema, config Config) error {
	mainModuleSchema := allSchema.Modules[0]
	err := generate.File(filepath.Join(buildDir, "main.go"), generate.GenerateMain, mainModuleSchema)
	if err != nil {
		return errors.WithStack(err)
	}
	err = generate.File(filepath.Join(buildDir, "go.mod"), generate.GenerateGoMod, goModConfig)
	if err != nil {
		return errors.WithStack(err)
	}

	for _, module := range allSchema.Modules[1:] {
		moduleDir := filepath.Join(buildDir, "_modules", module.Name)
		err = generate.File(filepath.Join(moduleDir, "main.go"), generate.GenerateExternalModule, module)
		if err != nil {
			return errors.WithStack(err)
		}
		err = generate.File(filepath.Join(moduleDir, "go.mod"), generate.GenerateGoMod, generate.GoModConfig{
			FTLSource: config.FTLSource,
			Replace: map[string]string{
				path.Join(ftlModuleSourceBase, mainModuleSchema.Name): config.Dir,
			},
		})
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func gatherSchema(config Config, buildDir string) (*schema.Schema, generate.GoModConfig, error) {
	mainModuleSchema, err := ExtractModuleSchema(config.Dir)
	if err != nil {
		return nil, generate.GoModConfig{}, errors.WithStack(err)
	}

	goModConfig := generate.GoModConfig{
		FTLSource: config.FTLSource,
		Replace: map[string]string{
			path.Join(ftlModuleSourceBase, mainModuleSchema.Name): config.Dir,
		},
	}
	allSchema := &schema.Schema{
		Modules: []*schema.Module{mainModuleSchema},
	}
	for _, module := range config.Modules {
		moduleSchema, err := ExtractModuleSchema(module)
		if err != nil {
			return nil, generate.GoModConfig{}, errors.WithStack(err)
		}
		allSchema.Modules = append(allSchema.Modules, moduleSchema)
		goModConfig.Replace[path.Join(ftlModuleSourceBase, moduleSchema.Name)] = filepath.Join(buildDir, "_modules", moduleSchema.Name)
	}
	if err := schema.Validate(allSchema); err != nil {
		return nil, generate.GoModConfig{}, errors.WithStack(err)
	}
	return allSchema, goModConfig, nil
}

func ensureBuildDir(dir string) (string, error) {
	fullDir, err := filepath.Abs(dir)
	if err != nil {
		return "", errors.WithStack(err)
	}
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", errors.WithStack(err)
	}
	cacheDir := filepath.Join(userCacheDir, "ftl-go-runtime-compile", sha256.Sum([]byte(fullDir)).String())
	err = os.MkdirAll(cacheDir, 0o750)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return cacheDir, nil
}
