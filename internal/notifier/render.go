package notifier

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
)

//go:embed templates/*.html
var templateFS embed.FS

type renderer struct {
	templates *template.Template
}

func newRenderer() (*renderer, error) {
	t, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse email templates: %w", err)
	}
	return &renderer{templates: t}, nil
}

func (r *renderer) render(name string, data any) (string, error) {
	var buf bytes.Buffer
	if err := r.templates.ExecuteTemplate(&buf, name+".html", data); err != nil {
		return "", fmt.Errorf("render template %q: %w", name, err)
	}
	return buf.String(), nil
}
