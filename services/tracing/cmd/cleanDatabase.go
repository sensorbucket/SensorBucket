package cmd

import (
	"fmt"
	"time"

	"github.com/urfave/cli/v2"
	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/services/tracing/tracing"
)

func cmdCleanDatabase(cmd *cli.Context) error {
	DB_DSN := env.Must("DB_DSN")
	db, err := createDB(DB_DSN)
	if err != nil {
		return fmt.Errorf("could not create database connection: %w", err)
	}

	svc := tracing.Create(db)
	if err := svc.PeriodicCleanup(time.Duration(cmd.Float64("days")) * time.Hour * 24); err != nil {
		return fmt.Errorf("could not perform database cleanup: %w", err)
	}

	if err := db.Close(); err != nil {
		return fmt.Errorf("could not properly close database: %w", err)
	}

	return nil
}
