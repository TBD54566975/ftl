{{- $moduleName := .Name -}}
// Code generated by FTL. DO NOT EDIT.
{{.Comments|comment -}}
package ftl.{{.Name}}

{{- $imports := (.|imports)}}
{{- if $imports}}
{{range $import := $imports}}
import {{$import}}
{{- end}}
{{else}}
{{end -}}

{{range .Decls}}
{{- if .IsExported}}
{{- if is "Data" . }}
{{- if and (eq $moduleName "builtin") (eq .Name "Empty")}}
{{.Comments|comment -}}
@Data
class Empty
{{- else if .Fields}}
{{.Comments|comment -}}
@Data
data class {{.Name|title}}
{{- if .TypeParameters}}<
{{- range $i, $tp := .TypeParameters}}
{{- if $i}}, {{end}}{{$tp}}
{{- end -}}
>{{- end}}(
  {{- range .Fields}}
  val {{.Name}}: {{type $ .Type}},
  {{- end}}
)
{{end}}

{{- else if is "Verb" . }}
{{.Comments|comment -}}@Verb
@Ignore
{{- if and (eq (type $ .Request) "Unit") (eq (type $ .Response) "Unit")}}
fun {{.Name|lowerCamel}}(context: Context): Unit = throw
    NotImplementedError("Verb stubs should not be called directly, instead use context.callEmpty(::{{.Name|lowerCamel}}, ...)")
{{- else if eq (type $ .Request) "Unit"}}
fun {{.Name|lowerCamel}}(context: Context): {{type $ .Response}} = throw
    NotImplementedError("Verb stubs should not be called directly, instead use context.callSource(::{{.Name|lowerCamel}}, ...)")
{{- else if eq (type $ .Response) "Unit"}}
fun {{.Name|lowerCamel}}(context: Context, req: {{type $ .Request}}): Unit = throw
    NotImplementedError("Verb stubs should not be called directly, instead use context.callSink(::{{.Name|lowerCamel}}, ...)")
{{- else}}
fun {{.Name|lowerCamel}}(context: Context, req: {{type $ .Request}}): {{type $ .Response}} = throw
    NotImplementedError("Verb stubs should not be called directly, instead use context.call(::{{.Name|lowerCamel}}, ...)")
{{- end}}
{{- end}}

{{- end}}
{{- end}}
