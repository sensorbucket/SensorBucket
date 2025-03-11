package featuresofinterest

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twpayne/go-geom/encoding/ewkb"
	"sensorbucket.nl/sensorbucket/internal/pagination"
)

var pq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

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

	q := pq.Select("id", "name", "description", "encoding_type", "feature", "properties", "tenant_id").From("features_of_interest")
	if len(filter.TenantID) > 0 {
		q = q.Where(sq.Eq{"tenant_id": filter.TenantID})
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
		var feature FeatureOfInterest
		if err := rows.Scan(
			&feature.ID, &feature.Name, &feature.Description, &feature.EncodingType, &feature.Feature, &feature.Properties, &feature.TenantID,
		); err != nil {
			return nil, fmt.Errorf("while scanning FeatureOfInterest: %w", err)
		}
		features = append(features, feature)
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
		"foi.id", "foi.name", "foi.description", "foi.encoding_type", "foi.feature", "foi.properties", "foi.tenant_id",
	).From("features_of_interest foi").Where(sq.Eq{"foi.id": id})
	for _, mod := range mods {
		q = mod(q)
	}

	query, params, err := q.ToSql()
	if err != nil {
		panic(err)
	}
	row := store.databasePool.QueryRow(ctx, query, params...)

	var feature FeatureOfInterest
	if err := row.Scan(
		&feature.ID, &feature.Name, &feature.Description, &feature.EncodingType, &feature.Feature, &feature.Properties, &feature.TenantID,
	); errors.Is(err, sql.ErrNoRows) {
		return nil, ErrFeatureOfInterestNotFound
	} else if err != nil {
		return nil, fmt.Errorf("in GetFeatureOfInterest, while scanning row: %w", err)
	}

	return &feature, nil
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

func (store *StorePSQL) UpdateFeatureOfInterest(ctx context.Context, id int64, opts UpdateFeatureOfInterestOpts) error {
	q := pq.Update("features_of_interest").Where(sq.Eq{"id": id})

	if opts.Name == nil && opts.Description == nil && opts.EncodingType == nil && opts.Feature == nil && opts.Properties == nil {
		return errors.New("in StorePSQL/UpdateFeatureOfInterest: no properties to update")
	}

	if opts.Name != nil {
		q = q.Set("name", *opts.Name)
	}
	if opts.Description != nil {
		q = q.Set("description", *opts.Description)
	}
	if opts.EncodingType != nil {
		q = q.Set("encoding_type", *opts.EncodingType)
	}
	if opts.Feature != nil {
		q = q.Set("feature", opts.Feature)
	}
	if opts.Properties != nil {
		q = q.Set("properties", *opts.Properties)
	}

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
		foi.Name, foi.Description, foi.EncodingType, foi.Feature, foi.Properties, foi.TenantID,
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

	q := pq.Update("features_of_interest").Where(sq.Eq{"id": foi.ID})

	if foi.Name != model.Name {
		q = q.Set("name", foi.Name)
	}
	if foi.Description != model.Description {
		q = q.Set("description", foi.Description)
	}
	if foi.EncodingType != model.EncodingType {
		q = q.Set("encoding_type", foi.EncodingType)
	}
	if foi.Feature != model.Feature {
		q = q.Set("feature", &ewkb.Point{Point: foi.Feature.SetSRID(4362)})
	}
	if !bytes.Equal(foi.Properties, model.Properties) {
		q = q.Set("properties", foi.Properties)
	}

	query, params, err := q.ToSql()
	if err != nil {
		panic(err)
	}
	if _, err := store.databasePool.Exec(ctx, query, params...); err != nil {
		return err
	}
	return nil
}
