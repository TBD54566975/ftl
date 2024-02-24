package buildengine

import (
	"context"
	"fmt"
	sets "github.com/deckarep/golang-set/v2"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	kotlinruntime "github.com/TBD54566975/ftl/kotlin-runtime"
	"github.com/TBD54566975/scaffolder"
	"github.com/beevik/etree"
	"golang.org/x/exp/maps"
)

type externalModuleContext struct {
	module Module
	*schema.Schema
}

func (e externalModuleContext) ExternalModules() []*schema.Module {
	depsSet := make(map[string]struct{})
	for _, dep := range e.module.Dependencies {
		depsSet[dep] = struct{}{}
	}

	modules := make([]*schema.Module, 0)
	for _, module := range e.Modules {
		if _, exists := depsSet[module.Name]; exists || module.Name == "builtin" {
			modules = append(modules, module)
		}
	}
	return modules
}

func buildKotlin(ctx context.Context, sch *schema.Schema, module Module) error {
	logger := log.FromContext(ctx)
	if err := SetPOMProperties(ctx, filepath.Join(module.Dir, "..")); err != nil {
		return fmt.Errorf("unable to update ftl.version in %s: %w", module.Dir, err)
	}

	if err := generateExternalModules(ctx, module, sch); err != nil {
		return fmt.Errorf("unable to generate external modules for %s: %w", module.Module, err)
	}

	if err := prepareFTLRoot(module); err != nil {
		return fmt.Errorf("unable to prepare FTL root for %s: %w", module.Module, err)
	}

	logger.Debugf("Using build command '%s'", module.Build)
	err := exec.Command(ctx, log.Debug, module.Dir, "bash", "-c", module.Build).RunBuffered(ctx)
	if err != nil {
		return fmt.Errorf("failed to build module: %w", err)
	}

	return nil
}

// SetPOMProperties updates the ftl.version and ftlEndpoint properties in the
// pom.xml file in the given base directory.
func SetPOMProperties(ctx context.Context, baseDir string) error {
	logger := log.FromContext(ctx)
	ftlVersion := ftl.Version
	if ftlVersion == "dev" {
		ftlVersion = "1.0-SNAPSHOT"
	}

	ftlEndpoint := os.Getenv("FTL_ENDPOINT")
	if ftlEndpoint == "" {
		ftlEndpoint = "http://127.0.0.1:8892"
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

	endpoint := properties.SelectElement("ftlEndpoint")
	if endpoint == nil {
		logger.Warnf("unable to find <properties>/<ftlEndpoint> in %s", pomFile)
	} else {
		endpoint.SetText(ftlEndpoint)
	}

	return tree.WriteToFile(pomFile)
}

func prepareFTLRoot(module Module) error {
	buildDir := filepath.Join(module.Dir, "target")
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
		return fmt.Errorf("unable to configure detekt for %s: %w", module.Module, err)
	}

	mainFilePath := filepath.Join(buildDir, "main")

	mainFile := `#!/bin/bash
exec java -cp "classes:$(cat classpath.txt)" xyz.block.ftl.main.MainKt
`
	if err := os.WriteFile(mainFilePath, []byte(mainFile), 0700); err != nil { //nolint:gosec
		return fmt.Errorf("unable to configure main executable for %s: %w", module.Module, err)
	}
	return nil
}

func generateExternalModules(ctx context.Context, module Module, sch *schema.Schema) error {
	logger := log.FromContext(ctx)
	config := module.ModuleConfig
	funcs := maps.Clone(scaffoldFuncs)

	// Wipe the modules directory to ensure we don't have any stale modules.
	_ = os.RemoveAll(filepath.Join(config.Dir, "target", "generated-sources", "ftl"))

	logger.Debugf("Generating external modules")
	return internal.ScaffoldZip(kotlinruntime.ExternalModuleTemplates(), config.Dir, externalModuleContext{
		module: module,
		Schema: sch,
	}, scaffolder.Functions(funcs))
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
		_ = schema.Visit(m, func(n schema.Node, next func() error) error {
			switch n := n.(type) {
			case *schema.DataRef:
				decl := m.Resolve(schema.Ref{
					Module: n.Module,
					Name:   n.Name,
				})
				if decl != nil {
					if data, ok := decl.Decl.(*schema.Data); ok {
						if len(data.Fields) == 0 {
							imports.Add("ftl.builtin.Empty")
							break
						}
					}
				}

				if n.Module == "" {
					break
				}

				imports.Add("ftl." + n.Module + "." + n.Name)

				for _, tp := range n.TypeParameters {
					tpRef, err := schema.ParseDataRef(tp.String())
					if err != nil {
						return err
					}
					if tpRef.Module != "" && tpRef.Module != m.Name {
						imports.Add("ftl." + tpRef.Module + "." + tpRef.Name)
					}
				}
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
	case *schema.DataRef:
		decl := module.Resolve(schema.Ref{
			Module: t.Module,
			Name:   t.Name,
		})
		if decl != nil {
			if data, ok := decl.Decl.(*schema.Data); ok {
				if len(data.Fields) == 0 {
					return "Empty"
				}
			}
		}

		desc := t.Name
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
