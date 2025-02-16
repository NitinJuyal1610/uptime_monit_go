package server

import (
	"net/http"
	handler "nitinjuyal1610/uptimeMonitor/internal/handlers"
	"nitinjuyal1610/uptimeMonitor/internal/repository"
	service "nitinjuyal1610/uptimeMonitor/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) SetupRoutes() http.Handler {
	//chi routes and middlewares
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	//server has now database and port

	//initialize repositories
	urlRepository := repository.NewUrlRepository(s.db)
	//initialize service
	urlService := service.NewUrlService(urlRepository)
	// initialize handlers
	generalHandler := handler.NewGeneralHandler()
	urlHandler := handler.NewUrlHandler(urlService)

	//routes
	r.Get("/", generalHandler.HelloWorld)
	r.Post("/monitors", urlHandler.CreateURLMonitor)
	r.Get("/monitors", urlHandler.GetURLMonitors)
	r.Get("/monitors/{id}", urlHandler.GetMonitorById)

	return r
}
