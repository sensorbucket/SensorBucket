package store

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"sensorbucket.nl/internal/measurements"
)

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

	_, err := s.db.Exec(`INSERT INTO measurements (
			thing_urn,
			timestamp,
			value,
			measurement_type,
			measurement_type_unit,
			location_id,
			ST_SetSRID(ST_MakePoint($7, $8),4326)
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		m.ThingURN,
		m.Timestamp,
		m.Value,
		m.MeasurementType,
		m.MeasurementTypeUnit,
		locID,
		m.Coordinates[0],
		m.Coordinates[1],
	)
	return err
}
