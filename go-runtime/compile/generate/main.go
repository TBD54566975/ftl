package generate

import (
	_ "embed"
	"io"
	"strings"
	"text/template"

	"github.com/alecthomas/errors"

	"github.com/TBD54566975/ftl/schema"
)

//go:embed main.go.tmpl
var mainTmplSource string
var mainTmpl = template.Must(template.New("main.go.tmpl").
	Funcs(template.FuncMap{
		"ExportGoName": func(s string) string { return strings.ToTitle(s[:1]) + s[1:] },
	}).
	Parse(mainTmplSource))

func GenerateMain(w io.Writer, module *schema.Module) error {
	return errors.WithStack(mainTmpl.Execute(w, module))
}
