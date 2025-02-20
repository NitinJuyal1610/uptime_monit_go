package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

type ClientHandler struct {
	//------TODO
}

func NewClientHandler() *ClientHandler {
	return &ClientHandler{}
}

func (gh *ClientHandler) RenderDashboard(w http.ResponseWriter, r *http.Request) {
	//------server html
	templatePath, err := filepath.Abs("web/templates/index.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Error resolving template path:", err)
		return
	}

	// Parse template
	temp, err := template.ParseFiles(templatePath)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Error parsing template:", err)
		return
	}

	// Set Content-Type header
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Execute template
	err = temp.Execute(w, nil)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Error executing template:", err)
	}
}
