package generate

import (
	_ "embed"
	"io"
	"text/template"
)

//go:embed go.mod.tmpl
var goModTmplSource string
var goModTmpl = template.Must(template.New("go.mod.tmpl").Parse(goModTmplSource))

type GoModConfig struct {
	// Replace directives
	Replace map[string]string
}

// GenerateGoMod generates a go.mod file.
func GenerateGoMod(w io.Writer, config GoModConfig, importRoot string) error {
	return goModTmpl.Execute(w, config)
}
