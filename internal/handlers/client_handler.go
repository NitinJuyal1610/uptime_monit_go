package handler

import (
	"net/http"
	templates "nitinjuyal1610/uptimeMonitor/web"
)

type ClientHandler struct {
	tempateManager *templates.TemplateManager
}

func NewClientHandler(tm *templates.TemplateManager) *ClientHandler {
	return &ClientHandler{
		tempateManager: tm,
	}
}

func (gh *ClientHandler) RenderDashboard(w http.ResponseWriter, r *http.Request) {
	gh.tempateManager.Render(w, "index.html", map[string]any{})
}

func (gh *ClientHandler) RenderLogin(w http.ResponseWriter, r *http.Request) {
	gh.tempateManager.Render(w, "auth.html", map[string]any{})
}
