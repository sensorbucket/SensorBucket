package processinginfra

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/services/core/processing"
)

var pq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

var _ processing.Store = (*PSQLStore)(nil)

type PSQLStore struct {
	db *sqlx.DB
}

func NewPSQLStore(db *sqlx.DB) *PSQLStore {
	return &PSQLStore{db}
}

func (s *PSQLStore) CreatePipeline(p *processing.Pipeline) error {
	tx, err := s.db.BeginTxx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	if err := createPipeline(tx, p); err != nil {
		if rb := tx.Rollback(); rb != nil {
			err = fmt.Errorf("rollback error %w while handling error: %w", rb, err)
		}
		return err
	}
	if err := createPipelineSteps(tx, p.ID, p.Steps); err != nil {
		if rb := tx.Rollback(); rb != nil {
			err = fmt.Errorf("rollback error %w while handling error: %w", rb, err)
		}
		return err
	}

	return tx.Commit()
}

func (s *PSQLStore) UpdatePipeline(p *processing.Pipeline) error {
	tx, err := s.db.BeginTxx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	if _, err := tx.Exec(
		`UPDATE "pipelines" SET "description" = $1, "status" = $2, "last_status_change" = $3 WHERE "id" = $4`,
		p.Description, p.Status, p.LastStatusChange, p.ID,
	); err != nil {
		if rb := tx.Rollback(); rb != nil {
			err = fmt.Errorf("rollback error %w while handling error: %w", rb, err)
		}
		return err
	}

	if _, err := tx.Exec(`DELETE FROM "pipeline_steps" WHERE "pipeline_id" = $1`, p.ID); err != nil {
		if rb := tx.Rollback(); rb != nil {
			err = fmt.Errorf("rollback error %w while handling error: %w", rb, err)
		}
		return err
	}

	if err := createPipelineSteps(tx, p.ID, p.Steps); err != nil {
		if rb := tx.Rollback(); rb != nil {
			err = fmt.Errorf("rollback error %w while handling error: %w", rb, err)
		}
		return err
	}

	return tx.Commit()
}

type pipelinePaginationQuery struct {
	CreatedAt time.Time `pagination:"created_at,ASC"`
	ID        string    `pagination:"id,ASC"`
}

func (s *PSQLStore) ListPipelines(filter processing.PipelinesFilter, p pagination.Request) (pagination.Page[processing.Pipeline], error) {
	var page pagination.Page[processing.Pipeline]
	var err error
	// Fetch pipelines
	q := pq.Select("id", "description", "status", "last_status_change", "created_at").From("pipelines")
	if len(filter.Status) > 0 {
		q = q.Where(sq.Eq{"status": filter.Status})
	} else {
		q = q.Where(sq.NotEq{"status": processing.PipelineInactive})
	}
	if len(filter.Step) > 0 {
		pipelineIDsThatHaveSteps := pq.Select("pipeline_id").Prefix("id IN (").Suffix(")").Distinct().From("pipeline_steps").Where(sq.Eq{"image": filter.Step})
		q = q.Where(pipelineIDsThatHaveSteps)
	}
	if len(filter.ID) > 0 {
		q = q.Where(sq.Eq{"id": filter.ID})
	}

	// Pagination
	cursor, err := pagination.GetCursor[pipelinePaginationQuery](p)
	if err != nil {
		return page, fmt.Errorf("list pipelines, error getting pagination cursor: %w", err)
	}
	q, err = pagination.Apply(q, cursor)
	if err != nil {
		return page, err
	}

	query, params, err := q.ToSql()
	if err != nil {
		return page, err
	}
	// Perform query
	row, err := s.db.Queryx(query, params...)
	if err != nil {
		return page, err
	}
	defer row.Close()

	// Map rows to model
	pIDs := make([]string, 0)
	pipelines := make([]processing.Pipeline, 0)
	for row.Next() {
		p := processing.Pipeline{
			Steps: []string{},
		}
		if err := row.Scan(
			&p.ID, &p.Description, &p.Status, &p.LastStatusChange, &p.CreatedAt,
			&cursor.Columns.CreatedAt, &cursor.Columns.ID,
		); err != nil {
			return page, err
		}
		pIDs = append(pIDs, p.ID)
		pipelines = append(pipelines, p)
	}
	if len(pipelines) == 0 {
		return page, nil
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
		return page, err
	}
	// Perform query
	row, err = s.db.Queryx(query, params...)
	if err != nil {
		return page, err
	}
	defer row.Close()

	// Map steps to pipeline
	stepMap := make(map[string][]string)
	for row.Next() {
		var pID string
		var pStep string
		if err := row.Scan(&pID, &pStep); err != nil {
			return page, err
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

	// Create pagination page
	page = pagination.CreatePageT(pipelines, cursor)
	return page, nil
}

func (s *PSQLStore) GetPipeline(id string) (*processing.Pipeline, error) {
	return getPipeline(s.db, id)
}

// Private methods which have DB interface injected. Allows for transactional queries
func getPipeline(db DB, id string) (*processing.Pipeline, error) {
	var p processing.Pipeline
	if err := db.QueryRowx(`SELECT id, description, status, last_status_change, created_at FROM pipelines WHERE id=$1`, id).Scan(&p.ID, &p.Description, &p.Status, &p.LastStatusChange, &p.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, processing.ErrPipelineNotFound
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

func createPipeline(db DB, p *processing.Pipeline) error {
	if _, err := db.Exec(`INSERT INTO "pipelines" ("id", "description", "status", "last_status_change", "created_at") VALUES ($1, $2, $3, $4, $5)`, p.ID, p.Description, p.Status, p.LastStatusChange, p.CreatedAt); err != nil {
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
