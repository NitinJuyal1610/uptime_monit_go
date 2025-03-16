package server

import (
	"fmt"
	"log"
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

	dbPort, _ := strconv.Atoi(os.Getenv("DB_PORT"))
	dbConfig := &config.PostgresConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     dbPort,
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  "disable",
	}
	connStr := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName, dbConfig.SSLMode)

	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.DBName, dbConfig.SSLMode)
	
	db := config.NewPostgresConnection(connStr, dbPort)

	if err:=config.RunMigrations(databaseURL);err!=nil{
		log.Fatalf("Migrations Failed: %v",err)
	}
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
