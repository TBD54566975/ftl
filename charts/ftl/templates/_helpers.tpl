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
{{- define "ftl-provisioner.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ftl.fullname" . }}
app.kubernetes.io/component: provisioner
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
{{- define "ftl-cron.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ftl.fullname" . }}
app.kubernetes.io/component: cron
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
{{- define "ftl-http-ingress.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ftl.fullname" . }}
app.kubernetes.io/component: http-ingress
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
{{- define "ftl-timeline.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ftl.fullname" . }}
app.kubernetes.io/component: timeline
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
{{- define "ftl-lease.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ftl.fullname" . }}
app.kubernetes.io/component: lease
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
{{- define "ftl-console.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ftl.fullname" . }}
app.kubernetes.io/component: console
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
{{- define "ftl-admin.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ftl.fullname" . }}
app.kubernetes.io/component: admin
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
