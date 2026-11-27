// Package templates
package templates

import (
	"embed"
	"fmt"
	"html/template"
)

//go:embed *
var files embed.FS

func NewTemplates() (*template.Template, error) {
	tmpl, err := template.ParseFS(files, "*.html")
	if err != nil {
		return nil, fmt.Errorf("parse email templates: %w", err)
	}
	return tmpl, nil
}
