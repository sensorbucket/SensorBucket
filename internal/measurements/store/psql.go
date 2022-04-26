package store

import (
	"database/sql"
	"time"

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

func (s *MeasurementStorePSQL) Query(start, end time.Time, filters measurements.QueryFilters) ([]measurements.Measurement, error) {
	q := pq.Select("thing_urn", "timestamp", "value", "measurement_type", "measurement_type_unit", "location_id", "ST_X(coordinates::geometry) as lon", "ST_Y(coordinates::geometry) as lat").
		From("measurements").
		Where("timestamp >= ? AND timestamp <= ?", start, end)

	if len(filters.ThingURNs) > 0 {
		q = q.Where(sq.Eq{"thing_urn": filters.ThingURNs})
	}
	if len(filters.MeasurementTypes) > 0 {
		q = q.Where(sq.Eq{"measurement_type": filters.MeasurementTypes})
	}
	if len(filters.LocationIDs) > 0 {
		q = q.Where(sq.Eq{"location_id": filters.LocationIDs})
	}

	rows, err := q.RunWith(s.db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []measurements.Measurement
	for rows.Next() {
		var m measurements.Measurement
		err = rows.Scan(&m.ThingURN, &m.Timestamp, &m.Value, &m.MeasurementType, &m.MeasurementTypeUnit, &m.LocationID, &m.Coordinates[0], &m.Coordinates[1])
		if err != nil {
			return nil, err
		}
		list = append(list, m)
	}

	return list, nil
}
