package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

func NewPostgresConnection(connStr string,dbPort int) *pgxpool.Pool {
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


func RunMigrations(databaseURL string) error {

	migrationURL := "file://migrations"
	m, err := migrate.New(migrationURL, databaseURL)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run migrations: %v", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}