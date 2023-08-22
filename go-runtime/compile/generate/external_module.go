package generate

import (
	_ "embed" // For embedding templates.
	"fmt"
	"io"
	"path"
	"reflect"
	"strings"
	"text/template"

	"github.com/alecthomas/errors"
	"github.com/iancoleman/strcase"
	"golang.org/x/exp/maps"

	"github.com/TBD54566975/ftl/backend/schema"
)

type externalModuleCtx struct {
	ImportRoot string
	*schema.Module
}

//go:embed external_module.go.tmpl
var moduleTmplSource string
var moduleTmpl = template.Must(template.New("external_module.go.tmpl").
	Funcs(template.FuncMap{
		"title": strcase.ToCamel,
		"comment": func(s []string) string {
			if len(s) == 0 {
				return ""
			}
			return "// " + strings.Join(s, "\n// ")
		},
		"type": genType,
		"is": func(kind string, t schema.Node) bool {
			return reflect.Indirect(reflect.ValueOf(t)).Type().Name() == kind
		},
		"imports": func(m *externalModuleCtx) []string {
			pkgs := map[string]bool{}
			_ = schema.Visit(m, func(n schema.Node, next func() error) error {
				switch n := n.(type) {
				case *schema.DataRef:
					if n.Module != "" {
						pkgs[path.Join(m.ImportRoot, n.Module)] = true
					}
				case *schema.Time:
					pkgs["time"] = true
				default:
				}
				return next()
			})
			return maps.Keys(pkgs)
		},
	}).
	Parse(moduleTmplSource))

// ExternalModule Go stubs for the given module.
func ExternalModule(w io.Writer, module *schema.Module, importRoot string) error {
	return errors.WithStack(moduleTmpl.Execute(w, &externalModuleCtx{
		ImportRoot: importRoot,
		Module:     module,
	}))
}

func genType(t schema.Type) string {
	switch t := t.(type) {
	case *schema.DataRef:
		return t.Name

	case *schema.VerbRef:
		return t.Name

	case *schema.Float:
		return "float64"

	case *schema.Time:
		return "time.Time"

	case *schema.Int, *schema.Bool, *schema.String:
		return strings.ToLower(t.String())

	case *schema.Array:
		return "[]" + genType(t.Element)

	case *schema.Map:
		return "map[" + genType(t.Key) + "]" + genType(t.Value)
	}
	panic(fmt.Sprintf("unsupported type %T", t))
}
