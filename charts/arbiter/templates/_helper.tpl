
{{- define "kubeVersion" -}}
  {{- default .Capabilities.KubeVersion.Major .Values.kubeVersionOverride -}}
{{- end -}}


{{- define "le20" -}}
  {{- (semverCompare "< v1.21" (include "kubeVersion" .)) -}}
{{- end -}}

