package cast

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

type Page struct {
	RawFrontmatter, RawBody string

	Body string
}

func LoadPage(mdPath string, cfg *Config, episodes []*Episode) (*Page, error) {
	if _, err := os.Stat(mdPath); err != nil {
		if os.IsNotExist(err) {
			return &Page{}, nil
		}
		return nil, err
	}
	bs, err := os.ReadFile(mdPath)
	if err != nil {
		return nil, err
	}
	content := strings.ReplaceAll(strings.TrimSpace(string(bs)), "\r\n", "\n")

	var idx *Page
	if !strings.HasPrefix(content, "---\n") {
		idx = &Page{RawBody: content}
	} else {
		frontmater, body, err := splitFrontMatterAndBody(content)
		if err != nil {
			return nil, err
		}
		idx = &Page{
			RawFrontmatter: frontmater,
			RawBody:        body,
		}
	}
	if err := idx.build(cfg, episodes); err != nil {
		return nil, err
	}
	return idx, nil
}

func LoadIndex(rootDir string, cfg *Config, episodes []*Episode) (*Page, error) {
	idxMD := filepath.Join(rootDir, "index.md")
	return LoadPage(idxMD, cfg, episodes)
}

func (idx *Page) build(cfg *Config, episodes []*Episode) error {
	tmpl, err := template.New("index").Funcs(sprig.FuncMap()).
		Funcs(template.FuncMap{"html": htmlFunc}).
		Parse(idx.RawBody)
	if err != nil {
		return err
	}
	arg := struct {
		Channel  *ChannelConfig
		Episodes []*Episode
	}{
		Channel:  cfg.Channel,
		Episodes: episodes,
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, arg); err != nil {
		return err
	}

	md := NewMarkdown()
	var mdBuf bytes.Buffer
	if err := md.Convert(buf.Bytes(), &mdBuf); err != nil {
		return err
	}
	idx.Body = mdBuf.String()
	return nil
}
