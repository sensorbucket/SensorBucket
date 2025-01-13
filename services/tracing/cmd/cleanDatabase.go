package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"sensorbucket.nl/sensorbucket/services/tracing/tracing"
)

func cmdCleanDatabase(cmd *cli.Context) error {
	db, err := createDB()
	if err != nil {
		return fmt.Errorf("could not create database connection: %w", err)
	}

	svc := tracing.Create(db)
	if err := svc.PeriodicCleanup(); err != nil {
		return fmt.Errorf("could not perform database cleanup: %w", err)
	}

	if err := db.Close(); err != nil {
		return fmt.Errorf("could not properly close database: %w", err)
	}

	return nil
}
