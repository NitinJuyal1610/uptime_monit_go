package server

import (
	"net/http"
	handler "nitinjuyal1610/uptimeMonitor/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) SetupRoutes() http.Handler {
	//chi routes and middlewares
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	// initialize handlers
	generalHandler := handler.NewGeneralHandler()
	urlHandler := handler.NewUrlHandler(s.Services.UrlService)

	//routes
	r.Get("/", generalHandler.HelloWorld)
	r.Post("/monitors", urlHandler.CreateURLMonitor)
	r.Get("/monitors", urlHandler.GetURLMonitors)
	r.Get("/monitors/{id}", urlHandler.GetMonitorById)

	return r
}
