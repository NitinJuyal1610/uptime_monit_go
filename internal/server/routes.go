package server

import (
	"net/http"
	handler "nitinjuyal1610/uptimeMonitor/internal/handlers"
	templates "nitinjuyal1610/uptimeMonitor/web"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) SetupRoutes() http.Handler {
	//chi routes and middlewares
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	// initialize handlers
	templateManager, _ := templates.NewManager()

	clientHandler := handler.NewClientHandler()
	urlHandler := handler.NewUrlHandler(s.Services.UrlService, templateManager)

	//routes
	r.Get("/", clientHandler.RenderDashboard)
	r.Post("/api/monitors", urlHandler.CreateURLMonitor)
	r.Get("/api/monitors", urlHandler.GetURLMonitors)
	r.Get("/api/monitors/{id}", urlHandler.GetMonitorById)

	return r
}
