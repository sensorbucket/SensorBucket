package deviceinfra

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/services/core/devices"
	"sensorbucket.nl/sensorbucket/services/core/featuresofinterest"
)

var (
	pq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	_ devices.DeviceStore = (*PSQLStore)(nil)
)

type DB interface {
	Select(dest any, query string, args ...any) error
	Get(dest any, query string, args ...any) error
	sqlx.Execer
	sqlx.Queryer
}

type PSQLStore struct {
	db *sqlx.DB
}

func NewPSQLStore(db *sqlx.DB) *PSQLStore {
	return &PSQLStore{
		db: db,
	}
}

type DeviceModel struct {
	devices.Device
}

type DevicePaginationQuery struct {
	CreatedAt time.Time `pagination:"device.created_at,ASC"`
	ID        int64     `pagination:"device.id,ASC"`
}

func (s *PSQLStore) ListInBoundingBox(ctx context.Context, filter devices.DeviceFilter, p pagination.Request) (*pagination.Page[devices.Device], error) {
	return newDeviceQueryBuilder().WithPagination(p).WithFilters(filter).WithinBoundingBox(filter.BoundingBoxFilter).Query(ctx, s.db)
}

func (s *PSQLStore) ListInRange(ctx context.Context, filter devices.DeviceFilter, p pagination.Request) (*pagination.Page[devices.Device], error) {
	return newDeviceQueryBuilder().WithPagination(p).WithFilters(filter).WithinRange(filter.RangeFilter).Query(ctx, s.db)
}

func (s *PSQLStore) List(ctx context.Context, filter devices.DeviceFilter, p pagination.Request) (*pagination.Page[devices.Device], error) {
	return newDeviceQueryBuilder().WithPagination(p).WithFilters(filter).Query(ctx, s.db)
}

func (s *PSQLStore) Find(ctx context.Context, id int64) (*devices.Device, error) {
	return find(ctx, s.db, id)
}

func (s *PSQLStore) Delete(ctx context.Context, dev *devices.Device) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}

	_, err = pq.Delete("sensors").Where(sq.Eq{"device_id": dev.ID}).RunWith(tx).Exec()
	if err != nil {
		if rb := tx.Rollback(); rb != nil {
			err = fmt.Errorf("rollback failed with %w while handling error: %w", rb, err)
		}
		return err
	}

	_, err = pq.Delete("devices").Where(sq.Eq{"id": dev.ID}).RunWith(tx).Exec()
	if err != nil {
		if rb := tx.Rollback(); rb != nil {
			err = fmt.Errorf("rollback failed with %w while handling error: %w", rb, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

type SensorPaginationQuery struct {
	CreatedAt time.Time `pagination:"sensor.created_at,ASC"`
	ID        int64     `pagination:"sensor.id,ASC"`
}
type listSensorsRow struct {
	sensor                        devices.Sensor
	featureOfInterestID           sql.NullInt64
	featureOfInterestName         sql.NullString
	featureOfInterestDescription  sql.NullString
	featureOfInterestEncodingType sql.NullString
	featureOfInterestFeature      sql.Null[json.RawMessage]
	featureOfInterestProperties   sql.Null[json.RawMessage]
	featureOfInterestTenantID     sql.NullInt64
}

func (row *listSensorsRow) ToModel() devices.Sensor {
	if !row.featureOfInterestID.Valid {
		return row.sensor
	}
	row.sensor.FeatureOfInterest = &featuresofinterest.FeatureOfInterest{
		ID:           row.featureOfInterestID.Int64,
		Name:         row.featureOfInterestName.String,
		Description:  row.featureOfInterestDescription.String,
		EncodingType: row.featureOfInterestEncodingType.String,
		Feature:      row.featureOfInterestFeature.V,
		Properties:   row.featureOfInterestProperties.V,
		TenantID:     row.featureOfInterestTenantID.Int64,
	}
	return row.sensor
}

func (s *PSQLStore) ListSensors(ctx context.Context, p pagination.Request) (*pagination.Page[devices.Sensor], error) {
	cursor, err := pagination.GetCursor[SensorPaginationQuery](p)
	if err != nil {
		return nil, fmt.Errorf("list sensors, error getting pagination cursor: %w", err)
	}
	sensors, err := listSensorsPaginated(ctx, s.db, &cursor)
	if err != nil {
		return nil, fmt.Errorf("in ListSensors: %w", err)
	}
	page := pagination.CreatePageT(sensors, cursor)
	return &page, nil
}

func (s *PSQLStore) GetSensor(ctx context.Context, id int64) (*devices.Sensor, error) {
	return getSensor(ctx, s.db, id)
}

func (s *PSQLStore) createDevice(_ context.Context, dev *devices.Device) error {
	if err := s.db.Get(&dev.ID,
		`
			INSERT INTO "devices" (
				"code", "description", "tenant_id", "properties", "location",
				"altitude", "location_description", "state", "created_at"
			)
			VALUES ($1, $2, $3, $4, ST_POINT($5, $6), $7, $8, $9, $10)
			RETURNING id
		`,
		dev.Code, dev.Description, dev.TenantID, dev.Properties,
		dev.Longitude, dev.Latitude, dev.Altitude, dev.LocationDescription,
		dev.State, dev.CreatedAt,
	); err != nil {
		return err
	}
	return nil
}

func (s *PSQLStore) updateDevice(ctx context.Context, dev *devices.Device) error {
	q := sq.Update("devices").
		SetMap(map[string]interface{}{
			"description":          dev.Description,
			"properties":           dev.Properties,
			"location":             sq.Expr("ST_POINT(?, ?)", dev.Longitude, dev.Latitude),
			"altitude":             dev.Altitude,
			"location_description": dev.LocationDescription,
			"state":                dev.State,
		}).
		Where(sq.Eq{"id": dev.ID})
	q = auth.ProtectedQuery(ctx, "tenant_id", q)
	_, err := q.PlaceholderFormat(sq.Dollar).RunWith(s.db).Exec()
	if err != nil {
		return err
	}

	return nil
}

func (s *PSQLStore) updateSensors(ctx context.Context, devID int64, sensors []devices.Sensor) error {
	// Replace sensors
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}

	// Get delta with db
	dbSensors, err := listSensors(ctx, tx, ListSensorsFilter{DeviceID: []int64{devID}})
	if err != nil {
		if rb := tx.Rollback(); rb != nil {
			err = fmt.Errorf("rollback failed with %w while handling error: %w", rb, err)
		}
		return err
	}
	var createdSensors []*devices.Sensor
	var updatedSensors []*devices.Sensor
	var deletedSensors []int64
create_delta_loop:
	for ix := range sensors {
		s := &sensors[ix]
		// If the sensor has ID 0 then it is new and must be created
		if s.ID == 0 {
			createdSensors = append(createdSensors, s)
			continue
		}
		// If the sensor is present in db and this list then assume it is updated
		for _, dbs := range dbSensors {
			if dbs.ID == s.ID {
				updatedSensors = append(updatedSensors, s)
				continue create_delta_loop
			}
		}
		// If we reach this, it means the sensor has an ID but is not present
		// in the database. Should not be possible!
		log.Printf("[WARNING] device/store_psql#updateSensors was called with an unknown sensor (id: %d) that is neither new or already exists\n", s.ID)
	}
	// If a sensor is present in the database but not in the new list then it was removed
deleted_delta_loop:
	for _, dbs := range dbSensors {
		for _, s := range sensors {
			if s.ID == dbs.ID {
				continue deleted_delta_loop
			}
		}
		deletedSensors = append(deletedSensors, dbs.ID)
	}

	if err := createSensors(tx, createdSensors); err != nil {
		if rb := tx.Rollback(); rb != nil {
			err = fmt.Errorf("rollback failed with %w while handling error: %w", rb, err)
		}
		return err
	}
	if err := updateSensors(ctx, tx, updatedSensors); err != nil {
		if rb := tx.Rollback(); rb != nil {
			err = fmt.Errorf("rollback failed with %w while handling error: %w", rb, err)
		}
		return err
	}
	if err := deleteSensors(ctx, tx, deletedSensors); err != nil {
		if rb := tx.Rollback(); rb != nil {
			err = fmt.Errorf("rollback failed with %w while handling error: %w", rb, err)
		}
		return err
	}

	// Commit changes
	if err := tx.Commit(); err != nil {
		if rb := tx.Rollback(); rb != nil {
			err = fmt.Errorf("rollback failed with %w while handling error: %w", rb, err)
		}
		return err
	}
	return nil
}

func find(ctx context.Context, db DB, id int64) (*devices.Device, error) {
	var dev DeviceModel

	err := auth.ProtectedQuery(ctx, "tenant_id", pq.Select(
		"id", "code", "description", "tenant_id", "properties", "location_description",
		`ST_X("location"::geometry) AS longitude`, `ST_Y("location"::geometry) AS latitude`,
		"altitude", "state",
	).From("devices").Where(sq.Eq{"id": id})).RunWith(db).Scan(
		&dev.ID, &dev.Code,
		&dev.Description, &dev.TenantID, &dev.Properties, &dev.LocationDescription,
		&dev.Longitude, &dev.Latitude,
		&dev.Altitude, &dev.State,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, devices.ErrDeviceNotFound
		}
		return nil, fmt.Errorf("in find: %w", err)
	}

	sensors, err := listSensors(ctx, db, ListSensorsFilter{DeviceID: []int64{dev.ID}})
	if err != nil {
		return nil, fmt.Errorf("in find, listSensors: %w", err)
	}
	dev.Sensors = sensors

	return &dev.Device, nil
}

func listSensorsPaginated(ctx context.Context, db DB, cursor *pagination.Cursor[SensorPaginationQuery]) ([]devices.Sensor, error) {
	var err error
	if cursor == nil {
		newCursor, err := pagination.GetCursor[SensorPaginationQuery](pagination.Request{})
		if err != nil {
			return nil, fmt.Errorf("in listSensorsPaginated while creating empty cursor because none was given: %w", err)
		}
		cursor = &newCursor
	}

	q := pq.Select(
		"sensor.id", "sensor.code", "sensor.description", "sensor.device_id", "sensor.external_id", "sensor.properties", "sensor.archive_time",
		"sensor.brand", "sensor.created_at", "sensor.is_fallback", "sensor.tenant_id",
		"feature.id", "feature.name", "feature.description", "feature.encoding_type", "feature.feature", "feature.properties", "feature.tenant_id",
	).From("sensors sensor").LeftJoin("features_of_interest feature ON sensor.feature_of_interest_id = feature.id")
	q, err = pagination.Apply(q, *cursor)
	if err != nil {
		return nil, err
	}
	q = auth.ProtectedQuery(ctx, "sensor.tenant_id", q)

	rows, err := q.RunWith(db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sensors []devices.Sensor
	for rows.Next() {
		var row listSensorsRow
		err := rows.Scan(
			&row.sensor.ID,
			&row.sensor.Code,
			&row.sensor.Description,
			&row.sensor.DeviceID,
			&row.sensor.ExternalID,
			&row.sensor.Properties,
			&row.sensor.ArchiveTime,
			&row.sensor.Brand,
			&row.sensor.CreatedAt,
			&row.sensor.IsFallback,
			&row.sensor.TenantID,
			&row.featureOfInterestID,
			&row.featureOfInterestName,
			&row.featureOfInterestDescription,
			&row.featureOfInterestEncodingType,
			&row.featureOfInterestFeature,
			&row.featureOfInterestProperties,
			&row.featureOfInterestTenantID,
			&cursor.Columns.CreatedAt,
			&cursor.Columns.ID,
		)
		if err != nil {
			return nil, err
		}
		sensors = append(sensors, row.ToModel())
	}
	return sensors, nil
}

type ListSensorsFilter struct {
	DeviceID []int64
}

func listSensors(_ context.Context, db DB, filter ListSensorsFilter) ([]devices.Sensor, error) {
	var err error

	q := pq.Select(
		"sensor.id", "sensor.code", "sensor.description", "sensor.device_id", "sensor.external_id", "sensor.properties", "sensor.archive_time",
		"sensor.brand", "sensor.created_at", "sensor.is_fallback", "sensor.tenant_id",
		"feature.id", "feature.name", "feature.description", "feature.encoding_type", "feature.feature", "feature.properties", "feature.tenant_id",
	).From("sensors sensor").LeftJoin("features_of_interest feature ON sensor.feature_of_interest_id = feature.id").
		Where(sq.Eq{"sensor.device_id": filter.DeviceID})

	rows, err := q.RunWith(db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sensors []devices.Sensor
	for rows.Next() {
		var row listSensorsRow
		err := rows.Scan(
			&row.sensor.ID,
			&row.sensor.Code,
			&row.sensor.Description,
			&row.sensor.DeviceID,
			&row.sensor.ExternalID,
			&row.sensor.Properties,
			&row.sensor.ArchiveTime,
			&row.sensor.Brand,
			&row.sensor.CreatedAt,
			&row.sensor.IsFallback,
			&row.sensor.TenantID,
			&row.featureOfInterestID,
			&row.featureOfInterestName,
			&row.featureOfInterestDescription,
			&row.featureOfInterestEncodingType,
			&row.featureOfInterestFeature,
			&row.featureOfInterestProperties,
			&row.featureOfInterestTenantID,
		)
		if err != nil {
			return nil, err
		}
		sensors = append(sensors, row.ToModel())
	}
	return sensors, nil
}

func createSensors(tx DB, sensors []*devices.Sensor) error {
	if len(sensors) == 0 {
		return nil
	}
	q := pq.Insert("sensors").Columns(
		"code", "brand", "description", "archive_time", "properties", "external_id",
		"device_id", "created_at", "is_fallback", "tenant_id", "feature_of_interest_id",
	).Suffix("RETURNING id")
	for _, s := range sensors {
		featureID := sql.NullInt64{
			Int64: s.FeatureOfInterest.ID,
			Valid: s.FeatureOfInterest.ID > 0,
		}
		q = q.Values(
			s.Code, s.Brand, s.Description, s.ArchiveTime, s.Properties, s.ExternalID,
			s.DeviceID, s.CreatedAt, s.IsFallback, s.TenantID, featureID,
		)
	}
	query, params, err := q.ToSql()
	if err != nil {
		return err
	}
	var ids []int64
	if err := tx.Select(&ids, query, params...); err != nil {
		return err
	}
	for ix := range ids {
		sensors[ix].ID = ids[ix]
	}
	return nil
}

func getSensor(ctx context.Context, tx DB, id int64) (*devices.Sensor, error) {
	q := pq.Select(
		"sensor.id", "sensor.code", "sensor.description", "sensor.brand", "sensor.archive_time", "sensor.external_id",
		"sensor.properties", "sensor.created_at", "sensor.device_id", "sensor.is_fallback", "sensor.tenant_id",
		"feature.id", "feature.name", "feature.description", "feature.encoding_type", "feature.feature", "feature.properties", "feature.tenant_id",
	).From("sensors sensor").LeftJoin("features_of_interest feature ON sensor.feature_of_interest_id = feature.id").
		Where(sq.Eq{"sensor.id": id})
	q = auth.ProtectedQuery(ctx, "sensor.tenant_id", q)

	var row listSensorsRow
	err := q.RunWith(tx).Scan(
		&row.sensor.ID, &row.sensor.Code, &row.sensor.Description, &row.sensor.Brand, &row.sensor.ArchiveTime,
		&row.sensor.ExternalID, &row.sensor.Properties, &row.sensor.CreatedAt, &row.sensor.DeviceID, &row.sensor.IsFallback,
		&row.sensor.TenantID,
		&row.featureOfInterestID,
		&row.featureOfInterestName,
		&row.featureOfInterestDescription,
		&row.featureOfInterestEncodingType,
		&row.featureOfInterestFeature,
		&row.featureOfInterestProperties,
		&row.featureOfInterestTenantID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, devices.ErrSensorNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("in getSensor: %w", err)
	}
	sensor := row.ToModel()
	return &sensor, nil
}

func updateSensors(ctx context.Context, tx DB, sensors []*devices.Sensor) error {
	if len(sensors) == 0 {
		return nil
	}
	for _, s := range sensors {
		updateMap := map[string]any{
			"code":                   s.Code,
			"brand":                  s.Brand,
			"description":            s.Description,
			"archive_time":           s.ArchiveTime,
			"properties":             s.Properties,
			"external_id":            s.ExternalID,
			"device_id":              s.DeviceID,
			"is_fallback":            s.IsFallback,
			"feature_of_interest_id": nil,
		}
		if s.FeatureOfInterest != nil {
			updateMap["feature_of_interest_id"] = s.FeatureOfInterest.ID
		}

		q := pq.Update("sensors").Where(sq.Eq{"id": s.ID}).
			SetMap(updateMap)
		q = auth.ProtectedQuery(ctx, "tenant_id", q)
		_, err := q.RunWith(tx).Exec()
		if err != nil {
			return err
		}
	}
	return nil
}

func deleteSensors(ctx context.Context, tx DB, sensors []int64) error {
	if len(sensors) == 0 {
		return nil
	}
	q := pq.Delete("sensors").Where(sq.Eq{"id": sensors})
	q = auth.ProtectedQuery(ctx, "tenant_id", q)
	_, err := q.RunWith(tx).Exec()
	return err
}

func (s *PSQLStore) Save(ctx context.Context, dev *devices.Device) error {
	var err error
	if dev.ID == 0 {
		err = s.createDevice(ctx, dev)
	} else {
		err = s.updateDevice(ctx, dev)
	}
	if err != nil {
		return err
	}

	if err := s.updateSensors(ctx, dev.ID, dev.Sensors); err != nil {
		return err
	}

	return nil
}
