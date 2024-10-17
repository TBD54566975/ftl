{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "ftl.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "ftl.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
app.kubernetes.io/name: {{ include "ftl.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Values.customLabels -}}
{{ toYaml .Values.customLabels }}
{{- end -}}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "ftl-controller.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ftl.fullname" . }}
app.kubernetes.io/component: controller
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
{{- define "ftl-runner.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ftl.fullname" . }}
app.kubernetes.io/component: runner
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
