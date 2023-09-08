package main

import (
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

func newPSQLStore(db *sqlx.DB) *PSQLStore {
	return &PSQLStore{db}
}

func (s *PSQLStore) WorkersExists(ids []uuid.UUID) ([]uuid.UUID, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	q := sq.Select("id").From("user_workers").Where(sq.Eq{"id": ids})
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
		"id", "language", "organisation", "major", "revision", "status", "status_info",
		"source", "entrypoint",
	).
		From("user_workers")
	cursor := pagination.GetCursor[UserWorkerPaginationQuery](req)
	q, err = pagination.Apply(q, cursor)
	if err != nil {
		return nil, fmt.Errorf("could not apply pagination: %w", err)
	}

	rows, err := q.RunWith(s.db).Query()
	if err != nil {
		return nil, fmt.Errorf("error querying rows: %w", err)
	}

	workers := []UserWorker{}
	for rows.Next() {
		var worker UserWorker
		if err := rows.Scan(
			&worker.ID, &worker.Language, &worker.Organisation, &worker.Major, &worker.Revision,
			&worker.Status, &worker.StatusInfo, &worker.Source, &worker.Entrypoint, &cursor.Columns.ID,
		); err != nil {
			return nil, fmt.Errorf("error scanning worker from database: %w", err)
		}
		workers = append(workers, worker)
	}

	page := pagination.CreatePageT(workers, cursor)
	return &page, nil
}
