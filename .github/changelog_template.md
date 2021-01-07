{{define "GroupTemplate" -}}
{{- range .Grouped}}
### {{ .Name }}

{{range .Items -}}
* [{{.CommitHashShort}}]({{.CommitURL}}) {{.Title}} ({{if .IsPull}}[contributed]({{.PullURL}}) by {{end}}[{{.Author}}]({{.AuthorURL}}))
{{end -}}
{{end -}}
{{end -}}
{{define "FlatTemplate" -}}
{{range .Items -}}
* [{{.CommitHashShort}}]({{.CommitURL}}) {{.Title}} ({{if .IsPull}}[contributed]({{.PullURL}}) by {{end}}[{{.Author}}]({{.AuthorURL}}))
{{end -}}
{{end -}}
{{define "DefaultTemplate" -}}
## Release Notes {{.Version}}
{{if len .Grouped -}}
{{template "GroupTemplate" . -}}   
{{- else}}
{{template "FlatTemplate" . -}}
{{end}}
{{end -}}
{{template "DefaultTemplate" . -}}