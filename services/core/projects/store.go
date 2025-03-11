package projects

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/lo"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/auth"
)

var pq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

var _ Store = (*PostgresqlStore)(nil)

type PostgresqlStore struct {
	db *pgxpool.Pool
}

func NewPostgresStore(db *pgxpool.Pool) *PostgresqlStore {
	return &PostgresqlStore{
		db: db,
	}
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

func (store *PostgresqlStore) EditProject(ctx context.Context, params EditProjectParams) error {
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

	result, err := store.db.Exec(ctx, query, queryParams)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrProjectNotFound
	}

	return nil
}

type listProjectsRow struct {
	ProjectID                      int64
	ProjectName                    string
	ProjectDescription             string
	ProjectFeatureObservationTypes pgtype.Array[string]
	FeatureID                      sql.NullInt64
	FeatureName                    sql.NullString
	FeatureDescription             sql.NullString
	FeatureEncodingType            sql.NullString
	FeatureGeometry                any
}

type listCursor struct {
	Offset uint64
}

func (store *PostgresqlStore) ListProjects(ctx context.Context, filter ProjectsFilter, req pagination.Request) (*pagination.Page[*Project], error) {
	cursor, err := pagination.GetCursor[listCursor](req)
	if err != nil {
		return nil, fmt.Errorf("", err)
	}
	// paginate the projects table seperately
	projectsQ := pq.Select("id, name, description").From("projects").OrderBy("id ASC").Offset(cursor.Columns.Offset).Limit(cursor.Limit)
	projectsQ = auth.ProtectedQuery(ctx, "tenant_id", projectsQ)

	q := pq.Select(`
    project.id, project.name, project.description,
    project_feature.interested_observation_types,
    feature.id, feature.name, feature.description, feature.encoding_type, feature.feature
  `).FromSelect(projectsQ, "project").
		LeftJoin("project_feature_of_interest project_feature ON project_feature.project_id = project.id").
		LeftJoin("feature_of_interest feature ON project_feature.feature_of_interest_id = feature.id")
	// OrderBy("project.id ASC")

	query, params, err := q.ToSql()
	if err != nil {
		panic(err)
	}
	rows, err := store.db.Query(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("", err)
	}
	defer rows.Close()

	projectMap := map[int64]*Project{}
	var row listProjectsRow
	for rows.Next() {
		if err := rows.Scan(
			&row.ProjectID, &row.ProjectName, &row.ProjectDescription,
			&row.ProjectFeatureObservationTypes,
			&row.FeatureID, &row.FeatureName, &row.FeatureDescription, &row.FeatureEncodingType, &row.FeatureGeometry,
		); err != nil {
			return nil, fmt.Errorf("", err)
		}

		project, ok := projectMap[row.ProjectID]
		// Create project if not yet exists
		if !ok {
			project = &Project{
				ID:                 row.ProjectID,
				Name:               row.ProjectName,
				Description:        row.ProjectDescription,
				FeaturesOfInterest: make([]ProjectFeatureOfInterest, 0),
			}
			projectMap[row.ProjectID] = project
		}
		// If this project has no features continue
		if !row.FeatureID.Valid {
			continue
		}
		// Otherwise add it to the project
		project.FeaturesOfInterest = append(project.FeaturesOfInterest, ProjectFeatureOfInterest{
			InterestedObservationTypes: row.ProjectFeatureObservationTypes.Elements,
			FeatureOfInterest: FeatureOfInterest{
				ID:          row.FeatureID.Int64,
				Name:        row.FeatureName.String,
				Description: row.FeatureName.String,
				// TODO add encoding type and geography
			},
		})
	}
	projects := lo.Values(projectMap)
	cursor.Columns.Offset += uint64(len(projects))
	page := pagination.CreatePageT(projects, cursor)
	return &page, nil
}

func (store *PostgresqlStore) RemoveProject(ctx context.Context, id int64) error {
	return nil
}

func (store *PostgresqlStore) SetProjectFeatureOfInterest(ctx context.Context, params ModifyProjectFeatureOfInterestParams) error {
	return nil
}
