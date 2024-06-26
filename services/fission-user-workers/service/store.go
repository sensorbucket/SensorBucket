package userworkers

import (
	"fmt"
	"regexp"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/internal/pagination"
)

var _ Store = (*PSQLStore)(nil)

type PSQLStore struct {
	db *sqlx.DB
}

func NewPSQLStore(db *sqlx.DB) *PSQLStore {
	return &PSQLStore{db}
}

func (s *PSQLStore) WorkersExists(ids []uuid.UUID, filters WorkerFilters) ([]uuid.UUID, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	q := sq.Select("id").From("user_workers").Where(sq.Eq{"id": ids, "state": StateEnabled})
	q = applyFilters(q, filters)
	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(s.db).Query()
	if err != nil {
		return nil, fmt.Errorf("error querying for worker ids: %w", err)
	}
	defer rows.Close()

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

func applyFilters(q sq.SelectBuilder, filters WorkerFilters) sq.SelectBuilder {
	if len(filters.ID) > 0 {
		ids := lo.Filter(filters.ID, func(id string, _ int) bool {
			return R_UUID.MatchString(id)
		})
		q = q.Where(sq.Eq{"id": ids})
	}
	if filters.State > StateUnknown {
		q = q.Where(sq.Eq{"state": filters.State})
	}
	if len(filters.TenantID) > 0 {
		q = q.Where(sq.Eq{"tenant_id": filters.TenantID})
	}
	return q
}

var R_UUID = regexp.MustCompile(`^[0-9a-fA-F]{8}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{12}$`)

func (s *PSQLStore) ListUserWorkers(filters WorkerFilters, req pagination.Request) (*pagination.Page[UserWorker], error) {
	var err error
	q := sq.Select(
		"id", "name", "description", "state", "language", "tenant_id", "revision",
		"status", "status_info", "source", "entrypoint",
	).
		From("user_workers")
	q = applyFilters(q, filters)

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
	defer rows.Close()

	workers := []UserWorker{}
	for rows.Next() {
		var worker UserWorker
		if err := rows.Scan(
			&worker.ID, &worker.Name, &worker.Description, &worker.State, &worker.Language, &worker.TenantID,
			&worker.Revision, &worker.Status, &worker.StatusInfo, &worker.ZipSource, &worker.Entrypoint,
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
            id, name, description, state, language, tenant_id, revision,
            status, source, entrypoint
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
        )`,
		worker.ID, worker.Name, worker.Description, worker.State, worker.Language, worker.TenantID,
		worker.Revision, worker.Status, worker.ZipSource, worker.Entrypoint,
	); err != nil {
		return fmt.Errorf("could not create worker in store: %w", err)
	}
	return nil
}

func (s *PSQLStore) GetWorkerByID(id uuid.UUID, filters WorkerFilters) (*UserWorker, error) {
	var worker UserWorker
	q := sq.Select("id", "name", "description", "state", "language", "tenant_id", "revision", "status", "source", "entrypoint").
		From("user_workers").
		Where(sq.Eq{"id": id})
	q = applyFilters(q, filters)
	err := q.PlaceholderFormat(sq.Dollar).RunWith(s.db).Scan(
		&worker.ID, &worker.Name, &worker.Description, &worker.State, &worker.Language, &worker.TenantID,
		&worker.Revision, &worker.Status, &worker.ZipSource, &worker.Entrypoint,
	)
	if err != nil {
		return nil, fmt.Errorf("error querying for worker ids: %w", err)
	}
	return &worker, nil
}

func (s *PSQLStore) UpdateWorker(worker *UserWorker) error {
	if _, err := s.db.Exec(
		`UPDATE user_workers SET source=$1, revision=$2, state=$3, description=$4, name=$5 WHERE id=$6`,
		worker.ZipSource, worker.Revision, worker.State, worker.Description, worker.Name, worker.ID,
	); err != nil {
		return fmt.Errorf("could not update worker in store: %w", err)
	}
	return nil
}
