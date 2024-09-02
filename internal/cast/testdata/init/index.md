# {{.Channel.Title}}

{{.Channel.Description}}

{{range .Episodes}}
- [{{.Title}}]({{.URL}}) ({{.PubDate.Format "2006-01-02"}})
{{end}}
