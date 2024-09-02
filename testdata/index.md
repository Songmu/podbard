# {{.Channel.Title}}

{{.Channel.Description}}

{{range .Episodes}}
## [{{.Title}}]({{.URL}})

{{.Description}}
{{end}}
