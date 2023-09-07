package main

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

type dbStore struct {
	sensorbucketDb *sqlx.DB
	tracingDb      *sqlx.DB
}

func (s *dbStore) DeleteExpiredData() error {
	// Delete the tracing information linked to archive ingress dto's that are expired
	stepsDeleted, err := exec(s.tracingDb, `DELETE FROM steps
	WHERE tracing_id IN (
		SELECT DISTINCT s.tracing_id
		FROM steps s
		JOIN archived_ingress_dtos a ON s.tracing_id = a.tracing_id
		WHERE a.expires_at <= now()
	);`) // TODO: what timezones are regular timestamps stored in?
	if err != nil {
		return fmt.Errorf("delete tracing steps: %w", err)
	}
	log.Printf("Deleted %d steps from tracing database\n", stepsDeleted)

	// Delete archive dto's
	ingressDeleted, err := exec(s.tracingDb, `DELETE FROM archived_ingress_dtos WHERE expires_at <= now()`)
	if err != nil {
		return fmt.Errorf("delete archived ingress: %w", err)
	}
	log.Printf("Deleted %d archived ingress dtos from tracing database", ingressDeleted)

	// Finally delete the measurements
	measurementsDeleted, err := exec(s.sensorbucketDb, `DELETE FROM measurements WHERE measurement_expiration <= now()`)
	if err != nil {
		return fmt.Errorf("delete measurements: %w", err)
	}
	log.Printf("Deleted %d measurements from sensorbucket database", measurementsDeleted)
	return nil
}

func exec(db *sqlx.DB, q string) (int64, error) {
	res, err := db.Exec(q)
	if err != nil {
		return 0, fmt.Errorf("exec: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected: %w", err)
	}
	return affected, nil
}
