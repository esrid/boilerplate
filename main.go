// Package main is the entry point of the application.
package main

import (
	"boilerplate/config"
	"boilerplate/config/db"
	"boilerplate/internal/server"
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	// Setup logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.New()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Construct DSN
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSslmode)

	db.RunMigrations(dsn)

	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		slog.Error("Unable to connect to database", "error", err)
		os.Exit(1)
	}

	defer conn.Close(context.Background())

	slog.Info("Successfully connected to database")

	router := server.NewRouter(conn)
	server.Start(router)
}
