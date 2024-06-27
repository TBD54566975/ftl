package builder

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/scaffolder"
	"github.com/reugn/go-quartz/logger"
)

type externalModuleContext struct {
	ModuleDir string
	*schema.Schema
	// GoVersion    string
	// FTLVersion   string
	// Main string
	// Replacements []*modfile.Replace
}

// type goVerb struct {
// 	Name        string
// 	Package     string
// 	MustImport  string
// 	HasRequest  bool
// 	HasResponse bool
// }

type mainModuleContext struct {
	// GoVersion  string
	// FTLVersion string
	Name string
	// Verbs        []goVerb
	// Replacements []*modfile.Replace
	// SumTypes     []goSumType
}

func Build(ctx context.Context, sch *schema.Schema, name, moduleDir, build, deployDir, schemaFilename string) error {
	if _, err := os.Stat(deployDir); os.IsNotExist(err) {
		if err := os.MkdirAll(deployDir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create deploy directory: %w", err)
		}
	}
	// TODO: update module package FTL dependencies

	logger.Debugf("Generating external modules")
	if err := generateExternalModules(externalModuleContext{
		ModuleDir: moduleDir,
		// GoVersion:    goModVersion,
		// FTLVersion:   ftlVersion,
		Schema: sch,
		// Main:         config.Module,
		// Replacements: replacements,
	}); err != nil {
		return fmt.Errorf("failed to generate external modules: %w", err)
	}

	goExecPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not get go executable path: %w", err)
	}
	err = exec.Command(ctx, log.Debug,
		filepath.Dir(goExecPath),
		"./ftl-swift-compile",
		"--name", name,
		"--root-path", moduleDir,
		"--deploy-path", deployDir,
		"--schema-filename", schemaFilename).RunBuffered(ctx)
	if err != nil {
		return fmt.Errorf("failed to build module %q: %w", name, err)
	}

	// Process the schema data here

	module, err := schema.ModuleFromProtoFile(filepath.Join(moduleDir, "_ftl", schemaFilename))
	if err != nil {
		return fmt.Errorf("failed to load module schema: %w", err)
	}

	// scaffold main package
	funcs := maps.Clone(scaffoldFuncs)
	if err := internal.ScaffoldZip(buildTemplateFiles(), moduleDir, mainModuleContext{
		// GoVersion:    goModVersion,
		// FTLVersion:   ftlVersion,
		Name: module.Name,
		// Verbs:        goVerbs,
		// Replacements: replacements,
		// SumTypes:     getSumTypes(result.Module, sch, result.NativeNames),
	}, scaffolder.Exclude("^go.mod$"), scaffolder.Functions(funcs)); err != nil {
		return err
	}

	err = exec.Command(ctx, log.Debug, moduleDir, "bash", "-c", build).RunBuffered(ctx)
	if err != nil {
		return fmt.Errorf("failed to compile module: %w", err)
	}

	return nil
}

func generateExternalModules(context externalModuleContext) error {
	// Wipe the modules directory to ensure we don't have any stale modules.
	// err := os.RemoveAll(filepath.Join(context.ModuleDir, "_ftl", "swift", "modules"))
	// if err != nil {
	// 	return err
	// }

	// funcs := maps.Clone(scaffoldFuncs)
	// return internal.ScaffoldZip(externalModuleTemplateFiles(), context.ModuleDir, context, scaffolder.Exclude("^go.mod$"), scaffolder.Functions(funcs))
	return nil
}

var scaffoldFuncs = scaffolder.FuncMap{
	// "comment": schema.EncodeComments,
	// "type":    genType,
	// "is": func(kind string, t schema.Node) bool {
	// 	return stdreflect.Indirect(stdreflect.ValueOf(t)).Type().Name() == kind
	// },
	// "imports": func(m *schema.Module) map[string]string {
	// 	imports := map[string]string{}
	// 	_ = schema.VisitExcludingMetadataChildren(m, func(n schema.Node, next func() error) error { //nolint:errcheck
	// 		switch n := n.(type) {
	// 		case *schema.Ref:
	// 			if n.Module == "" || n.Module == m.Name {
	// 				break
	// 			}
	// 			imports[path.Join("ftl", n.Module)] = "ftl" + n.Module

	// 			for _, tp := range n.TypeParameters {
	// 				tpRef, err := schema.ParseRef(tp.String())
	// 				if err != nil {
	// 					panic(err)
	// 				}
	// 				if tpRef.Module != "" && tpRef.Module != m.Name {
	// 					imports[path.Join("ftl", tpRef.Module)] = "ftl" + tpRef.Module
	// 				}
	// 			}

	// 		case *schema.Time:
	// 			imports["time"] = "stdtime"

	// 		case *schema.Optional, *schema.Unit:
	// 			imports["github.com/TBD54566975/ftl/go-runtime/ftl"] = ""

	// 		case *schema.Topic:
	// 			if n.IsExported() {
	// 				imports["github.com/TBD54566975/ftl/go-runtime/ftl"] = ""
	// 			}
	// 		default:
	// 		}
	// 		return next()
	// 	})
	// 	return imports
	// },
	// "value": func(v schema.Value) string {
	// 	switch t := v.(type) {
	// 	case *schema.StringValue:
	// 		return fmt.Sprintf("%q", t.Value)
	// 	case *schema.IntValue:
	// 		return strconv.Itoa(t.Value)
	// 	case *schema.TypeValue:
	// 		return t.Value.String()
	// 	}
	// 	panic(fmt.Sprintf("unsupported value %T", v))
	// },
	// "enumInterfaceFunc": func(e schema.Enum) string {
	// 	r := []rune(e.Name)
	// 	for i, c := range r {
	// 		if unicode.IsUpper(c) {
	// 			r[i] = unicode.ToLower(c)
	// 		} else {
	// 			break
	// 		}
	// 	}
	// 	return string(r)
	// },
	// "basicType": func(m *schema.Module, v schema.EnumVariant) bool {
	// 	switch val := v.Value.(type) {
	// 	case *schema.IntValue, *schema.StringValue:
	// 		return false // This func should only return true for type enums
	// 	case *schema.TypeValue:
	// 		if _, ok := val.Value.(*schema.Ref); !ok {
	// 			return true
	// 		}
	// 	}
	// 	return false
	// },
	// "mainImports": func(ctx mainModuleContext) []string {
	// 	imports := sets.NewSet[string]()
	// 	if len(ctx.Verbs) > 0 {
	// 		imports.Add(ctx.Name)
	// 	}
	// 	for _, v := range ctx.Verbs {
	// 		imports.Add(strings.TrimPrefix(v.MustImport, "ftl/"))
	// 	}
	// 	for _, st := range ctx.SumTypes {
	// 		if i := strings.LastIndex(st.Discriminator, "."); i != -1 {
	// 			imports.Add(st.Discriminator[:i])
	// 		}
	// 		for _, v := range st.Variants {
	// 			if i := strings.LastIndex(v.Type, "."); i != -1 {
	// 				imports.Add(v.Type[:i])
	// 			}
	// 		}
	// 	}
	// 	out := imports.ToSlice()
	// 	slices.Sort(out)
	// 	return out
	// },
	// "schemaType": schemaType,
	// // A standalone enum variant is one that is purely an alias to a type and does not appear
	// // elsewhere in the schema.
	// "isStandaloneEnumVariant": func(v schema.EnumVariant) bool {
	// 	tv, ok := v.Value.(*schema.TypeValue)
	// 	if !ok {
	// 		return false
	// 	}
	// 	if ref, ok := tv.Value.(*schema.Ref); ok {
	// 		return ref.Name != v.Name
	// 	}

	// 	return false
	// },
	// "sumTypes": func(m *schema.Module) []*schema.Enum {
	// 	out := []*schema.Enum{}
	// 	for _, d := range m.Decls {
	// 		switch d := d.(type) {
	// 		// Type enums (i.e. sum types) are all the non-value enums
	// 		case *schema.Enum:
	// 			if !d.IsValueEnum() && d.IsExported() {
	// 				out = append(out, d)
	// 			}
	// 		default:
	// 		}
	// 	}
	// 	return out
	// },
}
