package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func NewPostgresConnection() *pgxpool.Pool {

	port, _ := strconv.Atoi(os.Getenv("DB_PORT"))

	config := &PostgresConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     port,
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  "disable",
	}
	//
	connStr := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	poolConfig, err := pgxpool.ParseConfig(connStr)

	if err != nil {
		log.Fatalf("Unable to parse database config: %v", err)
	}

	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = time.Minute * 30
	pgpool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)

	if err != nil {
		log.Fatalf("Unable to create connection pool: %v \n", err)
	}

	if err := pgpool.Ping(context.Background()); err != nil {
		log.Fatalf("Cannot reach the database pool: %v", err)
	}

	fmt.Println("Successfully connected!")
	return pgpool
}

func CloseConnection(db *pgxpool.Pool) {
	db.Close()
}
