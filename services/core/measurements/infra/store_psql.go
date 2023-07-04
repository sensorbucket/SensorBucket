package measurementsinfra

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/services/core/measurements"
)

var pq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// Ensure MeasurementStorePSQL implements MeasurementStore
var _ measurements.Store = (*MeasurementStorePSQL)(nil)

// MeasurementStorePSQL Implements the measurementstore with a PostgreSQL database as backend
type MeasurementStorePSQL struct {
	db *sqlx.DB
}

func NewPSQL(db *sqlx.DB) *MeasurementStorePSQL {
	return &MeasurementStorePSQL{
		db: db,
	}
}

func createInsertQuery(m measurements.Measurement) (string, []any, error) {
	values := map[string]any{}

	values["uplink_message_id"] = m.UplinkMessageID
	values["organisation_id"] = m.OrganisationID
	values["organisation_name"] = m.OrganisationName
	values["organisation_address"] = m.OrganisationAddress
	values["organisation_zipcode"] = m.OrganisationZipcode
	values["organisation_city"] = m.OrganisationCity
	values["organisation_chamber_of_commerce_id"] = m.OrganisationChamberOfCommerceID
	values["organisation_headquarter_id"] = m.OrganisationHeadquarterID
	values["organisation_state"] = m.OrganisationState
	values["organisation_archive_time"] = m.OrganisationArchiveTime
	values["device_id"] = m.DeviceID
	values["device_code"] = m.DeviceCode
	values["device_description"] = m.DeviceDescription
	values["device_location"] = sq.Expr("ST_SETSRID(ST_POINT(?,?),4326)", m.DeviceLongitude, m.DeviceLatitude)
	values["device_altitude"] = m.DeviceAltitude
	values["device_location_description"] = m.DeviceLocationDescription
	values["device_state"] = m.DeviceState
	values["device_properties"] = m.DeviceProperties
	values["sensor_id"] = m.SensorID
	values["sensor_code"] = m.SensorCode
	values["sensor_description"] = m.SensorDescription
	values["sensor_external_id"] = m.SensorExternalID
	values["sensor_properties"] = m.SensorProperties
	values["sensor_brand"] = m.SensorBrand
	values["sensor_archive_time"] = m.SensorArchiveTime
	values["datastream_id"] = m.DatastreamID
	values["datastream_description"] = m.DatastreamDescription
	values["datastream_observed_property"] = m.DatastreamObservedProperty
	values["datastream_unit_of_measurement"] = m.DatastreamUnitOfMeasurement
	values["measurement_timestamp"] = m.MeasurementTimestamp
	values["measurement_value"] = m.MeasurementValue
	values["measurement_location"] = sq.Expr("ST_SETSRID(ST_POINT(?,?),4326)", m.MeasurementLongitude, m.MeasurementLatitude)
	values["measurement_altitude"] = m.MeasurementAltitude
	values["measurement_expiration"] = m.MeasurementExpiration
	values["created_at"] = m.CreatedAt

	return pq.Insert("measurements").SetMap(values).ToSql()
}

func (s *MeasurementStorePSQL) Insert(m measurements.Measurement) error {
	query, params, err := createInsertQuery(m)
	if err != nil {
		return fmt.Errorf("could not generate query: %w", err)
	}

	_, err = s.db.Exec(query, params...)
	return err
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

func (s *MeasurementStorePSQL) Query(query measurements.Filter, r pagination.Request) (*pagination.Page[measurements.Measurement], error) {
	var err error
	q := pq.Select(
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
		"created_at",
	).
		From("measurements").
		Where("measurement_timestamp >= ?", query.Start)

	if len(query.DeviceIDs) > 0 {
		q = q.Where(sq.Eq{"device_id": query.DeviceIDs})
	}
	if len(query.SensorCodes) > 0 {
		q = q.Where(sq.Eq{"sensor_code": query.SensorCodes})
	}
	if len(query.Datastream) > 0 {
		q = q.Where(sq.Eq{"datastream_id": query.Datastream})
	}

	// pagination
	cursor := pagination.GetCursor[MeasurementQueryPage](r)
	q, err = pagination.Apply(q, cursor)
	if err != nil {
		return nil, err
	}

	rows, err := q.RunWith(s.db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]measurements.Measurement, 0, cursor.Limit)
	for rows.Next() {
		var m measurements.Measurement
		err = rows.Scan(
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

func (s *MeasurementStorePSQL) FindDatastream(sensorID int64, obs string) (*measurements.Datastream, error) {
	var ds measurements.Datastream
	query := `
		SELECT
			"id", "description", "sensor_id", "observed_property", "unit_of_measurement",
			"created_at"
		FROM 
			"datastreams"
		WHERE
			"sensor_id"=$1 AND "observed_property"=$2
	`
	if err := s.db.Get(&ds, query, sensorID, obs); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, measurements.ErrDatastreamNotFound
		}
		return nil, fmt.Errorf("database error querying datastream: %w", err)
	}
	return &ds, nil
}

func (s *MeasurementStorePSQL) CreateDatastream(ds *measurements.Datastream) error {
	_, err := s.db.Exec(`
	INSERT INTO
		"datastreams" (
			"id", "description", "sensor_id", "observed_property", "unit_of_measurement",
			"created_at"
		)
	VALUES 
		($1, $2, $3, $4, $5, $6)
	`, ds.ID, ds.Description, ds.SensorID, ds.ObservedProperty, ds.UnitOfMeasurement, ds.CreatedAt)
	if err != nil {
		return fmt.Errorf("database error inserting datastream: %w", err)
	}
	return nil
}

type datastreamPageQuery struct {
	CreatedAt time.Time `pagination:"created_at,ASC"`
	ID        uuid.UUID `pagination:"id,ASC"`
}

func (s *MeasurementStorePSQL) ListDatastreams(filter measurements.DatastreamFilter, r pagination.Request) (*pagination.Page[measurements.Datastream], error) {
	var err error
	var ds = []measurements.Datastream{}
	q := pq.Select(
		"id", "description", "sensor_id", "observed_property", "unit_of_measurement", "created_at",
	).From("datastreams")

	if len(filter.Sensor) > 0 {
		q = q.Where(sq.Eq{"sensor_id": filter.Sensor})
	}

	cursor := pagination.GetCursor[datastreamPageQuery](r)
	q, err = pagination.Apply(q, cursor)
	if err != nil {
		return nil, err
	}

	rows, err := q.RunWith(s.db).Query()
	if err != nil {
		return nil, fmt.Errorf("error selecting datastreams from db: %w", err)
	}

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
