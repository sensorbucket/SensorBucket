package featuresofinterest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"sensorbucket.nl/sensorbucket/internal/pagination"
)

var (
	pq     = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	logger = slog.Default().With("component", "services/core/featuresofinterest")
)

var _ Store = (*StorePSQL)(nil)

type StorePSQL struct {
	databasePool *pgxpool.Pool
}

func NewStorePSQL(pool *pgxpool.Pool) *StorePSQL {
	return &StorePSQL{
		databasePool: pool,
	}
}

type featureOfInterestPagination struct {
	Offset uint64
}

func (store *StorePSQL) ListFeaturesOfInterest(ctx context.Context, filter FeatureOfInterestFilter, pageReq pagination.Request) (*pagination.Page[FeatureOfInterest], error) {
	cursor, err := pagination.GetCursor[featureOfInterestPagination](pageReq)
	if err != nil {
		return nil, fmt.Errorf("while decoding cursor: %w", err)
	}

	q := pq.Select("id", "name", "description", "encoding_type", "ST_AsBinary(feature)", "properties", "tenant_id").From("features_of_interest")
	if len(filter.TenantID) > 0 {
		q = q.Where(sq.Eq{"tenant_id": filter.TenantID})
	}
	if filter.Properties != nil {
		q = q.Where("properties::jsonb @> ?::jsonb", filter.Properties)
	}
	q = q.Offset(cursor.Columns.Offset).Limit(cursor.Limit)

	query, params, err := q.ToSql()
	if err != nil {
		panic(err)
	}
	rows, err := store.databasePool.Query(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("while querying database: %w", err)
	}

	features := make([]FeatureOfInterest, 0)
	for rows.Next() {
		var model FeatureOfInterest
		if err := rows.Scan(
			&model.ID, &model.Name, &model.Description, &model.EncodingType, &model.Feature, &model.Properties, &model.TenantID,
		); err != nil {
			return nil, fmt.Errorf("while scanning FeatureOfInterest: %w", err)
		}
		features = append(features, model)
	}

	cursor.Columns.Offset += uint64(len(features))
	page := pagination.CreatePageT(features, cursor)

	return &page, nil
}

func (store *StorePSQL) GetFeatureOfInterest(ctx context.Context, id int64, filter FeatureOfInterestFilter) (*FeatureOfInterest, error) {
	return store.getFeatureOfInterest(ctx, id, func(q sq.SelectBuilder) sq.SelectBuilder {
		if len(filter.TenantID) > 0 {
			q = q.Where(sq.Eq{"tenant_id": filter.TenantID})
		}
		return q
	})
}

type queryMod func(sq.SelectBuilder) sq.SelectBuilder

func (store *StorePSQL) getFeatureOfInterest(ctx context.Context, id int64, mods ...queryMod) (*FeatureOfInterest, error) {
	q := pq.Select(
		"foi.id", "foi.name", "foi.description", "foi.encoding_type", "ST_AsBinary(foi.feature)", "foi.properties", "foi.tenant_id",
	).From("features_of_interest foi").Where(sq.Eq{"foi.id": id})
	for _, mod := range mods {
		q = mod(q)
	}

	query, params, err := q.ToSql()
	if err != nil {
		panic(err)
	}
	row := store.databasePool.QueryRow(ctx, query, params...)

	var model FeatureOfInterest
	if err := row.Scan(
		&model.ID, &model.Name, &model.Description, &model.EncodingType, &model.Feature, &model.Properties, &model.TenantID,
	); errors.Is(err, sql.ErrNoRows) {
		return nil, ErrFeatureOfInterestNotFound
	} else if err != nil {
		return nil, fmt.Errorf("in GetFeatureOfInterest, while scanning row: %w", err)
	}

	foi := model
	return &foi, nil
}

func (store *StorePSQL) DeleteFeatureOfInterest(ctx context.Context, id int64) error {
	q := pq.Delete("features_of_interest").Where(sq.Eq{"id": id}).Limit(1)
	query, params, err := q.ToSql()
	if err != nil {
		panic(err)
	}
	if _, err := store.databasePool.Exec(ctx, query, params...); err != nil {
		return err
	}
	return nil
}

func (store *StorePSQL) SaveFeatureOfInterest(ctx context.Context, foi *FeatureOfInterest) error {
	if foi == nil {
		panic("in FeatureOfInterest/StorePSQL/SaveFeatureOfInterest: foi parameter can never be nil")
	}

	if foi.ID == 0 {
		return store.insertFeatureOfInterest(ctx, foi)
	}
	return store.updateFeatureOfInterest(ctx, foi)
}

func (store *StorePSQL) insertFeatureOfInterest(ctx context.Context, foi *FeatureOfInterest) error {
	q := pq.Insert("features_of_interest").Columns(
		"name", "description", "encoding_type", "feature", "properties", "tenant_id",
	).Values(
		foi.Name, foi.Description, foi.EncodingType, sq.Expr("ST_GeomFromEWKB(?)", foi.Feature), foi.Properties, foi.TenantID,
	).Suffix(`RETURNING "id"`)

	query, params, err := q.ToSql()
	if err != nil {
		panic(err)
	}
	if err := store.databasePool.QueryRow(ctx, query, params...).Scan(&foi.ID); err != nil {
		return fmt.Errorf("in StorePSQL: error scanning row: %w", err)
	}

	return nil
}

func (store *StorePSQL) updateFeatureOfInterest(ctx context.Context, foi *FeatureOfInterest) error {
	model, err := store.getFeatureOfInterest(ctx, foi.ID)
	if err != nil {
		return nil
	}

	updateMap := map[string]any{
		"name":          foi.Name,
		"description":   foi.Description,
		"encoding_type": foi.EncodingType,
		"properties":    foi.Properties,
	}
	if foi.Feature != nil {
		updateMap["feature"] = sq.Expr("ST_GeomFromEWKB(?)", *foi.Feature)
	} else if model.Feature != nil {
		updateMap["feature"] = nil
	}
	q := pq.Update("features_of_interest").Where(sq.Eq{"id": foi.ID}).SetMap(updateMap)

	query, params, err := q.ToSql()
	if err != nil {
		panic(err)
	}
	if _, err := store.databasePool.Exec(ctx, query, params...); err != nil {
		return fmt.Errorf("in updateFeatureOfInterest, failed to run update query: %w", err)
	}
	return nil
}
