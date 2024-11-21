package cast

import (
	"html/template"
	"io"
	"path/filepath"

	"github.com/Masterminds/sprig/v3"
)

const templateDir = "template"

type castTemplate struct {
	*template.Template
}

// XXX: define argument types

func loadTemplate(rootDir string) (*castTemplate, error) {
	base := filepath.Join(rootDir, templateDir)
	glob := filepath.Join(base, "*.tmpl")

	tmpl, err := template.ParseGlob(glob)
	if err != nil {
		return nil, err
	}
	return &castTemplate{tmpl}, nil
}

func htmlFunc(html string) template.HTML {
	return template.HTML(html)
}

func (ct *castTemplate) Execute(w io.Writer, layout, name string, data interface{}) error {
	return template.Must(template.Must(
		ct.Lookup(layout).Clone()).
		Funcs(sprig.FuncMap()).
		Funcs(template.FuncMap{"html": htmlFunc}).
		AddParseTree("content", ct.Lookup(name).Tree)).
		ExecuteTemplate(w, layout, data)
}
