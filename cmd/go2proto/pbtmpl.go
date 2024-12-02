package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/TBD54566975/ftl/internal/schema/strcase"
)

var tmpl = template.Must(template.New("proto").
	Funcs(template.FuncMap{
		"typeof":       func(t any) string { return reflect.Indirect(reflect.ValueOf(t)).Type().Name() },
		"toLowerCamel": strcase.ToLowerCamel,
		"toUpperCamel": strcase.ToUpperCamel,
		"toLowerSnake": strcase.ToLowerSnake,
		"toUpperSnake": strcase.ToUpperSnake,
		"trimPrefix":   strings.TrimPrefix,
	}).
	Parse(`
// THIS FILE IS GENERATED; DO NOT MODIFY
syntax = "proto3";

package {{ .Package }};
{{ range .Imports }}
import "{{.}}";
{{- end}}
{{ range $name, $value := .Options }}
option {{ $name }} = {{ $value }};
{{- end }}
{{ range $decl := .OrderedDecls }}
{{- if eq (typeof $decl) "Message" }}
message {{ .Name }} {
{{- range $name, $field := .Fields }}
  {{ if .Repeated }}repeated {{else if .Optional}}optional {{ end }}{{ .Type }} {{ .Name | toLowerSnake }} = {{ .ID }};
{{- end }}
}
{{- else if eq (typeof $decl) "Enum" }}
enum {{ .Name }} {
{{- range $value, $name := .ByValue }}
  {{ $name | toUpperSnake }} = {{ $value }};
{{- end }}
}
{{- else if eq (typeof $decl) "SumType" }}
{{ $sumtype := . }}
message {{ .Name }} {
  oneof value {
{{- range $name, $id := .Variants }}
    {{ $name }} {{ trimPrefix $name $sumtype.Name | toLowerSnake }} = {{ $id }};
{{- end }}
  }
}
{{- end }}
{{ end }}
`))

type RenderContext struct {
	Config
	File
}

func render(out *os.File, config Config, file File) error {
	err := tmpl.Execute(out, RenderContext{
		Config: config,
		File:   file,
	})
	if err != nil {
		return fmt.Errorf("template error: %w", err)
	}
	return nil
}
