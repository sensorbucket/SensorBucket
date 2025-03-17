package measurementsinfra

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/services/core/measurements"
)

var pq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// Ensure MeasurementStorePSQL implements MeasurementStore
var _ measurements.Store = (*MeasurementStorePSQL)(nil)

// MeasurementStorePSQL Implements the measurementstore with a PostgreSQL database as backend
type MeasurementStorePSQL struct {
	databasePool *pgxpool.Pool
}

func NewPSQL(databasePool *pgxpool.Pool) *MeasurementStorePSQL {
	return &MeasurementStorePSQL{
		databasePool: databasePool,
	}
}

// Query returns measurements from the database
//
//   - The query is based on the filters provided in the query.
//   - The query is ordered by the timestamp descending.
//   - The query has a start and end date, though it is paginated.
//   - The query is limited to the limit specified in the pagination + 1 entry, if this extra entry is populated then we
//     know that there is a next page available and we use this entry to populate the cursor.
//     The cursor holds the timestamp of the first entry of the next page as seconds since epoch base64
type MeasurementQueryPage struct {
	MeasurementTimestamp time.Time `pagination:"measurement_timestamp,DESC"`
	ID                   int64     `pagination:"id,DESC"`
}

func (s *MeasurementStorePSQL) Query(ctx context.Context, filter measurements.Filter, r pagination.Request) (*pagination.Page[measurements.Measurement], error) {
	var err error
	q := pq.Select(
		"id",
		"uplink_message_id",
		"organisation_id",
		"organisation_name",
		"organisation_address",
		"organisation_zipcode",
		"organisation_city",
		"organisation_chamber_of_commerce_id",
		"organisation_headquarter_id",
		"organisation_archive_time",
		"organisation_state",
		"device_id",
		"device_code",
		"device_description",
		"ST_Y(device_location::geometry) as device_latitude",
		"ST_X(device_location::geometry) as device_longitude",
		"device_altitude",
		"device_location_description",
		"device_properties",
		"device_state",
		"sensor_id",
		"sensor_code",
		"sensor_description",
		"sensor_external_id",
		"sensor_properties",
		"sensor_brand",
		"datastream_id",
		"datastream_description",
		"datastream_observed_property",
		"datastream_unit_of_measurement",
		"measurement_timestamp",
		"measurement_value",
		"ST_Y(measurement_location::geometry) as measurement_latitude",
		"ST_X(measurement_location::geometry) as measurement_longitude",
		"measurement_altitude",
		"measurement_expiration",
		"feature_of_interest_id",
		"feature_of_interest_name",
		"feature_of_interest_description",
		"feature_of_interest_encoding_type",
		"ST_AsBinary(feature_of_interest_feature)",
		"feature_of_interest_properties",
		"created_at",
	).
		From("measurements")

	if !filter.Start.IsZero() {
		q = q.Where("measurement_timestamp >= ?", filter.Start)
	}
	if !filter.End.IsZero() {
		q = q.Where("measurement_timestamp <= ?", filter.End)
	}

	if len(filter.FeatureOfInterestID) > 0 {
		q = q.Where(sq.Eq{"feature_of_interest_id": filter.FeatureOfInterestID})
	}
	if len(filter.ObservedProperty) > 0 {
		q = q.Where(sq.Eq{"datastream_observed_property": filter.ObservedProperty})
	}
	if len(filter.SensorCodes) > 0 {
		q = q.Where(sq.Eq{"sensor_code": filter.SensorCodes})
	}
	if len(filter.Datastream) > 0 {
		q = q.Where(sq.Eq{"datastream_id": filter.Datastream})
	}
	if len(filter.TenantID) > 0 {
		q = q.Where(sq.Eq{"organisation_id": filter.TenantID})
	}

	// pagination
	cursor, err := pagination.GetCursor[MeasurementQueryPage](r)
	if err != nil {
		return nil, fmt.Errorf("Query Measurements, error getting pagination cursor: %w", err)
	}
	q, err = pagination.Apply(q, cursor)
	if err != nil {
		return nil, err
	}

	sqlQuery, params, err := q.ToSql()
	if err != nil {
		panic(err)
	}
	rows, err := s.databasePool.Query(ctx, sqlQuery, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]measurements.Measurement, 0, cursor.Limit)
	for rows.Next() {
		var m measurements.Measurement
		err = rows.Scan(
			&m.ID,
			&m.UplinkMessageID,
			&m.OrganisationID,
			&m.OrganisationName,
			&m.OrganisationAddress,
			&m.OrganisationZipcode,
			&m.OrganisationCity,
			&m.OrganisationChamberOfCommerceID,
			&m.OrganisationHeadquarterID,
			&m.OrganisationArchiveTime,
			&m.OrganisationState,
			&m.DeviceID,
			&m.DeviceCode,
			&m.DeviceDescription,
			&m.DeviceLatitude,
			&m.DeviceLongitude,
			&m.DeviceAltitude,
			&m.DeviceLocationDescription,
			&m.DeviceProperties,
			&m.DeviceState,
			&m.SensorID,
			&m.SensorCode,
			&m.SensorDescription,
			&m.SensorExternalID,
			&m.SensorProperties,
			&m.SensorBrand,
			&m.DatastreamID,
			&m.DatastreamDescription,
			&m.DatastreamObservedProperty,
			&m.DatastreamUnitOfMeasurement,
			&m.MeasurementTimestamp,
			&m.MeasurementValue,
			&m.MeasurementLatitude,
			&m.MeasurementLongitude,
			&m.MeasurementAltitude,
			&m.MeasurementExpiration,
			&m.FeatureOfInterestID,
			&m.FeatureOfInterestName,
			&m.FeatureOfInterestDescription,
			&m.FeatureOfInterestEncodingType,
			&m.FeatureOfInterestFeature,
			&m.FeatureOfInterestProperties,
			&m.CreatedAt,
			&cursor.Columns.MeasurementTimestamp,
			&cursor.Columns.ID,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, m)
	}

	page := pagination.CreatePageT(list, cursor)
	return &page, nil
}

func (s *MeasurementStorePSQL) FindDatastream(ctx context.Context, tenantID, sensorID int64, obs string) (*measurements.Datastream, error) {
	var ds measurements.Datastream
	query, params, err := pq.Select(
		"id", "description", "sensor_id", "observed_property", "unit_of_measurement",
		"created_at", "tenant_id",
	).From("datastreams").Where(sq.Eq{
		"sensor_id":         sensorID,
		"observed_property": obs,
		"tenant_id":         tenantID,
	}).ToSql()
	if err != nil {
		panic(err)
	}
	row := s.databasePool.QueryRow(ctx, query, params...)
	err = row.Scan(
		&ds.ID, &ds.Description, &ds.SensorID, &ds.ObservedProperty, &ds.UnitOfMeasurement, &ds.CreatedAt,
		&ds.TenantID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, measurements.ErrDatastreamNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("database error querying datastream: %w", err)
	}
	return &ds, nil
}

func (s *MeasurementStorePSQL) CreateDatastream(ctx context.Context, ds *measurements.Datastream) error {
	// TODO: Why can't uuid be marshalled by pgx?
	//
	uuidB, err := ds.ID.MarshalBinary()
	if err != nil {
		return err
	}
	_, err = s.databasePool.Exec(ctx, `
	INSERT INTO
		"datastreams" (
			id, "description", "sensor_id", "observed_property", "unit_of_measurement",
			"created_at", "tenant_id"
		)
	VALUES 
		($1, $2, $3, $4, $5, $6, $7)
	`, uuidB, ds.Description, ds.SensorID, ds.ObservedProperty, ds.UnitOfMeasurement, ds.CreatedAt, ds.TenantID)
	if err != nil {
		return fmt.Errorf("database error inserting datastream: %w", err)
	}
	return nil
}

func applyDatastreamFilter(q sq.SelectBuilder, filter measurements.DatastreamFilter) sq.SelectBuilder {
	if len(filter.Sensor) > 0 {
		q = q.Where(sq.Eq{"sensor_id": filter.Sensor})
	}
	if len(filter.ObservedProperty) > 0 {
		q = q.Where(sq.Eq{"observed_property": filter.ObservedProperty})
	}
	if len(filter.TenantID) > 0 {
		q = q.Where(sq.Eq{"tenant_id": filter.TenantID})
	}
	return q
}

type datastreamPageQuery struct {
	CreatedAt time.Time `pagination:"created_at,ASC"`
	ID        uuid.UUID `pagination:"id,ASC"`
}

func (s *MeasurementStorePSQL) ListDatastreams(ctx context.Context, filter measurements.DatastreamFilter, r pagination.Request) (*pagination.Page[measurements.Datastream], error) {
	var err error
	ds := []measurements.Datastream{}
	q := pq.Select(
		"id", "description", "sensor_id", "observed_property", "unit_of_measurement", "created_at",
	).From("datastreams")
	q = applyDatastreamFilter(q, filter)

	cursor, err := pagination.GetCursor[datastreamPageQuery](r)
	if err != nil {
		return nil, fmt.Errorf("list datastreams, error getting pagination cursor: %w", err)
	}
	q, err = pagination.Apply(q, cursor)
	if err != nil {
		return nil, err
	}

	query, params, err := q.ToSql()
	if err != nil {
		panic(err)
	}
	rows, err := s.databasePool.Query(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("error selecting datastreams from db: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var d measurements.Datastream
		err := rows.Scan(
			&d.ID,
			&d.Description,
			&d.SensorID,
			&d.ObservedProperty,
			&d.UnitOfMeasurement,
			&d.CreatedAt,
			&cursor.Columns.CreatedAt,
			&cursor.Columns.ID,
		)
		if err != nil {
			return nil, err
		}
		ds = append(ds, d)
	}

	page := pagination.CreatePageT(ds, cursor)
	return &page, nil
}

func (s *MeasurementStorePSQL) GetDatastream(ctx context.Context, id uuid.UUID, filter measurements.DatastreamFilter) (*measurements.Datastream, error) {
	var ds measurements.Datastream
	idB, _ := id.MarshalBinary()
	q := pq.Select(
		"id", "description", "sensor_id", "observed_property", "unit_of_measurement", "created_at",
	).From("datastreams").Where(sq.Eq{"id": idB})
	q = applyDatastreamFilter(q, filter)

	query, params, err := q.ToSql()
	if err != nil {
		panic(err)
	}
	err = s.databasePool.QueryRow(ctx, query, params...).Scan(
		&ds.ID, &ds.Description, &ds.SensorID, &ds.ObservedProperty, &ds.UnitOfMeasurement,
		&ds.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &ds, nil
}

func (s *MeasurementStorePSQL) FindOrCreateDatastream(ctx context.Context, tenantID, sensorID int64, observedProperty, UnitOfMeasurement string) (*measurements.Datastream, error) {
	var ds measurements.Datastream
	err := s.databasePool.QueryRow(ctx,
		`SELECT 
      id, description, sensor_id, observed_property, unit_of_measurement, created_at, tenant_id
     FROM find_or_create_datastream($1, $2, $3, $4)`,
		tenantID, sensorID, observedProperty, UnitOfMeasurement,
	).Scan(
		&ds.ID, &ds.Description, &ds.SensorID, &ds.ObservedProperty, &ds.UnitOfMeasurement, &ds.CreatedAt, &ds.TenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("could not query datastream: %w", err)
	}
	return &ds, nil
}

func (s *MeasurementStorePSQL) StoreMeasurements(ctx context.Context, measurements []measurements.Measurement) error {
	var batch pgx.Batch
	for _, measurement := range measurements {
		batch.Queue(`
INSERT INTO measurements (
			uplink_message_id,
			organisation_id,
			organisation_name,
			organisation_address,
			organisation_zipcode,
			organisation_city,
			organisation_chamber_of_commerce_id,
			organisation_headquarter_id,
			organisation_state,
			organisation_archive_time,
			device_id,
			device_code,
			device_description,
			device_location,
			device_altitude,
			device_location_description,
			device_state,
			device_properties,
			sensor_id,
			sensor_code,
			sensor_description,
			sensor_external_id,
			sensor_properties,
			sensor_brand,
			sensor_archive_time,
			datastream_id,
			datastream_description,
			datastream_observed_property,
			datastream_unit_of_measurement,
			measurement_timestamp,
			measurement_value,
			measurement_location,
			measurement_altitude,
			measurement_expiration,
      feature_of_interest_id,
      feature_of_interest_name,
      feature_of_interest_description,
      feature_of_interest_encoding_type,
      feature_of_interest_feature,
      feature_of_interest_properties,
			created_at
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11,
  $12,
  $13,
  ST_SETSRID(ST_POINT($14,$15),4326),
  $16,
  $17,
  $18,
  $19,
  $20,
  $21,
  $22,
  $23,
  $24,
  $25,
  $26,
  $27,
  $28,
  $29,
  $30,
  $31,
  $32,
  ST_SETSRID(ST_POINT($33,$34),4326),
  $35,
  $36,
  $37,
  $38,
  $39,
  $40,
  ST_GeomFromEWKB($41),
  $42,
  $43
);

`,
			measurement.UplinkMessageID,
			measurement.OrganisationID,
			measurement.OrganisationName,
			measurement.OrganisationAddress,
			measurement.OrganisationZipcode,
			measurement.OrganisationCity,
			measurement.OrganisationChamberOfCommerceID,
			measurement.OrganisationHeadquarterID,
			measurement.OrganisationState,
			measurement.OrganisationArchiveTime,
			measurement.DeviceID,
			measurement.DeviceCode,
			measurement.DeviceDescription,
			measurement.DeviceLongitude, measurement.DeviceLatitude,
			measurement.DeviceAltitude,
			measurement.DeviceLocationDescription,
			measurement.DeviceState,
			measurement.DeviceProperties,
			measurement.SensorID,
			measurement.SensorCode,
			measurement.SensorDescription,
			measurement.SensorExternalID,
			measurement.SensorProperties,
			measurement.SensorBrand,
			measurement.SensorArchiveTime,
			measurement.DatastreamID,
			measurement.DatastreamDescription,
			measurement.DatastreamObservedProperty,
			measurement.DatastreamUnitOfMeasurement,
			measurement.MeasurementTimestamp,
			measurement.MeasurementValue,
			measurement.MeasurementLongitude, measurement.MeasurementLatitude,
			measurement.MeasurementAltitude,
			measurement.MeasurementExpiration,
			measurement.FeatureOfInterestID,
			measurement.FeatureOfInterestName,
			measurement.FeatureOfInterestDescription,
			measurement.FeatureOfInterestEncodingType,
			measurement.FeatureOfInterestFeature,
			measurement.FeatureOfInterestProperties,
			measurement.CreatedAt,
		)
	}

	batchResult := s.databasePool.SendBatch(ctx, &batch)
	defer batchResult.Close()

	for range len(measurements) {
		_, err := batchResult.Exec()
		if err != nil {
			log.Printf("Batch inser resulted in an error: %s\n", err.Error())
		}
	}

	return nil
}
