package measurements

import (
	"encoding/json"
	"errors"
	"fmt"
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

// Query contains query information for a list of measurements
type Query struct {
	Start   time.Time
	End     time.Time
	Filters QueryFilters
}

// Pagination represents the pagination information for the measurements query.
type Pagination struct {
	Limit  int
	Cursor string
}

// iService is an interface for the service's exported interface, it can be used as a developer reference
type iService interface {
	StoreMeasurement(*Measurement) error
	QueryMeasurements(Query, Pagination) ([]Measurement, *Pagination, error)
}

// Ensure Service implements iService
var _ iService = (*Service)(nil)

// MeasurementStore stores measurement data
type MeasurementStore interface {
	Insert(*Measurement) error
	Query(Query, Pagination) ([]Measurement, *Pagination, error)
}

// LocationService is used to fetch location for an asset
type LocationService interface {
	FindLocationID(thingURN string) (*int, error)
}

// Service is the measurement service which stores measurement data.
type Service struct {
	store     MeasurementStore
	locations LocationService
}

func New(store MeasurementStore, locs LocationService) *Service {
	return &Service{
		store:     store,
		locations: locs,
	}
}

func (s *Service) StoreMeasurement(m *Measurement) error {
	if err := m.Validate(); err != nil {
		return fmt.Errorf("validation failed for measurement: %s", err)
	}

	locID, err := s.locations.FindLocationID(m.ThingURN)
	if err != nil {
		return fmt.Errorf("failed to find location for thing %s: %s", m.ThingURN, err)
	}
	m.LocationID = locID

	return s.store.Insert(m)
}

func (s *Service) QueryMeasurements(q Query, p Pagination) ([]Measurement, *Pagination, error) {
	measurements, nextPage, err := s.store.Query(q, p)
	if err != nil {
		return nil, nil, err
	}
	if nextPage != nil {
		nextPage.Limit = p.Limit
	}
	return measurements, nextPage, nil
}
