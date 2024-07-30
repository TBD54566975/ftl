package buildengine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/TBD54566975/scaffolder"
	"github.com/beevik/etree"
	sets "github.com/deckarep/golang-set/v2"
	"golang.org/x/exp/maps"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	kotlinruntime "github.com/TBD54566975/ftl/kotlin-runtime"
)

type externalModuleContext struct {
	module Module
	*schema.Schema
}

func (e externalModuleContext) ExternalModules() []*schema.Module {
	modules := make([]*schema.Module, 0, len(e.Modules))
	for _, module := range e.Modules {
		if module.Name == e.module.Config.Module {
			continue
		}
		modules = append(modules, module)
	}
	return modules
}

func buildKotlinModule(ctx context.Context, sch *schema.Schema, module Module) error {
	logger := log.FromContext(ctx)
	if err := SetPOMProperties(ctx, module.Config.Dir); err != nil {
		return fmt.Errorf("unable to update ftl.version in %s: %w", module.Config.Dir, err)
	}
	if err := generateExternalModules(ctx, module, sch); err != nil {
		return fmt.Errorf("unable to generate external modules for %s: %w", module.Config.Module, err)
	}
	if err := prepareFTLRoot(module); err != nil {
		return fmt.Errorf("unable to prepare FTL root for %s: %w", module.Config.Module, err)
	}

	logger.Infof("Using build command '%s'", module.Config.Build)
	err := exec.Command(ctx, log.Debug, module.Config.Dir, "bash", "-c", module.Config.Build).RunBuffered(ctx)
	if err != nil {
		return fmt.Errorf("failed to build module %q: %w", module.Config.Module, err)
	}

	return nil
}

// SetPOMProperties updates the ftl.version properties in the
// pom.xml file in the given base directory.
func SetPOMProperties(ctx context.Context, baseDir string) error {
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
	properties := root.SelectElement("properties")
	if properties == nil {
		return fmt.Errorf("unable to find <properties> in %s", pomFile)
	}
	version := properties.SelectElement("ftl.version")
	if version == nil {
		return fmt.Errorf("unable to find <properties>/<ftl.version> in %s", pomFile)
	}
	version.SetText(ftlVersion)

	return tree.WriteToFile(pomFile)
}

func prepareFTLRoot(module Module) error {
	buildDir := module.Config.Abs().DeployDir
	if err := os.MkdirAll(buildDir, 0700); err != nil {
		return err
	}

	fileContent := fmt.Sprintf(`
SchemaExtractorRuleSet:
  ExtractSchemaRule:
    active: true
    output: %s
`, buildDir)

	detektYmlPath := filepath.Join(buildDir, "detekt.yml")
	if err := os.WriteFile(detektYmlPath, []byte(fileContent), 0600); err != nil {
		return fmt.Errorf("unable to configure detekt for %s: %w", module.Config.Module, err)
	}

	mainFilePath := filepath.Join(buildDir, "main")

	mainFile := `#!/bin/bash
exec java -cp "classes:$(cat classpath.txt)" xyz.block.ftl.main.MainKt
`
	if err := os.WriteFile(mainFilePath, []byte(mainFile), 0700); err != nil { //nolint:gosec
		return fmt.Errorf("unable to configure main executable for %s: %w", module.Config.Module, err)
	}
	return nil
}

func generateExternalModules(ctx context.Context, module Module, sch *schema.Schema) error {
	logger := log.FromContext(ctx)
	funcs := maps.Clone(scaffoldFuncs)

	// Wipe the modules directory to ensure we don't have any stale modules.
	_ = os.RemoveAll(filepath.Join(module.Config.Dir, "target", "generated-sources", "ftl"))

	logger.Debugf("Generating external modules")
	return internal.ScaffoldZip(kotlinruntime.ExternalModuleTemplates(), module.Config.Dir, externalModuleContext{
		module: module,
		Schema: sch,
	}, scaffolder.Exclude("^go.mod$"), scaffolder.Functions(funcs))
}

var scaffoldFuncs = scaffolder.FuncMap{
	"comment": func(s []string) string {
		if len(s) == 0 {
			return ""
		}
		var sb strings.Builder
		sb.WriteString("/**\n")
		for _, line := range s {
			sb.WriteString(" * ")
			sb.WriteString(line)
			sb.WriteString("\n")
		}
		sb.WriteString(" */\n")
		return sb.String()
	},
	"type": genType,
	"is": func(kind string, t schema.Node) bool {
		return reflect.Indirect(reflect.ValueOf(t)).Type().Name() == kind
	},
	"imports": func(m *schema.Module) []string {
		imports := sets.NewSet[string]()
		_ = schema.VisitExcludingMetadataChildren(m, func(n schema.Node, next func() error) error { //nolint:errcheck
			switch n.(type) {
			case *schema.Data:
				imports.Add("xyz.block.ftl.Data")

			case *schema.Enum:
				imports.Add("xyz.block.ftl.Enum")

			case *schema.Verb:
				imports.Append("xyz.block.ftl.Context", "xyz.block.ftl.Ignore", "xyz.block.ftl.Verb")

			case *schema.Time:
				imports.Add("java.time.OffsetDateTime")

			default:
			}
			return next()
		})
		importsList := imports.ToSlice()
		sort.Strings(importsList)
		return importsList
	},
}

func genType(module *schema.Module, t schema.Type) string {
	switch t := t.(type) {
	case *schema.Ref:
		decl := module.Resolve(schema.Ref{
			Module: t.Module,
			Name:   t.Name,
		})
		if decl != nil {
			if data, ok := decl.Symbol.(*schema.Data); ok {
				if len(data.Fields) == 0 {
					return "ftl.builtin.Empty"
				}
			}
		}

		desc := t.Name
		if t.Module != "" {
			desc = "ftl." + t.Module + "." + desc
		}
		if len(t.TypeParameters) > 0 {
			desc += "<"
			for i, tp := range t.TypeParameters {
				if i != 0 {
					desc += ", "
				}
				desc += genType(module, tp)
			}
			desc += ">"
		}
		return desc

	case *schema.Time:
		return "OffsetDateTime"

	case *schema.Array:
		return "List<" + genType(module, t.Element) + ">"

	case *schema.Map:
		return "Map<" + genType(module, t.Key) + ", " + genType(module, t.Value) + ">"

	case *schema.Optional:
		return genType(module, t.Type) + "? = null"

	case *schema.Bytes:
		return "ByteArray"

	case *schema.Bool:
		return "Boolean"

	case *schema.Int:
		return "Long"

	case *schema.Float, *schema.String, *schema.Any, *schema.Unit:
		return t.String()
	}
	panic(fmt.Sprintf("unsupported type %T", t))
}
