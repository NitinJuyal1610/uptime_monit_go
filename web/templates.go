package templates

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
)

//go:embed templates/**/*.html
var templateFS embed.FS

type TemplateManager struct {
	templates *template.Template
}

func NewManager() (*TemplateManager, error) {
	tmpl, err := template.ParseFS(templateFS, "templates/**/*.html")
	if err != nil {
		fmt.Printf("Error parsing templates: %v", err)
		return nil, err
	}

	return &TemplateManager{templates: tmpl}, nil
}

func (m *TemplateManager) Render(w http.ResponseWriter, name string, data interface{}) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := m.templates.ExecuteTemplate(w, name, data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		fmt.Printf("Template render error (%s): %v", name, err)
	}
	return err
}
