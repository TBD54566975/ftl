package languageplugin

import (
	"archive/zip"
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/TBD54566975/scaffolder"
	"github.com/alecthomas/kong"
	"github.com/alecthomas/types/optional"
	"github.com/beevik/etree"
	"github.com/go-viper/mapstructure/v2"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"

	"github.com/TBD54566975/ftl"
	languagepb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/language"
	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/builderrors"
	"github.com/TBD54566975/ftl/internal/errors"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/TBD54566975/ftl/internal/projectconfig"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/watch"
	"github.com/TBD54566975/ftl/jvm-runtime/java"
	"github.com/TBD54566975/ftl/jvm-runtime/kotlin"
)

const JavaBuildToolMaven string = "maven"
const JavaBuildToolGradle string = "gradle"

type JavaConfig struct {
	BuildTool string `mapstructure:"build-tool"`
}

func loadJavaConfig(languageConfig any, language string) (JavaConfig, error) {
	var javaConfig JavaConfig
	err := mapstructure.Decode(languageConfig, &javaConfig)
	if err != nil {
		return JavaConfig{}, fmt.Errorf("failed to decode %s config: %w", language, err)
	}
	return javaConfig, nil
}

// ModuleJavaConfig is language-specific configuration for Java modules.
type javaPlugin struct {
	*internalPlugin
}

var _ = LanguagePlugin(&javaPlugin{})

func newJavaPlugin(ctx context.Context, language string) *javaPlugin {
	internal := newInternalPlugin(ctx, language, buildJava)
	return &javaPlugin{
		internalPlugin: internal,
	}
}

func (p *javaPlugin) ModuleConfigDefaults(ctx context.Context, dir string) (moduleconfig.CustomDefaults, error) {
	defaults := moduleconfig.CustomDefaults{
		GeneratedSchemaDir: optional.Some("src/main/ftl-module-schema"),
		// Watch defaults to files related to maven and gradle
		Watch: []string{"pom.xml", "src/**", "build/generated", "target/generated-sources"},
	}

	pom := filepath.Join(dir, "pom.xml")
	buildGradle := filepath.Join(dir, "build.gradle")
	buildGradleKts := filepath.Join(dir, "build.gradle.kts")
	if fileExists(pom) {
		defaults.LanguageConfig = map[string]any{
			"build-tool": JavaBuildToolMaven,
		}
		defaults.Build = optional.Some("mvn -B package")
		defaults.DeployDir = "target"
	} else if fileExists(buildGradle) || fileExists(buildGradleKts) {
		defaults.LanguageConfig = map[string]any{
			"build-tool": JavaBuildToolGradle,
		}
		defaults.Build = optional.Some("gradle build")
		defaults.DeployDir = "build"
	} else {
		return moduleconfig.CustomDefaults{}, fmt.Errorf("could not find JVM build file in %s", dir)
	}

	return defaults, nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func (p *javaPlugin) GetCreateModuleFlags(ctx context.Context) ([]*kong.Flag, error) {
	return []*kong.Flag{
		{
			Value: &kong.Value{
				Name:       "group",
				Help:       "The Maven groupId of the project.",
				Tag:        &kong.Tag{},
				HasDefault: true,
				Default:    "com.example",
			},
		},
	}, nil
}

func (p *javaPlugin) CreateModule(ctx context.Context, projConfig projectconfig.Config, c moduleconfig.ModuleConfig, flags map[string]string) error {
	logger := log.FromContext(ctx)
	config := c.Abs()
	group, ok := flags["group"]
	if !ok {
		return fmt.Errorf("group flag not set")
	}

	var source *zip.Reader
	if config.Language == "java" {
		source = java.Files()
	} else if config.Language == "kotlin" {
		source = kotlin.Files()
	} else {
		return fmt.Errorf("unknown jvm language %q", config.Language)
	}

	packageDir := strings.ReplaceAll(group, ".", "/")

	sctx := struct {
		Dir        string
		Name       string
		Group      string
		PackageDir string
	}{
		Dir:        config.Dir,
		Name:       config.Module,
		Group:      group,
		PackageDir: packageDir,
	}

	opts := []scaffolder.Option{scaffolder.Exclude("^go.mod$")}
	if !projConfig.Hermit {
		logger.Debugf("Excluding bin directory")
		opts = append(opts, scaffolder.Exclude("^bin"))
	}

	// scaffold at one directory above the module directory
	parentPath := filepath.Dir(config.Dir)
	if err := internal.ScaffoldZip(source, parentPath, sctx, opts...); err != nil {
		return fmt.Errorf("failed to scaffold: %w", err)
	}
	return nil
}

func (p *javaPlugin) GetDependencies(ctx context.Context, config moduleconfig.ModuleConfig) ([]string, error) {
	return p.internalPlugin.getDependencies(ctx, func() ([]string, error) {
		dependencies := map[string]bool{}
		// We also attempt to look at kotlin files
		// As the Java module supports both
		kotin, kotlinErr := extractKotlinFTLImports(config.Module, config.Dir)
		if kotlinErr == nil {
			// We don't really care about the error case, its probably a Java project
			for _, imp := range kotin {
				dependencies[imp] = true
			}
		}
		javaImportRegex := regexp.MustCompile(`^import ftl\.([A-Za-z0-9_.]+)`)

		err := filepath.WalkDir(filepath.Join(config.Dir, "src/main/java"), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return fmt.Errorf("failed to walk directory: %w", err)
			}
			if d.IsDir() || !(strings.HasSuffix(path, ".java")) {
				return nil
			}
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				matches := javaImportRegex.FindStringSubmatch(scanner.Text())
				if len(matches) > 1 {
					module := strings.Split(matches[1], ".")[0]
					if module == config.Module {
						continue
					}
					dependencies[module] = true
				}
			}
			return scanner.Err()
		})

		// We only error out if they both failed
		if err != nil && kotlinErr != nil {
			return nil, fmt.Errorf("%s: failed to extract dependencies from Java module: %w", config.Module, err)
		}
		modules := maps.Keys(dependencies)
		sort.Strings(modules)
		return modules, nil
	})
}

func extractKotlinFTLImports(self, dir string) ([]string, error) {
	dependencies := map[string]bool{}
	kotlinImportRegex := regexp.MustCompile(`^import ftl\.([A-Za-z0-9_.]+)`)

	err := filepath.WalkDir(filepath.Join(dir, "src/main/kotlin"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !(strings.HasSuffix(path, ".kt") || strings.HasSuffix(path, ".kts")) {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("could not open file while extracting dependencies: %w", err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			matches := kotlinImportRegex.FindStringSubmatch(scanner.Text())
			if len(matches) > 1 {
				module := strings.Split(matches[1], ".")[0]
				if module == self {
					continue
				}
				dependencies[module] = true
			}
		}
		return scanner.Err()
	})

	if err != nil {
		return nil, fmt.Errorf("%s: failed to extract dependencies from Kotlin module: %w", self, err)
	}
	modules := maps.Keys(dependencies)
	sort.Strings(modules)
	return modules, nil
}

func buildJava(ctx context.Context, projectRoot, stubsRoot string, bctx BuildContext, buildEnv []string, devMode bool, transaction watch.ModifyFilesTransaction) (BuildResult, error) {
	config := bctx.Config.Abs()
	logger := log.FromContext(ctx)
	javaConfig, err := loadJavaConfig(config.LanguageConfig, config.Language)
	if err != nil {
		return BuildResult{}, fmt.Errorf("failed to build module %q: %w", config.Module, err)
	}
	if javaConfig.BuildTool == JavaBuildToolMaven {
		if err := setPOMProperties(ctx, config.Dir); err != nil {
			// This is not a critical error, things will probably work fine
			// TBH updating the pom is maybe not the best idea anyway
			logger.Warnf("unable to update ftl.version in %s: %s", config.Dir, err.Error())
		}
	}
	logger.Infof("Using build command '%s'", config.Build)
	command := exec.Command(ctx, log.Debug, config.Dir, "bash", "-c", config.Build)
	err = command.RunBuffered(ctx)
	if err != nil {
		return BuildResult{}, fmt.Errorf("failed to build module %q: %w", config.Module, err)
	}

	buildErrs, err := loadProtoErrors(config)
	if err != nil {
		return BuildResult{}, fmt.Errorf("failed to load build errors: %w", err)
	}
	result := BuildResult{
		Errors: buildErrs,
	}
	if builderrors.ContainsTerminalError(buildErrs) {
		// skip reading schema
		return result, nil
	}

	moduleSchema, err := schema.ModuleFromProtoFile(filepath.Join(config.DeployDir, "schema.pb"))
	if err != nil {
		return BuildResult{}, fmt.Errorf("failed to read schema for module: %w", err)
	}

	result.Schema = moduleSchema
	result.Deploy = []string{"launch", "quarkus-app"}
	return result, nil
}

// setPOMProperties updates the ftl.version properties in the
// pom.xml file in the given base directory.
func setPOMProperties(ctx context.Context, baseDir string) error {
	logger := log.FromContext(ctx)
	ftlVersion := ftl.Version
	if ftlVersion == "dev" {
		ftlVersion = "1.0-SNAPSHOT"
	}

	pomFile := filepath.Clean(filepath.Join(baseDir, "pom.xml"))

	logger.Debugf("Setting ftl.version in %s to %s", pomFile, ftlVersion)

	tree := etree.NewDocument()
	if err := tree.ReadFromFile(pomFile); err != nil {
		return fmt.Errorf("unable to read %s: %w", pomFile, err)
	}
	root := tree.Root()

	parent := root.SelectElement("parent")
	versionSet := false
	if parent != nil {
		// You can't use properties in the parent
		// If they are using our parent then we want to update the version
		group := parent.SelectElement("groupId")
		artifact := parent.SelectElement("artifactId")
		if group.Text() == "xyz.block.ftl" && (artifact.Text() == "ftl-build-parent-java" || artifact.Text() == "ftl-build-parent-kotlin") {
			version := parent.SelectElement("version")
			if version != nil {
				version.SetText(ftlVersion)
				versionSet = true
			}
		}
	}

	err := updatePomProperties(root, pomFile, ftlVersion)
	if err != nil && !versionSet {
		// This is only a failure if we also did not update the parent
		return err
	}

	err = tree.WriteToFile(pomFile)
	if err != nil {
		return fmt.Errorf("unable to write %s: %w", pomFile, err)
	}
	return nil
}

func updatePomProperties(root *etree.Element, pomFile string, ftlVersion string) error {
	properties := root.SelectElement("properties")
	if properties == nil {
		return fmt.Errorf("unable to find <properties> in %s", pomFile)
	}
	version := properties.SelectElement("ftl.version")
	if version == nil {
		return fmt.Errorf("unable to find <properties>/<ftl.version> in %s", pomFile)
	}
	version.SetText(ftlVersion)
	return nil
}

func loadProtoErrors(config moduleconfig.AbsModuleConfig) ([]builderrors.Error, error) {
	errorsPath := filepath.Join(config.DeployDir, "errors.pb")
	if _, err := os.Stat(errorsPath); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	content, err := os.ReadFile(errorsPath)
	if err != nil {
		return nil, fmt.Errorf("could not load build errors file: %w", err)
	}

	errorspb := &languagepb.ErrorList{}
	err = proto.Unmarshal(content, errorspb)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal build errors %w", err)
	}
	return languagepb.ErrorsFromProto(errorspb), nil
}
