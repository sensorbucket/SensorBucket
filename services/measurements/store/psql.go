package store

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"sensorbucket.nl/services/measurements/service"
)

var pq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// Ensure MeasurementStorePSQL implements MeasurementStore
var _ service.MeasurementStore = (*MeasurementStorePSQL)(nil)

// MeasurementStorePSQL Implements the measurementstore with a PostgreSQL database as backend
type MeasurementStorePSQL struct {
	db *sqlx.DB
}

func NewPSQL(db *sqlx.DB) *MeasurementStorePSQL {
	return &MeasurementStorePSQL{
		db: db,
	}
}

func (s *MeasurementStorePSQL) Insert(m service.Measurement) error {
	query, params, err := newInsertBuilder().
		SetUplinkMessageID(m.UplinkMessageID).
		SetDevice(m.DeviceID, m.DeviceCode, m.DeviceDescription).
		SetTimestamp(m.Timestamp).
		SetValue(m.Value).
		SetMeasurementType(m.MeasurementType, m.MeasurementTypeUnit).
		SetMetadata(m.Metadata).
		TrySetSensor(m.SensorCode, m.SensorDescription, m.SensorExternalID).
		TrySetLocation(m.LocationID, m.LocationName, m.LocationLongitude, m.LocationLatitude).
		TrySetCoordinates(m.Longitude, m.Latitude).
		Build()
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
		"device_id",
		"device_code",
		"device_description",
		"sensor_code",
		"sensor_description",
		"sensor_external_id",
		"timestamp",
		"value",
		"measurement_type",
		"measurement_type_unit",
		"location_id",
		"location_name",
		"ST_X(location_coordinates::geometry) as location_lng",
		"ST_Y(location_coordinates::geometry) as location_lat",
		"metadata",
	).
		From("measurements").
		Where("timestamp >= ?", query.Start).
		OrderBy("timestamp DESC").
		Limit(uint64(p.Limit + 1))

	// Use cursor otherwise end time
	if !p.Timestamp.IsZero() {
		q = q.Where("timestamp <= to_timestamp(?)", p.Timestamp.Unix())
	} else {
		q = q.Where("timestamp <= ?", query.End)
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
	if len(query.Filters.LocationIDs) > 0 {
		q = q.Where(sq.Eq{"location_id": query.Filters.LocationIDs})
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
			&m.DeviceID,
			&m.DeviceCode,
			&m.DeviceDescription,
			&m.SensorCode,
			&m.SensorDescription,
			&m.SensorExternalID,
			&m.Timestamp,
			&m.Value,
			&m.MeasurementType,
			&m.MeasurementTypeUnit,
			&m.LocationID,
			&m.LocationName,
			&m.LocationLongitude,
			&m.LocationLatitude,
			&m.Metadata,
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
			nextPage.Timestamp = m.Timestamp
			nextPage.Skip = 0
			// If our timestamp stayed the same then we have to skip more
			if nextPage.Timestamp == p.Timestamp {
				nextPage.Skip = p.Skip
			}
			for i := len(list) - 1; i >= 0; i-- {
				if list[i].Timestamp != nextPage.Timestamp {
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
