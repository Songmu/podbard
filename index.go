package primcast

import (
	"bytes"
	"os"
	"strings"
	"text/template"
)

const indexMD = "index.md"

type Index struct {
	RawFrontmatter, RawBody string

	Body string
}

func LoadIndex(cfg *Config, episodes []*Episode) (*Index, error) {
	if _, err := os.Stat(indexMD); err != nil {
		if os.IsNotExist(err) {
			return &Index{}, nil
		}
		return nil, err
	}
	bs, err := os.ReadFile(indexMD)
	if err != nil {
		return nil, err
	}
	content := strings.TrimSpace(string(bs))

	var idx *Index
	if !strings.HasPrefix(content, "---\n") { // XXX: windows
		idx = &Index{RawBody: content}
	} else {
		frontmater, body, err := splitFrontMatterAndBody(content)
		if err != nil {
			return nil, err
		}
		idx = &Index{
			RawFrontmatter: frontmater,
			RawBody:        body,
		}
	}
	if err := idx.build(cfg, episodes); err != nil {
		return nil, err
	}
	return idx, nil
}

func (idx *Index) build(cfg *Config, episodes []*Episode) error {
	tmpl, err := template.New("index").Parse(idx.RawBody)
	if err != nil {
		return err
	}
	arg := struct {
		Config   *Config
		Episodes []*Episode
	}{
		Config:   cfg,
		Episodes: episodes,
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, arg); err != nil {
		return err
	}

	md := NewMarkdown()
	var mdBuf bytes.Buffer
	if err := md.Convert([]byte(buf.String()), &mdBuf); err != nil {
		return err
	}
	idx.Body = mdBuf.String()
	return nil
}
