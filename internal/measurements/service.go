package measurements

import (
	"encoding/json"
	"errors"
	"time"
)

// Measurement is a struct that contains the data for a measurement.
type Measurement struct {
	ThingURN            string          `json:"thing_urn"`
	Timestamp           time.Time       `json:"timestamp"`
	Value               float64         `json:"value"`
	MeasurementType     string          `json:"measurement_type"`
	MeasurementTypeUnit string          `json:"measurement_type_unit"`
	LocationID          *int            `json:"location_id"`
	Coordinates         [2]float64      `json:"coordinates"`
	Metadata            json.RawMessage `json:"metadata"`
}

func (m *Measurement) Validate() error {
	if m.ThingURN == "" {
		return errors.New("thing_urn is required")
	}
	if m.Timestamp.IsZero() {
		return errors.New("timestamp is required")
	}
	if m.MeasurementType == "" {
		return errors.New("measurement_type is required")
	}
	if m.MeasurementTypeUnit == "" {
		return errors.New("measurement_type_unit is required")
	}
	if len(m.Coordinates) != 2 {
		return errors.New("coordinates is required and must be a 2-element array (lon,lat)")
	}
	return nil
}

// QueryFilters represents the available filters for querying measurements
type QueryFilters struct {
	ThingURNs        []string
	LocationIDs      []int
	MeasurementTypes []string
}

// iService is an interface for the service's exported interface, it can be used as a developer reference
type iService interface {
	StoreMeasurement(*Measurement) error
	QueryMeasurements(start, end time.Time, filters QueryFilters) ([]Measurement, error)
}

// Ensure Service implements iService
var _ iService = (*Service)(nil)

// MeasurementStore stores measurement data
type MeasurementStore interface {
	Insert(*Measurement) error
	Query(start, end time.Time, filters QueryFilters) ([]Measurement, error)
}

// Service is the measurement service which stores measurement data.
type Service struct {
	store MeasurementStore
}

func New(store MeasurementStore) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) StoreMeasurement(m *Measurement) error {
	return s.store.Insert(m)
}

func (s *Service) QueryMeasurements(start, end time.Time, filters QueryFilters) ([]Measurement, error) {
	return s.store.Query(start, end, filters)
}
