// Package db provides database configurations for the application.
package db

import (
	"embed"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(databaseURL string) {
	if databaseURL == "" {
		log.Fatal("databaseURL is not set")
	}

	sourceInstance, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		log.Fatalf("failed to create source instance: %v", err)
	}

	m, err := migrate.NewWithSourceInstance(
		"iofs",
		sourceInstance,
		databaseURL,
	)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	log.Println("Migrations applied successfully")
}
