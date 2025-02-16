package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"nitinjuyal1610/uptimeMonitor/internal/config"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

// server returns a new server
type Server struct {
	port int
	db   *sql.DB
}

func New() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	NewServer := &Server{
		port: port,
		db:   config.InitDatabase(),
	}
	//----
	s := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.SetupRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	return s
}
