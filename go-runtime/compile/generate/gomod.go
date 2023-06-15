package generate

import (
	_ "embed"
	"io"
	"text/template"

	"github.com/alecthomas/errors"
)

//go:embed go.mod.tmpl
var goModTmplSource string
var goModTmpl = template.Must(template.New("go.mod.tmpl").Parse(goModTmplSource))

type GoModConfig struct {
	// Replace directives
	Replace map[string]string
	// If present, a path to the checked out FTL source code.
	FTLSource string
}

// GenerateGoMod generates a go.mod file.
func GenerateGoMod(w io.Writer, config GoModConfig) error {
	return errors.WithStack(goModTmpl.Execute(w, config))
}
