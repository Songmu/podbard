package cast

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

type Index struct {
	RawFrontmatter, RawBody string

	Body string
}

func LoadIndex(rootDir string, cfg *Config, episodes []*Episode) (*Index, error) {
	idxMD := filepath.Join(rootDir, "index.md")

	if _, err := os.Stat(idxMD); err != nil {
		if os.IsNotExist(err) {
			return &Index{}, nil
		}
		return nil, err
	}
	bs, err := os.ReadFile(idxMD)
	if err != nil {
		return nil, err
	}
	content := strings.ReplaceAll(strings.TrimSpace(string(bs)), "\r\n", "\n")

	var idx *Index
	if !strings.HasPrefix(content, "---\n") {
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
