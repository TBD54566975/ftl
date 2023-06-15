package generate

import (
	_ "embed" // For embedding templates.
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/template"

	"github.com/alecthomas/errors"
	"github.com/iancoleman/strcase"
	"golang.org/x/exp/maps"

	"github.com/TBD54566975/ftl/schema"
)

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
		"imports": func(m *schema.Module) []string {
			pkgs := map[string]bool{}
			_ = schema.Visit(m, func(n schema.Node, next func() error) error {
				switch n := n.(type) {
				case *schema.VerbRef:
					if n.Module != "" {
						pkgs[n.Module] = true
					}
				case *schema.DataRef:
					if n.Module != "" {
						pkgs[n.Module] = true
					}
				default:
				}
				return next()
			})
			return maps.Keys(pkgs)
		},
	}).
	Parse(moduleTmplSource))

// GenerateExternalModule Go stubs for the given module.
func GenerateExternalModule(w io.Writer, module *schema.Module) error {
	return errors.WithStack(moduleTmpl.Execute(w, module))
}

func genType(t schema.Type) string {
	switch t := t.(type) {
	case *schema.Float:
		return "float64"

	case *schema.Time:
		return "time.Time"

	case *schema.Int, *schema.Bool, *schema.String, *schema.DataRef, *schema.VerbRef:
		return strings.ToLower(t.String())

	case *schema.Array:
		return "[]" + genType(t.Element)

	case *schema.Map:
		return "map[" + genType(t.Key) + "]" + genType(t.Value)
	}
	panic(fmt.Sprintf("unsupported type %T", t))
}
