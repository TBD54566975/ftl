package sdkgo

import (
	_ "embed" // For embedding templates.
	"io"
	"strings"
	"text/template"

	"github.com/alecthomas/errors"
	"github.com/iancoleman/strcase"
	"golang.org/x/exp/maps"

	"github.com/TBD54566975/ftl/schema"
)

//go:embed module.go.tmpl
var tmplSource string
var tmpl = template.Must(template.New("module").
	Funcs(template.FuncMap{
		"title": strcase.ToCamel,
		"comment": func(s []string) string {
			return "// " + strings.Join(s, "\n// ")
		},
		"imports": func(m schema.Module) []string {
			pkgs := map[string]bool{}
			_ = schema.Visit(m, func(n schema.Node, next func() error) error {
				switch n := n.(type) {
				case schema.VerbRef:
					if n.Module != "" {
						pkgs[n.Module] = true
					}
				case schema.DataRef:
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
	Parse(tmplSource))

// Generate Go stubs for the given module.
func Generate(module schema.Module, w io.Writer) error {
	return errors.WithStack(tmpl.Execute(w, module))
}
