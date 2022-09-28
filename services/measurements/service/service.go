package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"
)

var (
	ErrLocationNotFound = errors.New("location not found")
)

type Measurement struct {
	UplinkMessageID     string          `json:"uplink_message_id"`
	DeviceID            int             `json:"device_id"`
	DeviceCode          string          `json:"device_code"`
	DeviceDescription   string          `json:"device_description"`
	DeviceConfiguration json.RawMessage `json:"device_configuration"`
	Timestamp           time.Time       `json:"timestamp"`
	Value               float64         `json:"value"`
	MeasurementType     string          `json:"measurement_type"`
	MeasurementTypeUnit string          `json:"measurement_type_unit"`
	Metadata            json.RawMessage `json:"metadata"`
	Longitude           *float64        `json:"longitude"`
	Latitude            *float64        `json:"latitude"`
	LocationID          *int64          `json:"location_id"`
	LocationName        *string         `json:"location_name"`
	LocationLongitude   *float64        `json:"location_longitude"`
	LocationLatitude    *float64        `json:"location_latitude"`
	SensorCode          *string         `json:"sensor_code"`
	SensorDescription   *string         `json:"sensor_description"`
	SensorExternalID    *string         `json:"sensor_external_id"`
	SensorConfiguration json.RawMessage `json:"sensor_configuration"`
}

func (m *Measurement) Validate() error {
	if m.DeviceID == 0 {
		return errors.New("Pipeline message must have a device attached to it")
	}
	if m.Timestamp.IsZero() {
		return errors.New("timestamp is required")
	}
	if m.MeasurementType == "" {
		return errors.New("measurement_type is required")
	}
	//if m.MeasurementTypeUnit == "" {
	//	return errors.New("measurement_type_unit is required")
	//}
	return nil
}

// QueryFilters represents the available filters for querying measurements
type QueryFilters struct {
	DeviceIDs        []string
	SensorCodes      []string
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
	StoreMeasurement(Measurement) error
	QueryMeasurements(Query, Pagination) ([]Measurement, *Pagination, error)
}

// Ensure Service implements iService
var _ iService = (*Service)(nil)

// MeasurementStore stores measurement data
type MeasurementStore interface {
	Insert(Measurement) error
	Query(Query, Pagination) ([]Measurement, *Pagination, error)
}

type LocationData struct {
	ID        int64
	Name      string
	Longitude float64
	Latitude  float64
}

// LocationService is used to fetch location for an asset
type LocationService interface {
	FindLocationID(thingURN string) (LocationData, error)
}

// Service is the measurement service which stores measurement data.
type Service struct {
	store     MeasurementStore
	locations LocationService
}

func New(store MeasurementStore) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) StoreMeasurement(m Measurement) error {
	log.Printf("Inserting measurements: %+v\n", m)
	if err := m.Validate(); err != nil {
		return fmt.Errorf("validation failed for measurement: %s", err)
	}

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
