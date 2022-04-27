package store

import (
	"database/sql"
	"encoding/binary"
	"encoding/hex"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"sensorbucket.nl/internal/measurements"
)

var pq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// Ensure MeasurementStorePSQL implements MeasurementStore
var _ measurements.MeasurementStore = (*MeasurementStorePSQL)(nil)

// MeasurementStorePSQL Implements the measurementstore with a PostgreSQL database as backend
type MeasurementStorePSQL struct {
	db *sqlx.DB
}

func NewPSQL(db *sqlx.DB) *MeasurementStorePSQL {
	return &MeasurementStorePSQL{
		db: db,
	}
}

func (s *MeasurementStorePSQL) Insert(m *measurements.Measurement) error {
	var locID sql.NullInt64
	if m.LocationID != nil {
		locID.Int64 = int64(*m.LocationID)
		locID.Valid = true
	}

	_, err := s.db.Exec("INSERT INTO measurements (thing_urn,timestamp,value,measurement_type,measurement_type_unit,location_id,coordinates,metadata) VALUES ($1, $2, $3, $4, $5, $6, ST_SetSRID(ST_MakePoint($7, $8),4326), $9)",
		m.ThingURN,
		m.Timestamp,
		m.Value,
		m.MeasurementType,
		m.MeasurementTypeUnit,
		locID,
		m.Coordinates[0],
		m.Coordinates[1],
		m.Metadata,
	)
	return err
}

// Query returns measurements from the database
//
// - The query is based on the filters provided in the query.
// - The query is ordered by the timestamp descending.
// - The query has a start and end date, though it is paginated.
// - The query is limited to the limit specified in the pagination + 1 entry, if this extra entry is populated then we
//   know that there is a next page available and we use this entry to populate the cursor.
//   The cursor holds the timestamp of the first entry of the next page as seconds since epoch base64
func (s *MeasurementStorePSQL) Query(query measurements.Query, p measurements.Pagination) ([]measurements.Measurement, *measurements.Pagination, error) {
	q := pq.Select("thing_urn", "timestamp", "value", "measurement_type", "measurement_type_unit", "location_id", "ST_X(coordinates::geometry) as lon", "ST_Y(coordinates::geometry) as lat", "metadata").
		From("measurements").
		Where("timestamp >= ?", query.Start).
		OrderBy("timestamp DESC").
		Limit(uint64(p.Limit + 1))

	// Use cursor otherwise end time
	if p.Cursor != "" {
		cursorTime, err := decodeCursor(p.Cursor)
		if err != nil {
			return nil, nil, err
		}
		q = q.Where("timestamp <= to_timestamp(?)", cursorTime)
	} else {
		q = q.Where("timestamp <= ?", query.End)
	}

	if len(query.Filters.ThingURNs) > 0 {
		q = q.Where(sq.Eq{"thing_urn": query.Filters.ThingURNs})
	}
	if len(query.Filters.MeasurementTypes) > 0 {
		q = q.Where(sq.Eq{"measurement_type": query.Filters.MeasurementTypes})
	}
	if len(query.Filters.LocationIDs) > 0 {
		q = q.Where(sq.Eq{"location_id": query.Filters.LocationIDs})
	}

	rows, err := q.RunWith(s.db).Query()
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	list := make([]measurements.Measurement, 0, p.Limit)
	var nextPage *measurements.Pagination
	for rows.Next() {
		var m measurements.Measurement
		err = rows.Scan(&m.ThingURN, &m.Timestamp, &m.Value, &m.MeasurementType, &m.MeasurementTypeUnit, &m.LocationID, &m.Coordinates[0], &m.Coordinates[1], &m.Metadata)
		if err != nil {
			return nil, nil, err
		}

		if len(list) < p.Limit {
			list = append(list, m)
		} else {
			ts := m.Timestamp.Unix()
			nextPage = &measurements.Pagination{
				Cursor: encodeCursor(p, uint64(ts)),
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

func encodeCursor(p measurements.Pagination, ts uint64) string {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, ts)
	return hex.EncodeToString(data)
}
