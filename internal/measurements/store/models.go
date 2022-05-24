package store

import (
	"database/sql"
	"encoding/json"
	"time"

	"sensorbucket.nl/internal/measurements"
)

type model struct {
	ThingURN            string          `db:"thing_urn"`
	Timestamp           time.Time       `db:"timestamp"`
	Value               float64         `db:"value"`
	MeasurementType     string          `db:"measurement_type"`
	MeasurementTypeUnit string          `db:"measurement_type_unit"`
	Longitude           sql.NullFloat64 `db:"longitude"`
	Latitude            sql.NullFloat64 `db:"latitude"`
	LocationID          sql.NullInt64   `db:"location_id"`
	LocationName        sql.NullString  `db:"location_name"`
	LocationLongitude   sql.NullFloat64 `db:"location_longitude"`
	LocationLatitude    sql.NullFloat64 `db:"location_latitude"`
	Metadata            json.RawMessage `db:"metadata"`
}

func toModel(m measurements.Measurement) model {
	var locID sql.NullInt64
	if m.LocationID != nil {
		locID.Int64 = int64(*m.LocationID)
		locID.Valid = true
	}

	var locName sql.NullString
	if m.LocationName != nil {
		locName.String = string(*m.LocationName)
		locName.Valid = true
	}

	var lng sql.NullFloat64
	if m.Longitude != nil {
		lng.Float64 = *m.Longitude
		lng.Valid = true
	}

	var lat sql.NullFloat64
	if m.Latitude != nil {
		lat.Float64 = *m.Latitude
		lat.Valid = true
	}

	var locLng sql.NullFloat64
	if m.LocationLongitude != nil {
		locLng.Float64 = *m.LocationLongitude
		locLng.Valid = true
	}

	var locLat sql.NullFloat64
	if m.LocationLatitude != nil {
		locLat.Float64 = *m.LocationLatitude
		locLat.Valid = true
	}

	return model{
		ThingURN:            m.ThingURN,
		Timestamp:           m.Timestamp,
		Value:               m.Value,
		MeasurementType:     m.MeasurementType,
		MeasurementTypeUnit: m.MeasurementTypeUnit,
		LocationID:          locID,
		LocationName:        locName,
		Longitude:           lng,
		Latitude:            lat,
		LocationLongitude:   locLng,
		LocationLatitude:    locLat,
		Metadata:            m.Metadata,
	}
}

func (m model) toValue() measurements.Measurement {
	measurement := measurements.Measurement{
		IntermediateMeasurement: measurements.IntermediateMeasurement{
			ThingURN:            m.ThingURN,
			Timestamp:           m.Timestamp,
			Value:               m.Value,
			MeasurementType:     m.MeasurementType,
			MeasurementTypeUnit: m.MeasurementTypeUnit,
			Metadata:            m.Metadata,
		},
	}

	if m.Longitude.Valid {
		measurement.Longitude = &m.Longitude.Float64
	}
	if m.Latitude.Valid {
		measurement.Latitude = &m.Latitude.Float64
	}
	if m.LocationID.Valid {
		measurement.LocationID = &m.LocationID.Int64
	}
	if m.LocationName.Valid {
		measurement.LocationName = &m.LocationName.String
	}
	if m.LocationLongitude.Valid {
		measurement.LocationLongitude = &m.LocationLongitude.Float64
	}
	if m.LocationLatitude.Valid {
		measurement.LocationLatitude = &m.LocationLatitude.Float64
	}

	return measurement
}
