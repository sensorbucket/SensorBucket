package main

import (
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

func newPSQLStore(db *sqlx.DB) *PSQLStore {
	return &PSQLStore{db}
}

func (s *PSQLStore) WorkersExists(ids []uuid.UUID) ([]uuid.UUID, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	q := sq.Select("id").From("workers").Where(sq.Eq{"id": ids})
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

func (s *PSQLStore) ListUserWorkers(req pagination.Request) (*pagination.Page[UserWorker], error) {
	return nil, errors.New("not implemented")
}
