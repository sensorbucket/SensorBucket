package measurements

import "time"

// Measurement is a struct that contains the data for a measurement.
type Measurement struct {
	Timestamp   time.Time `json:"timestamp,omitempty"`
	Serial      string    `json:"serial,omitempty"`
	Measurement float32   `json:"measurement,omitempty"`
}

// iService is an interface for the service's exported interface, it can be used as a developer reference
type iService interface {
	StoreMeasurement(*Measurement) error
}

// Ensure Service implements iService
var _ iService = (*Service)(nil)

// MeasurementStore stores measurement data
type MeasurementStore interface {
	Insert(*Measurement) error
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
