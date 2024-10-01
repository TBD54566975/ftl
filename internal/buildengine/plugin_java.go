package buildengine

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
	"github.com/beevik/etree"
	"golang.org/x/exp/maps"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/TBD54566975/ftl/jvm-runtime/java"
	"github.com/TBD54566975/ftl/jvm-runtime/kotlin"
)

type javaPlugin struct {
	*internalPlugin
}

var _ = LanguagePlugin(&javaPlugin{})

func newJavaPlugin(ctx context.Context, config moduleconfig.ModuleConfig) *javaPlugin {
	internal := newInternalPlugin(ctx, config, buildJava)
	return &javaPlugin{
		internalPlugin: internal,
	}
}

func (p *javaPlugin) CreateModule(ctx context.Context, config moduleconfig.ModuleConfig, includeBinDir bool, replacements map[string]string, group string) error {
	logger := log.FromContext(ctx)

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

	opts := []scaffolder.Option{scaffolder.Functions(scaffoldFuncs), scaffolder.Exclude("^go.mod$")}
	if !includeBinDir {
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

func (p *javaPlugin) GetDependencies(ctx context.Context) ([]string, error) {
	return p.internalPlugin.getDependencies(ctx, func() ([]string, error) {
		dependencies := map[string]bool{}
		// We also attempt to look at kotlin files
		// As the Java module supports both
		kotin, kotlinErr := extractKotlinFTLImports(p.config.Module, p.config.Dir)
		if kotlinErr == nil {
			// We don't really care about the error case, its probably a Java project
			for _, imp := range kotin {
				dependencies[imp] = true
			}
		}
		javaImportRegex := regexp.MustCompile(`^import ftl\.([A-Za-z0-9_.]+)`)

		err := filepath.WalkDir(filepath.Join(p.config.Dir, "src/main/java"), func(path string, d fs.DirEntry, err error) error {
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
					if module == p.config.Module {
						continue
					}
					dependencies[module] = true
				}
			}
			return scanner.Err()
		})

		// We only error out if they both failed
		if err != nil && kotlinErr != nil {
			return nil, fmt.Errorf("%s: failed to extract dependencies from Java module: %w", p.config.Module, err)
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

func buildJava(ctx context.Context, projectRoot string, config moduleconfig.AbsModuleConfig, sch *schema.Schema, buildEnv []string, devMode bool, transaction ModifyFilesTransaction) error {
	logger := log.FromContext(ctx)
	if config.Java.BuildTool == moduleconfig.JavaBuildToolMaven {
		if err := setPOMProperties(ctx, config.Dir); err != nil {
			// This is not a critical error, things will probably work fine
			// TBH updating the pom is maybe not the best idea anyway
			logger.Warnf("unable to update ftl.version in %s: %s", config.Dir, err.Error())
		}
	}
	logger.Infof("Using build command '%s'", config.Build)
	command := exec.Command(ctx, log.Debug, config.Dir, "bash", "-c", config.Build)
	err := command.RunBuffered(ctx)
	if err != nil {
		return fmt.Errorf("failed to build module %q: %w", config.Module, err)
	}
	return nil
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
