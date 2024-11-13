// ignore_for_file: unused_import
library {{ .Name | lower }};

import 'dart:convert';
import 'dart:typed_data';
import 'ftl_client.dart';
{{- range .Imports }}
import '{{. | lower }}.dart' as {{. | lower}};
{{- end}}

{{ range .Data }}
class {{ .Name | camel }}{{ .TypeParameters | dartTypeParameters }}{
{{- range .Fields }}
  {{ .Type | dartType }} {{ .Name }};
{{- end }}

  {{ .Name | camel }}({{ if .Fields }}{ {{ range .Fields }}{{ if not (eq (.Type | typename) "Optional")}} required{{end}} this.{{ .Name }}, {{ end}} }{{ end }});

  Map<String, dynamic> toJson() {
    return {
{{- range .Fields}}
      '{{ .Name }}': ((dynamic v) => {{ .Type | serialize }})({{ .Name }}),
{{- end }}
    };
  }

  factory {{ .Name | camel }}.fromJson(Map<String, dynamic> map{{ .TypeParameters | tpMappingFuncs }}) {
    return {{ .Name | camel }}(
      {{ . | fromJsonFields }}
    );
  }
}
{{ end}}

class {{ .Name | camel }}Client {
  final FTLHttpClient ftlClient;

  {{ .Name | camel }}Client({required this.ftlClient});
{{ range .Verbs }}
{{- $verb := . -}}
{{- range .Metadata }}
{{ if eq "MetadataIngress" (. | typename) }}
  Future<{{ $verb.Response | bodyType }}> {{ $verb.Name }}(
    {{ $verb.Request | bodyType }} request, { 
    Map<String, String>? headers,
  }) async {
    {{ if eq .Method "GET" -}}
    final response = await ftlClient.{{ .Method | lower }}(
      '{{ $verb | url }}', 
      requestJson: json.encode(request.toJson()),
      headers: headers,
    );
    {{ else -}}
    final response = await ftlClient.{{ .Method | lower }}('{{ $verb | url }}', request: request.toJson());
    {{ end -}}
    if (response.statusCode == 200) {
      final body = json.decode(utf8.decode(response.bodyBytes));
      return {{ $verb.Response | bodyType }}.fromJson(body);
    } else {
      throw Exception('Failed to get {{ $verb.Name }} response');
    }
  }
{{- end }}
{{- end }}
{{- end }}
}
