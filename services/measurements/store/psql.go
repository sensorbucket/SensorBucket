package store

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"sensorbucket.nl/sensorbucket/services/measurements/service"
)

var pq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// Ensure MeasurementStorePSQL implements MeasurementStore
var _ service.Store = (*MeasurementStorePSQL)(nil)

// MeasurementStorePSQL Implements the measurementstore with a PostgreSQL database as backend
type MeasurementStorePSQL struct {
	db *sqlx.DB
}

func NewPSQL(db *sqlx.DB) *MeasurementStorePSQL {
	return &MeasurementStorePSQL{
		db: db,
	}
}

func createInsertQuery(m service.Measurement) (string, []any, error) {
	values := map[string]any{}

	values["uplink_message_id"] = m.UplinkMessageID
	values["organisation_id"] = m.OrganisationID
	values["organisation_name"] = m.OrganisationName
	values["organisation_address"] = m.OrganisationAddress
	values["organisation_zipcode"] = m.OrganisationZipcode
	values["organisation_city"] = m.OrganisationCity
	values["organisation_coc"] = m.OrganisationCoC
	values["orgnisation_location_coc"] = m.OrganisationLocationCoC
	values["device_id"] = m.DeviceID
	values["device_code"] = m.DeviceCode
	values["device_description"] = m.DeviceDescription
	values["device_location"] = sq.Expr("ST_SETSRID(ST_POINT(?,?),4326)", m.DeviceLongitude, m.DeviceLatitude)
	values["device_location_description"] = m.DeviceLocationDescription
	values["device_properties"] = m.DeviceProperties
	values["sensor_id"] = m.SensorID
	values["sensor_code"] = m.SensorCode
	values["sensor_description"] = m.SensorDescription
	values["sensor_external_id"] = m.SensorExternalID
	values["sensor_properties"] = m.SensorProperties
	values["sensor_brand"] = m.SensorBrand
	values["measurement_type"] = m.MeasurementType
	values["measurement_unit"] = m.MeasurementUnit
	values["measurement_timestamp"] = m.MeasurementTimestamp
	values["measurement_value"] = m.MeasurementValue
	values["measurement_value_prefix"] = m.MeasurementValuePrefix
	values["measurement_value_prefix_factor"] = m.MeasurementValueFactor
	values["measurement_location"] = sq.Expr("ST_SETSRID(ST_POINT(?,?),4326)", m.MeasurementLongitude, m.MeasurementLatitude)

	return pq.Insert("measurement").SetMap(values).ToSql()
}

func (s *MeasurementStorePSQL) Insert(m service.Measurement) error {
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
func (s *MeasurementStorePSQL) Query(query service.Query, p service.Pagination) ([]service.Measurement, *service.Pagination, error) {
	q := pq.Select(
		"uplink_message_id",
		"organisation_id",
		"organisation_name",
		"organisation_address",
		"organisation_zipcode",
		"organisation_city",
		"organisation_coc",
		"orgnisation_location_coc",
		"device_id",
		"device_code",
		"device_description",
		"ST_Y(device_location) as device_latitude",
		"ST_X(device_location) as device_longitude",
		"device_location_description",
		"device_properties",
		"sensor_id",
		"sensor_code",
		"sensor_description",
		"sensor_external_id",
		"sensor_properties",
		"sensor_brand",
		"measurement_type",
		"measurement_unit",
		"measurement_timestamp",
		"measurement_value",
		"measurement_value_prefix",
		"measurement_value_prefix_factor",
		"ST_Y(measurement_location) as measurement_latitude",
		"ST_X(measurement_location) as measurement_longitude",
	).
		From("measurements").
		Where("measurement_timestamp >= ?", query.Start).
		OrderBy("measurement_timestamp DESC").
		Limit(uint64(p.Limit + 1))

	// Use cursor otherwise end time
	if !p.Timestamp.IsZero() {
		q = q.Where("measurement_timestamp <= to_timestamp(?)", p.Timestamp.Unix())
	} else {
		q = q.Where("measurement_timestamp <= ?", query.End)
	}
	q = q.Offset(uint64(p.Skip))

	if len(query.Filters.DeviceIDs) > 0 {
		q = q.Where(sq.Eq{"device_id": query.Filters.DeviceIDs})
	}
	if len(query.Filters.MeasurementTypes) > 0 {
		q = q.Where(sq.Eq{"measurement_type": query.Filters.MeasurementTypes})
	}
	if len(query.Filters.SensorCodes) > 0 {
		q = q.Where(sq.Eq{"sensor_code": query.Filters.SensorCodes})
	}

	rows, err := q.RunWith(s.db).Query()
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var nextPage *service.Pagination
	list := make([]service.Measurement, 0, p.Limit)
	for rows.Next() {
		var m service.Measurement
		err = rows.Scan(
			&m.UplinkMessageID,
			&m.OrganisationID,
			&m.OrganisationName,
			&m.OrganisationAddress,
			&m.OrganisationZipcode,
			&m.OrganisationCity,
			&m.OrganisationCoC,
			&m.OrganisationLocationCoC,
			&m.DeviceID,
			&m.DeviceCode,
			&m.DeviceDescription,
			&m.DeviceLatitude,
			&m.DeviceLongitude,
			&m.DeviceLocationDescription,
			&m.DeviceProperties,
			&m.SensorID,
			&m.SensorCode,
			&m.SensorDescription,
			&m.SensorExternalID,
			&m.SensorProperties,
			&m.SensorBrand,
			&m.MeasurementType,
			&m.MeasurementUnit,
			&m.MeasurementTimestamp,
			&m.MeasurementValue,
			&m.MeasurementValuePrefix,
			&m.MeasurementValueFactor,
			&m.MeasurementLatitude,
			&m.MeasurementLongitude,
		)
		if err != nil {
			return nil, nil, err
		}

		// We limit the query to p.limit + 1. As long as the list has not reached p.limit
		// we keep appending it. Once we get our +1 we will use that to  update the pagination
		if len(list) < p.Limit {
			list = append(list, m)
		} else {
			nextPage = &service.Pagination{}
			nextPage.Limit = p.Limit
			nextPage.Timestamp = m.MeasurementTimestamp
			nextPage.Skip = 0
			// If our timestamp stayed the same then we have to skip more
			if nextPage.Timestamp == p.Timestamp {
				nextPage.Skip = p.Skip
			}
			for i := len(list) - 1; i >= 0; i-- {
				if list[i].MeasurementTimestamp != nextPage.Timestamp {
					break
				}
				nextPage.Skip++
			}
		}
	}

	return list, nextPage, nil
}

// decodeCursor decodes the pagination cursor which is just a base64 encoded ISO8601 timestamp
func decodeCursor(cursor string) (uint64, error) {
	decoded, err := hex.DecodeString(cursor)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(decoded), nil
}

func encodeCursor(p service.Pagination, ts uint64) string {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, ts)
	return hex.EncodeToString(data)
}
