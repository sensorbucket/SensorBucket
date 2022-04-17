package store

import (
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

var SQL_INSERT = "INSERT INTO measurements (timestamp, subject, value) VALUES ($1, $2, $3)"

func (s *MeasurementStorePSQL) Insert(m *measurements.Measurement) error {
	if _, err := s.db.Exec(SQL_INSERT, m.Timestamp, m.Serial, m.Measurement); err != nil {
		return err
	}
	return nil
}

func (s *MeasurementStorePSQL) Migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS measurements (
			timestamp TIMESTAMP NOT NULL,
			subject VARCHAR(255) NOT NULL,
			value FLOAT NOT NULL,
			PRIMARY KEY (timestamp, subject)
		);
		SELECT create_hypertable('measurements', 'timestamp');
	`)

	return err
}
