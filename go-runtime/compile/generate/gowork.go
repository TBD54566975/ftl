package generate

import (
	_ "embed"
	"io"
	"text/template"

	"github.com/alecthomas/errors"
)

//go:embed go.work.tmpl
var goWorkTmplSource string
var goWorkTmpl = template.Must(template.New("go.mod.tmpl").Parse(goWorkTmplSource))

// GenerateGoWork generates a go.work file.
func GenerateGoWork(w io.Writer, modules []string, importRoot string) error {
	return errors.WithStack(goWorkTmpl.Execute(w, modules))
}
