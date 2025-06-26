package config

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
)

func NewDatabase(url string) *pgxpool.Pool {
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		log.Fatalf("unable to parse database URL: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("unable to create database connection pool: %v", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		log.Fatalf("unable to ping database: %v", err)
	}

	log.Println("Database connection pool successfully established.")

	err = runMigrations(pool)
	if err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	return pool
}

//go:embed migrations/*.sql
var embedMigrations embed.FS

func runMigrations(pool *pgxpool.Pool) error {
	url := pool.Config().ConnString()

	db, err := sql.Open("pgx", url)
	if err != nil {
		return fmt.Errorf("unable to open sql.DB for migrations: %w", err)
	}
	defer db.Close()

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("unable to set dialect for goose migrations: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("unable to accomplish goose migrations: %w", err)
	}

	log.Println("Database migrations applied successfully.")
	return nil
}
