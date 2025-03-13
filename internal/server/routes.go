package server

import (
	"net/http"
	handler "nitinjuyal1610/uptimeMonitor/internal/handlers"
	customMiddleware "nitinjuyal1610/uptimeMonitor/internal/middlewares"
	"nitinjuyal1610/uptimeMonitor/internal/session"
	"os"

	templates "nitinjuyal1610/uptimeMonitor/web"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) SetupRoutes() http.Handler {
	//chi routes and middlewaresclear
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	// initialize MANAGERs
	templateManager, _ := templates.NewManager()
	sessionManager := session.NewSessionManager(os.Getenv("SESSION_SECRET"))

	//custom middlewares
	authMiddleware := customMiddleware.NewAuthMiddleware(sessionManager)
	//initiate handlers
	clientHandler := handler.NewClientHandler(templateManager)
	urlHandler := handler.NewUrlHandler(s.Services.UrlService, templateManager)
	statHandler := handler.NewStatHandler(s.Services.StatService, templateManager)
	authHandler := handler.NewAuthHandler(s.Services.AuthService, templateManager, sessionManager)
	//routes
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	r.Route("/", func(r chi.Router) {
		r.Use(authMiddleware.Authenticate)

		r.Get("/", clientHandler.RenderDashboard)
		// monitor routes fine
		r.Post("/api/monitors", urlHandler.CreateURLMonitor)
		r.Get("/api/monitors", urlHandler.GetURLMonitors)
		r.Get("/api/monitors/{id}", urlHandler.GetMonitorById)

		// stats route fine
		r.Get("/api/monitors/{id}/stats", statHandler.GetMonitorStats)
		r.Get("/api/monitors/{id}/avg-response-graph", statHandler.GetAvgResponseGraph)
		r.Get("/api/monitors/{id}/stats", statHandler.GetMonitorStats)
		r.Get("/api/monitors/{id}/uptime-graph", statHandler.GetUptimeGraph)
		r.Get("/api/monitors/{id}/detailed-time-graph", statHandler.GetDetailedTimeGraph)

		//
		r.Post("/api/auth/logout", authHandler.Logout)
	})

	r.Get("/login", clientHandler.RenderLogin)

	//now auth routes
	r.Post("/api/auth/login", authHandler.Login)
	r.Post("/api/auth/signup", authHandler.SignUp)

	return r
}
