package main

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"github.com/TBD54566975/ftl/internal/schema/strcase"
)

var _ fmt.Stringer

var go2protoTmpl = template.Must(template.New("go2proto.mapper.go").
	Funcs(template.FuncMap{
		"typeof": func(t any) string { return reflect.Indirect(reflect.ValueOf(t)).Type().Name() },
		// Return true if the type is a builtin proto type.
		"isBuiltin": func(t Field) bool {
			switch t.Type {
			case "int32", "int64", "uint32", "uint64", "float", "double", "bool", "string":
				return true
			}
			return false
		},
		"goProtoImport": func(g Go2ProtoContext) (string, error) {
			unquoted, err := strconv.Unquote(g.Options["go_package"])
			if err != nil {
				return "", fmt.Errorf("go_package must be a quoted string: %w", err)
			}
			parts := strings.Split(unquoted, ";")
			return parts[0], nil
		},
		"sumTypeVariantName": func(s string, v string) string {
			return strings.TrimPrefix(v, s)
		},
		"toLower":      strings.ToLower,
		"toUpper":      strings.ToUpper,
		"toLowerCamel": strcase.ToLowerCamel,
		"toUpperCamel": strcase.ToUpperCamel,
		"toLowerSnake": strcase.ToLowerSnake,
		"toUpperSnake": strcase.ToUpperSnake,
		"trimPrefix":   strings.TrimPrefix,
	}).
	Parse(`// Code generated by go2proto. DO NOT EDIT.

package {{ .GoPackage }}

import "fmt"
import destpb "{{ . | goProtoImport }}"
import "google.golang.org/protobuf/proto"
import "google.golang.org/protobuf/types/known/timestamppb"
import "google.golang.org/protobuf/types/known/durationpb"

var _ fmt.Stringer
var _ = timestamppb.Timestamp{}
var _ = durationpb.Duration{}

{{range $decl := .OrderedDecls }}
{{- if eq (typeof $decl) "Message" }}
func (x *{{ .Name }}) ToProto() *destpb.{{ .Name }} {
	out := &destpb.{{ .Name }}{}
{{- range $field := .Fields }}
{{- if . | isBuiltin }}
{{- if $field.Optional}}
	out.{{ $field.EscapedName }} = proto.{{ $field.ProtoGoType | toUpperCamel }}({{ $field.ProtoGoType }}({{if $field.Pointer}}*{{end}}x.{{ $field.Name }}))
{{- else }}
	out.{{ $field.EscapedName }} = {{ $field.ProtoGoType }}(x.{{ $field.Name }})
{{- end}}
{{- else if eq $field.Type "google.protobuf.Timestamp" }}
	out.{{ $field.EscapedName }} = timestamppb.New(x.{{ $field.Name }})
{{- else if eq $field.Type "google.protobuf.Duration" }}
	out.{{ $field.EscapedName }} = durationpb.New(x.{{ $field.Name }})
{{- else if eq (.Type | $.TypeOf) "Message" }}
	out.{{ $field.EscapedName }} = x.{{ $field.Name }}.ToProto()
{{- else if eq (.Type | $.TypeOf) "Enum" }}
	out.{{ $field.EscapedName }} = x.{{ $field.Name }}.ToProto()
{{- else if eq (.Type | $.TypeOf) "SumType" }}
	out.{{ $field.EscapedName }} = {{ $field.Type }}ToProto(x.{{ $field.Name }})
{{- end}}
{{- end}}
	return out
}
{{- else if eq (typeof $decl) "Enum" }}
func (x {{ .Name }}) ToProto() destpb.{{ .Name }} {
	return destpb.{{ .Name }}(x)
}
{{- else if eq (typeof $decl) "SumType" }}
{{- $sumtype := . }}
func {{ .Name }}ToProto(value {{ .Name }}) *destpb.{{ .Name }} {
	switch value := value.(type) {
	{{- range $variant, $id := .Variants }}
	case *{{ $variant }}:
		return &destpb.{{ $sumtype.Name }}{
			Value: &destpb.{{ $sumtype.Name }}_{{ sumTypeVariantName $sumtype.Name $variant }}{value.ToProto()},
		}
	{{- end }}
	default:
		panic(fmt.Sprintf("unknown variant: %T", value))
	}
}
{{- end}}
{{ end}}
		`))

type Go2ProtoContext struct {
	Config
	File
}

func renderToProto(out *os.File, config Config, file File) error {
	if config.Options["go_package"] == "" {
		return fmt.Errorf("go_package must be set in the protobuf options")
	}
	err := go2protoTmpl.Execute(out, Go2ProtoContext{
		Config: config,
		File:   file,
	})
	if err != nil {
		return fmt.Errorf("template error: %w", err)
	}
	return nil
}