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

func (s *PSQLStore) UpdatePipeline(id string, p service.UpdatePipelineDTO) error {
	tx, err := s.db.BeginTxx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	// TODO: Implement checks if property should be updated. (IF p.Description != nil)
	if _, err := tx.Exec(`UPDATE "pipelines" SET "description" = $1 WHERE "id" = $2`, p.Description, id); err != nil {
		return tx.Rollback()
	}

	if _, err := tx.Exec(`DELETE FROM "pipeline_steps" WHERE "pipeline_id" = $1`, id); err != nil {
		return tx.Rollback()
	}

	q := pq.Insert("pipeline_steps").Columns("pipeline_id", "pipeline_step", "image")
	for step, image := range p.Steps {
		q = q.Values(id, step, image)
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

func (s *PSQLStore) ListPipelines() ([]service.Pipeline, error) {
	//
	// Fetch pipelines
	row, err := s.db.Queryx("SELECT id, description FROM pipelines")
	if err != nil {
		return nil, err
	}
	// Map rows to model
	pIDs := make([]string, 0)
	pipelines := make([]service.Pipeline, 0)
	for row.Next() {
		p := service.Pipeline{
			Steps: []string{},
		}
		if err := row.Scan(&p.ID, &p.Description); err != nil {
			return nil, err
		}
		pIDs = append(pIDs, p.ID)
		pipelines = append(pipelines, p)
	}
	if len(pipelines) == 0 {
		return pipelines, nil
	}

	//
	// Fetch steps
	query, params, _ := pq.
		Select("pipeline_id", "image").
		From("pipeline_steps").
		Where(sq.Eq{"pipeline_id": pIDs}).
		OrderBy("pipeline_step ASC").
		ToSql()
	row, err = s.db.Queryx(query, params...)
	if err != nil {
		return nil, err
	}
	// Map steps to pipeline
	stepMap := make(map[string][]string)
	for row.Next() {
		var pID string
		var pStep string
		if err := row.Scan(&pID, &pStep); err != nil {
			return nil, err
		}

		m, ok := stepMap[pID]
		if !ok {
			m = []string{}
		}
		m = append(m, pStep)
		stepMap[pID] = m
	}

	for ix := range pipelines {
		p := &pipelines[ix]
		steps, ok := stepMap[p.ID]
		if !ok {
			continue
		}
		p.Steps = steps
	}

	return pipelines, nil
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
