package cast

import (
	"bytes"
	"html/template"
	"path/filepath"
)

const templateDir = "tmpl"

type castTemplate struct {
	*template.Template
}

func loadTemplate(rootDir string) (*castTemplate, error) {
	base := filepath.Join(rootDir, templateDir)
	glob := filepath.Join(base, "*.tmpl")

	tmpl, err := template.ParseGlob(glob)
	if err != nil {
		return nil, err
	}
	return &castTemplate{tmpl}, nil
}

func (ct *castTemplate) execute(layout, name string, data interface{}) (string, error) {
	var buf bytes.Buffer

	template.Must(template.Must(
		ct.Lookup(layout).Clone()).
		AddParseTree("content", ct.Lookup(name).Tree)).
		ExecuteTemplate(&buf, layout, data)

	return buf.String(), nil
}
