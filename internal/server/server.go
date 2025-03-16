package server

import (
	"fmt"
	"net/http"
	"nitinjuyal1610/uptimeMonitor/internal/config"
	service "nitinjuyal1610/uptimeMonitor/internal/services"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
)

// server returns a new server
type Server struct {
	port     int
	db       *pgxpool.Pool
	Services *service.Services
}

func New() (*Server, *http.Server) {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	db := config.NewPostgresConnection()
	//services init
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
