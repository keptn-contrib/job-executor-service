{{/*
Expand the name of the chart.
*/}}
{{- define "job-executor-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "job-executor-service.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "job-executor-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "job-executor-service.labels" -}}
helm.sh/chart: {{ include "job-executor-service.chart" . }}
{{ include "job-executor-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "job-executor-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "job-executor-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "job-executor-service.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "job-executor-service.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{- define "distributor.default-container-security-context" }}
securityContext:
  runAsNonRoot: true
  runAsUser: 65532
  readOnlyRootFilesystem: false
  allowPrivilegeEscalation: false
  privileged: false
{{- end}}

{{- define "distributor.container-security-context" -}}
{{- if (.Values.distributor).containerSecurityContext -}}
{{- if .Values.distributor.containerSecurityContext.overwrite -}}
securityContext:
{{- range $key, $value := omit .Values.distributor.containerSecurityContext "overwrite" }}
  {{ $key }}: {{- toYaml $value | nindent 4 }}
{{- end -}}
{{- else -}}
{{ include "distributor.default-container-security-context" . }}
{{- end -}}
{{- else -}}
{{ include "distributor.default-container-security-context" . }}
{{- end -}}
{{- end -}}
