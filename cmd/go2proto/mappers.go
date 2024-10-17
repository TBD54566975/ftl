package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"github.com/TBD54566975/ftl/internal/schema/strcase"
)

var mappersTmpl = template.Must(template.New("mappers.g2p.go").
	Funcs(template.FuncMap{
		"typeof":       func(t any) string { return reflect.Indirect(reflect.ValueOf(t)).Type().Name() },
		"toLowerCamel": strcase.ToLowerCamel,
		"toUpperCamel": strcase.ToUpperCamel,
		"toLowerSnake": strcase.ToLowerSnake,
		"toUpperSnake": strcase.ToUpperSnake,
		"trimPrefix":   strings.TrimPrefix,
	}).
	Parse(`
package {{ .Package }}

import {{.ProtoAlias}} "{{.ProtoImport}}"
import "google.golang.org/protobuf/proto"

{{ range $decl := .OrderedDecls }}
{{- if eq (typeof $decl) "Message" }}
func (m *{{.Name}}) ToProto() proto.Message {
	return &{{$.ProtoAlias}}.{{.Name}}{
{{- range $name, $field := .Fields }}
		{{ .Name | toUpperCamel }}: m.{{ .Name }},
{{- end }}
	}
}
{{- else if eq (typeof $decl) "Enum" }}
{{- else if eq (typeof $decl) "SumType" }}
{{- end }}
{{ end }}
`))

type MappersContext struct {
	Package     string // Our package name
	ProtoImport string // The import path for the proto package
	ProtoAlias  string // The import alias for the proto package
	Config
	File
}

func renderMappers(localPath string, config Config, file File) error {
	protoPackage, ok := config.Options["go_package"]
	if !ok {
		return fmt.Errorf("missing go_package option")
	}
	var err error
	protoPackage, err = strconv.Unquote(protoPackage)
	if err != nil {
		return fmt.Errorf("parse go_package: %w", err)
	}
	goPackageParts := strings.Split(protoPackage, ";")
	protoImport := goPackageParts[0]
	protoAlias := path.Base(goPackageParts[0])
	if len(goPackageParts) == 2 {
		protoAlias = goPackageParts[1]
	}
	out, err := os.Create(filepath.Join(localPath, "mappers.g2p.go~"))
	if err != nil {
		return fmt.Errorf("create mappers: %w", err)
	}
	defer out.Close()           //nolint:errcheck
	defer os.Remove(out.Name()) //nolint:errcheck
	err = mappersTmpl.Execute(out, MappersContext{
		Package:     filepath.Base(localPath),
		ProtoImport: protoImport,
		ProtoAlias:  protoAlias,
		Config:      config,
		File:        file,
	})
	if err != nil {
		return fmt.Errorf("render mappers: %w", err)
	}
	return os.Rename(out.Name(), strings.TrimRight(out.Name(), "~"))
}
