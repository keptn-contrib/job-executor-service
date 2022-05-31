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
app.kubernetes.io/version: {{ .Chart.AppVersion | trunc 63 | trimSuffix "-" | quote }}
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

{{/*
Create the name of the job service account to use
*/}}
{{- define "job-executor-service.jobConfig.serviceAccountName" -}}
{{- if ((.Values.jobConfig).serviceAccount).create | default true }}
{{- default (printf "%s-%s" (include "job-executor-service.fullname" .) (default "default-job-account" ((.Values.jobConfig).serviceAccount).name)) }}
{{- else }}
{{- default "default-job-account" ((.Values.jobConfig).serviceAccount).name }}
{{- end }}
{{- end }}

{{- define "job-executor-service.remote-control-plane.endpoint" }}
    {{- if ((.Values.remoteControlPlane).autoDetect).enabled }}
        {{- $detectedNamespace := include "job-executor-service.remote-control-plane.namespace" .}}
        {{- (printf "http://api-gateway-nginx.%s/api" $detectedNamespace) }}
    {{- else }}
        {{- (printf "%s://%s/api" .Values.remoteControlPlane.api.protocol .Values.remoteControlPlane.api.hostname) }}
    {{- end }}
{{- end }}

{{- define "job-executor-service.remote-control-plane.configuration-endpoint" }}
    {{- (printf "%s/configuration-service" (include "job-executor-service.remote-control-plane.endpoint" .)) }}
{{- end }}


{{/*
Helper functions of the auto detection feature of Keptn
*/}}
{{- define "job-executor-service.remote-control-plane.namespace" -}}
    {{- $detectedKeptnApiGateways := list }}

    {{- /* Find api-gateway-nginx service, which is used as keptn api gatway */ -}}
    {{- $services := lookup "v1" "Service" (.Values.remoteControlPlane.autoDetect.namespace | default "") "" }}
    {{- range $index, $srv := $services.items }}
        {{- if and (eq "api-gateway-nginx" $srv.metadata.name ) (hasPrefix "keptn-" ( get $srv.metadata.labels "app.kubernetes.io/part-of" )) }}
            {{- $detectedKeptnApiGateways = append $detectedKeptnApiGateways $srv }}
        {{- end }}
    {{- end }}

    {{- if eq (len $detectedKeptnApiGateways) 0 }}
        {{- fail "Unable to detect Keptn in the kubernetes cluster!" }}
    {{- end }}
    {{- if gt (len $detectedKeptnApiGateways) 1 }}
        {{- fail (printf "Detected more than one Keptn installation: %+v" $detectedKeptnApiGateways) }}
    {{- end }}

    {{- (index $detectedKeptnApiGateways 0).metadata.namespace }}
{{- end }}

{{- define "job-executor-service.remote-control-plane.token" -}}
    {{- if ((.Values.remoteControlPlane).autoDetect).enabled }}
        {{- $detectedNamespace := include "job-executor-service.remote-control-plane.namespace" . }}
        {{- $apisecret := (lookup "v1" "Secret" $detectedNamespace "keptn-api-token") }}

        {{- if $apisecret }}
            {{- b64dec (index $apisecret.data "keptn-api-token") }}
        {{- else }}
            {{- fail "Please provide an api token" }}
        {{- end }}
    {{- else if eq ((.Values.remoteControlPlane).api.authMode) "token" }}
         {{- required "A valid API Token is required!" .Values.remoteControlPlane.api.token }}
    {{- else }}
         {{- .Values.remoteControlPlane.api.token }}
    {{- end }}
{{- end }}
