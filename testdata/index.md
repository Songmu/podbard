# {{.Config.Channel.Title}}

{{.Config.Channel.Description}}

{{range .Episodes}}
## [{{.Title}}]({{.URL}})

{{.Description}}
{{end}}
