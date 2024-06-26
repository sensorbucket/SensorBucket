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
	if err != nil {
		return fmt.Errorf("could not insert new measurement: %w", err)
	}
	return nil
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
		"created_at",
	).
		From("measurements")

	if !query.Start.IsZero() {
		q = q.Where("measurement_timestamp >= ?", query.Start)
	}
	if !query.End.IsZero() {
		q = q.Where("measurement_timestamp <= ?", query.End)
	}

	if len(query.DeviceIDs) > 0 {
		q = q.Where(sq.Eq{"device_id": query.DeviceIDs})
	}
	if len(query.SensorCodes) > 0 {
		q = q.Where(sq.Eq{"sensor_code": query.SensorCodes})
	}
	if len(query.Datastream) > 0 {
		q = q.Where(sq.Eq{"datastream_id": query.Datastream})
	}
	if len(query.TenantID) > 0 {
		q = q.Where(sq.Eq{"organisation_id": query.TenantID})
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

	rows, err := q.RunWith(s.db).Query()
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

func (s *MeasurementStorePSQL) FindDatastream(tenantID, sensorID int64, obs string) (*measurements.Datastream, error) {
	var ds measurements.Datastream
	err := pq.Select("id", "description", "sensor_id", "observed_property", "unit_of_measurement",
		"created_at", "tenant_id").From("datastreams").Where(sq.Eq{
		"sensor_id":         sensorID,
		"observed_property": obs,
		"tenant_id":         tenantID,
	}).RunWith(s.db).Scan(
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

func (s *MeasurementStorePSQL) CreateDatastream(ds *measurements.Datastream) error {
	// TODO: Why can't uuid be marshalled by pgx?
	//
	uuidB, err := ds.ID.MarshalBinary()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`
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

func (s *MeasurementStorePSQL) ListDatastreams(filter measurements.DatastreamFilter, r pagination.Request) (*pagination.Page[measurements.Datastream], error) {
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

	rows, err := q.RunWith(s.db).Query()
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

func (s *MeasurementStorePSQL) GetDatastream(id uuid.UUID, filter measurements.DatastreamFilter) (*measurements.Datastream, error) {
	var ds measurements.Datastream
	idB, _ := id.MarshalBinary()
	q := pq.Select(
		"id", "description", "sensor_id", "observed_property", "unit_of_measurement", "created_at",
	).From("datastreams").Where(sq.Eq{"id": idB})
	q = applyDatastreamFilter(q, filter)

	err := q.RunWith(s.db).Scan(
		&ds.ID, &ds.Description, &ds.SensorID, &ds.ObservedProperty, &ds.UnitOfMeasurement, &ds.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &ds, nil
}
