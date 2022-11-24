package store

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"sensorbucket.nl/sensorbucket/services/pipeline/service"
)

var pq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

var _ service.Store = (*PSQLStore)(nil)

type PSQLStore struct {
	db *sqlx.DB
}

func NewPSQLStore(db *sqlx.DB) *PSQLStore {
	return &PSQLStore{db}
}

func (s *PSQLStore) CreatePipeline(p *service.Pipeline) error {
	tx, err := s.db.BeginTxx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`INSERT INTO "pipelines" ("id", "description") VALUES ($1, $2)`, p.ID, p.Description); err != nil {
		return tx.Rollback()
	}

	q := pq.Insert("pipeline_steps").Columns("pipeline_id", "pipeline_step", "image")
	for step, image := range p.Steps {
		q = q.Values(p.ID, step, image)
	}
	query, params, err := q.ToSql()
	if err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.Exec(query, params...); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s *PSQLStore) GetPipeline(id string) (*service.Pipeline, error) {
	var p service.Pipeline
	if err := s.db.QueryRow(`SELECT id, description FROM pipelines WHERE id=$1`, id).Scan(&p.ID, &p.Description); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrPipelineNotFound
		}
		return nil, err
	}
	p.Steps = make([]string, 0)

	if err := s.db.Select(
		&p.Steps,
		`SELECT image FROM pipeline_steps WHERE pipeline_id=$1 ORDER BY pipeline_step ASC`,
		id,
	); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return &p, nil
}
