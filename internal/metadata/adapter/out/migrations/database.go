package migrations

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrationsFromFile(dbAddr, migrationsURL string, log *slog.Logger) error {
	log.Info("running database migrations")

	m, err := migrate.New(migrationsURL, dbAddr)
	if err != nil {
		return fmt.Errorf("failed to init migrate instance: %w", err)
	}
	defer m.Close()

	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Info("database is up to date, no migrations applied")
			return nil
		}
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Info("database migrations applied successfully")
	return nil
}
