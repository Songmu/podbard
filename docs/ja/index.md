# {{.Channel.Title}}

{{.Channel.Description}}

## RSS Feed
<input type="text" value="{{.Channel.FeedURL}}" size="80" readonly>

## Episodes
{{range .Episodes -}}
- [{{.Title}}]({{.URL.Path}}) ({{.PubDate.Format "2006-01-02 15:04"}})
{{end}}
