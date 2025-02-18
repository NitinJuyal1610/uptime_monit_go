package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"nitinjuyal1610/uptimeMonitor/internal/config"
	service "nitinjuyal1610/uptimeMonitor/internal/services"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

// server returns a new server
type Server struct {
	port     int
	db       *sql.DB
	Services *service.Services
}

func New() (*Server, *http.Server) {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	db := config.InitDatabase()
	//services
	svcs := service.NewServices(db)
	srv := &Server{
		port:     port,
		db:       db,
		Services: svcs,
	}
	//----
	s := &http.Server{
		Addr:         fmt.Sprintf(":%d", srv.port),
		Handler:      srv.SetupRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	return srv, s
}
