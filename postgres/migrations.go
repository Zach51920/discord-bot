package postgres

import (
	"errors"
	migrator "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log/slog"
	"os"
)

const sourceUrl = "file://migrations"

func RunMigrations() {
	databaseUrl := os.Getenv("MIGRATION_DB_URL")
	if databaseUrl == "" {
		slog.Warn("Unable to run migrations", "error", "databaseUrl is not set")
		return
	}

	migrate, err := migrator.New(sourceUrl, databaseUrl)
	if err != nil {
		slog.Warn("Error loading migrations", "error", err)
		return
	}
	defer migrate.Close()

	if err = migrate.Up(); err != nil && !errors.Is(err, migrator.ErrNoChange) {
		slog.Warn("Error applying migrations", "error", err)
		return
	}

	version, isDirty, _ := migrate.Version()
	if isDirty {
		slog.Warn("Migrator is in a dirty state", "version", version)
		return
	}
	slog.Info("Migrations executed successfully", "version", version)
}
