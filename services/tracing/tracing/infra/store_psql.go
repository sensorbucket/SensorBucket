package tracinginfra

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"sensorbucket.nl/sensorbucket/services/tracing/tracing"
)

func NewStorePSQL(db *sqlx.DB) *stepStore {
	return &stepStore{
		db: db,
	}
}

func (s *stepStore) Insert(step tracing.Step) error {
	if _, err := s.db.Exec(
		`INSERT INTO "steps" ("tracing_id", "step_index", "steps_remaining", "start_time", "error") VALUES ($1, $2, $3, $4, $5)`,
		step.TracingID,
		step.StepIndex,
		step.StepsRemaining,
		step.StartTime,
		step.Error); err != nil {
		return err
	}

	return nil
}

type stepStore struct {
	db *sqlx.DB
}
