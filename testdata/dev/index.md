# {{.Channel.Title}}

{{.Channel.Description}}

Feed: <input type="text" value="{{.Channel.FeedURL}}" readonly>

## Episodes
{{range .Episodes -}}
- [{{.Title}}]({{.URL.Path}}) ({{.PubDate.Format "2006-01-02"}})
{{end}}
