package database

import (
	"embed"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/your-org/atlas/backend/internal/config"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunMigrations применяет все pending-миграции из embedded файлов.
func RunMigrations(cfg config.DatabaseConfig) error {
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("migrations source: %w", err)
	}

	dsn := fmt.Sprintf(
		"pgx5://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.SSLMode,
	)

	m, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	if err != nil {
		return fmt.Errorf("migrate init: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("Migrations: no changes")
			return nil
		}
		return fmt.Errorf("migrate up: %w", err)
	}

	version, _, _ := m.Version()
	log.Printf("Migrations: applied successfully, version=%d", version)
	return nil
}
