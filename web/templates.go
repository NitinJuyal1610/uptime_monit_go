package templates

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"reflect"
)

//go:embed templates/*.html templates/*/*.html templates/*/*/*.html

var templateFS embed.FS

type TemplateManager struct {
	templates *template.Template
}

func dict(values ...any) (map[string]any, error) {
	if len(values)%2 != 0 {
		return nil, errors.New("invalid dict call")
	}
	dict := make(map[string]any, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, errors.New("dict keys must be strings")
		}
		dict[key] = values[i+1]
	}
	return dict, nil
}

func add(a, b int) int {
	return a + b
}

func mul(a, b int) int {
	return a * b
}

// mod returns the remainder of a divided by b
func mod(a, b int) int {
	return a % b
}

// len returns the length of a slice, array, map, or string
func len(v any) int {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array ||
		rv.Kind() == reflect.Map || rv.Kind() == reflect.String {
		return rv.Len()
	}
	return 0
}

func NewManager() (*TemplateManager, error) {
	tmpl := template.New("root").Funcs(template.FuncMap{
		"dict": dict,
		"add":  add,
		"mod":  mod,
		"mul":  mul,
	})
	tmpl, err := tmpl.ParseFS(templateFS, "templates/*.html", "templates/*/*.html", "templates/*/*/*.html")

	// for _, val := range tmpl.Templates() {
	// 	fmt.Println(val.Name())
	// }
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
		http.Error(w, "Error rendering template \n", http.StatusInternalServerError)
		fmt.Printf("Template render error (%s): %v \n", name, err)
	}
	return err
}
