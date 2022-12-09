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

	if err := createPipeline(tx, p); err != nil {
		tx.Rollback()
		return err
	}
	if err := createPipelineSteps(tx, p.ID, p.Steps); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s *PSQLStore) UpdatePipeline(p *service.Pipeline) error {
	tx, err := s.db.BeginTxx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`UPDATE "pipelines" SET "description" = $1, "status" = $2 WHERE "id" = $3`, p.Description, p.Status, p.ID); err != nil {
		return tx.Rollback()
	}

	if _, err := tx.Exec(`DELETE FROM "pipeline_steps" WHERE "pipeline_id" = $1`, p.ID); err != nil {
		tx.Rollback()
		return err
	}

	if err := createPipelineSteps(tx, p.ID, p.Steps); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s *PSQLStore) ListPipelines(filter service.PipelinesFilter) ([]service.Pipeline, error) {
	//
	// Fetch pipelines
	// Create query
	q := pq.Select("id", "description", "status").From("pipelines")
	if filter.OnlyInactive {
		q = q.Where(sq.Eq{"status": service.PipelineInactive})
	} else {
		q = q.Where(sq.NotEq{"status": service.PipelineInactive})
	}
	query, params, err := q.ToSql()
	if err != nil {
		return nil, err
	}
	// Perform query
	row, err := s.db.Queryx(query, params...)
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
		if err := row.Scan(&p.ID, &p.Description, &p.Status); err != nil {
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
	// Build query
	query, params, err = pq.
		Select("pipeline_id", "image").
		From("pipeline_steps").
		Where(sq.Eq{"pipeline_id": pIDs}).
		OrderBy("pipeline_step ASC").
		ToSql()
	if err != nil {
		return nil, err
	}
	// Perform query
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
	return getPipeline(s.db, id)
}

// Private methods which have DB interface injected. Allows for transactional queries
func getPipeline(db DB, id string) (*service.Pipeline, error) {
	var p service.Pipeline
	if err := db.QueryRowx(`SELECT id, description, status FROM pipelines WHERE id=$1`, id).Scan(&p.ID, &p.Description, &p.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrPipelineNotFound
		}
		return nil, err
	}
	p.Steps = []string{}

	if err := db.Select(
		&p.Steps,
		`SELECT image FROM pipeline_steps WHERE pipeline_id=$1 ORDER BY pipeline_step ASC`,
		id,
	); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return &p, nil
}

func createPipeline(db DB, p *service.Pipeline) error {
	if _, err := db.Exec(`INSERT INTO "pipelines" ("id", "description", "status") VALUES ($1, $2, $3)`, p.ID, p.Description, p.Status); err != nil {
		return err
	}

	return nil
}

func createPipelineSteps(db DB, id string, steps []string) error {
	if len(steps) == 0 {
		return nil
	}

	q := pq.Insert("pipeline_steps").Columns("pipeline_id", "pipeline_step", "image")
	for step, image := range steps {
		q = q.Values(id, step, image)
	}
	query, params, err := q.ToSql()
	if err != nil {
		return err
	}

	if _, err := db.Exec(query, params...); err != nil {
		return err
	}

	return nil
}
