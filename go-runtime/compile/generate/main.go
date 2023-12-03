package generate

import (
	_ "embed"
	"io"
	"strings"
	"text/template"

	"github.com/TBD54566975/ftl/backend/schema"
)

//go:embed main.go.tmpl
var mainTmplSource string
var mainTmpl = template.Must(template.New("main.go.tmpl").
	Funcs(template.FuncMap{
		"ExportGoName": func(s string) string { return strings.ToTitle(s[:1]) + s[1:] },
	}).
	Parse(mainTmplSource))

type mainTmplCtx struct {
	ImportRoot string
	*schema.Module
}

func Main(w io.Writer, module *schema.Module, importRoot string) error {
	return mainTmpl.Execute(w, &mainTmplCtx{
		ImportRoot: importRoot,
		Module:     module,
	})
}
