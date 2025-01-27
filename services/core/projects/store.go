package projects

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

var pq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

var _ Store = (*PostgresqlStore)(nil)

type PostgresqlStore struct {
	db *pgxpool.Pool
}

const createProjectSQL = `
  insert into projects (name, description, tenant_id) values ($1, $2, $3)
  returning id;
`

func (store *PostgresqlStore) CreateProject(ctx context.Context, project *Project) error {
	row := store.db.QueryRow(ctx, createProjectSQL)
	if err := row.Scan(&project.ID); err != nil {
		return fmt.Errorf("could not insert project: %w", err)
	}
	return nil
}

func (store PostgresqlStore) UpdateProject(ctx context.Context, params EditProjectParams) error {
	q := pq.Update("projects").Where(sq.Eq{
		"id": params.ID,
	})

	if params.Name != nil {
		q = q.Set("name", params.Name)
	}
	if params.Description != nil {
		q = q.Set("description", params.Description)
	}

	query, queryParams, _ := q.ToSql()

	_, err := store.db.Exec(ctx, query, queryParams)
	if err != nil {
		return err
	}
	return nil
}
