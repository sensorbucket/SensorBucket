package userworkers

import (
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"sensorbucket.nl/sensorbucket/internal/pagination"
)

var _ Store = (*PSQLStore)(nil)

type PSQLStore struct {
	db *sqlx.DB
}

func NewPSQLStore(db *sqlx.DB) *PSQLStore {
	return &PSQLStore{db}
}

func (s *PSQLStore) WorkersExists(ids []uuid.UUID) ([]uuid.UUID, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	q := sq.Select("id").From("user_workers").Where(sq.Eq{"id": ids, "state": StateEnabled})
	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(s.db).Query()
	if err != nil {
		return nil, fmt.Errorf("error querying for worker ids: %w", err)
	}
	existingIDs := make([]uuid.UUID, 0, len(ids))
	for rows.Next() {
		var id uuid.UUID
		err := rows.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("error scanning ID: %w", err)
		}
		existingIDs = append(existingIDs, id)
	}
	return existingIDs, nil
}

type UserWorkerPaginationQuery struct {
	ID uuid.UUID `pagination:"id,ASC"`
}

func (s *PSQLStore) ListUserWorkers(req pagination.Request) (*pagination.Page[UserWorker], error) {
	var err error
	q := sq.Select(
		"id", "name", "description", "state", "language", "organisation", "major", "revision",
		"status", "status_info", "source", "entrypoint",
	).
		From("user_workers").Where(sq.Eq{"state": StateEnabled})

	cursor, err := pagination.GetCursor[UserWorkerPaginationQuery](req)
	if err != nil {
		return nil, fmt.Errorf("could not get pagination cursor: %w", err)
	}
	q, err = pagination.Apply(q, cursor)
	if err != nil {
		return nil, fmt.Errorf("could not apply pagination: %w", err)
	}

	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(s.db).Query()
	if err != nil {
		return nil, fmt.Errorf("error querying rows: %w", err)
	}

	workers := []UserWorker{}
	for rows.Next() {
		var worker UserWorker
		if err := rows.Scan(
			&worker.ID, &worker.Name, &worker.Description, &worker.State, &worker.Language, &worker.Organisation,
			&worker.Major, &worker.Revision, &worker.Status, &worker.StatusInfo, &worker.ZipSource, &worker.Entrypoint,
			&cursor.Columns.ID,
		); err != nil {
			return nil, fmt.Errorf("error scanning worker from database: %w", err)
		}
		workers = append(workers, worker)
	}

	page := pagination.CreatePageT(workers, cursor)
	return &page, nil
}

func (s *PSQLStore) CreateWorker(worker *UserWorker) error {
	if _, err := s.db.Exec(
		`INSERT INTO user_workers (
            id, name, description, state, language, organisation, major, revision,
            status, source, entrypoint
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
        )`,
		worker.ID, worker.Name, worker.Description, worker.State, worker.Language, worker.Organisation,
		worker.Major, worker.Revision, worker.Status, worker.ZipSource, worker.Entrypoint,
	); err != nil {
		return fmt.Errorf("could not create worker in store: %w", err)
	}
	return nil
}

func (s *PSQLStore) GetWorkerByID(id uuid.UUID) (*UserWorker, error) {
	var worker UserWorker
	if err := s.db.Get(&worker, "SELECT * FROM user_workers WHERE id=$1", id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWorkerNotFound
		}
		return nil, err
	}
	return &worker, nil
}

func (s *PSQLStore) UpdateWorker(worker *UserWorker) error {
	if _, err := s.db.Exec(
		`UPDATE user_workers SET source=$1, revision=$2, state=$3, description=$4 WHERE id=$5`,
		worker.ZipSource, worker.Revision, worker.State, worker.Description, worker.ID,
	); err != nil {
		return fmt.Errorf("could not update worker in store: %w", err)
	}
	return nil
}
