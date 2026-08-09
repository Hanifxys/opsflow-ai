package database

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations executes SQL migrations from a source URL against database DSN.
func RunMigrations(migrationsPath, dsn string) error {
	m, err := migrate.New("file://"+migrationsPath, dsn)
	if err != nil {
		return fmt.Errorf("failed to init migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run up migrations: %w", err)
	}

	return nil
}
