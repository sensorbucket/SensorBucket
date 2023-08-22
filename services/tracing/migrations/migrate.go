package migrations

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.sql
var fs embed.FS

func MigratePostgres(db *sql.DB) error {
	psql, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create postgres instance for migrations: %w", err)
	}

	src, err := iofs.New(fs, ".")
	if err != nil {
		return fmt.Errorf("could not create source for migrations: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", psql)
	if err != nil {
		return fmt.Errorf("could not create migrator instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("could not migrate: %w", err)
	}

	return nil
}
