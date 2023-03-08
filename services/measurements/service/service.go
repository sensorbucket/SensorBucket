package service

//go:generate moq -pkg service_test -out mock_test.go . Store DatastreamFinderCreater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	deviceservice "sensorbucket.nl/sensorbucket/services/device/service"
)

var (
	ErrMissingDeviceInMeasurement    = errors.New("received measurement where device was not set, can't store")
	ErrMissingTimestampInMeasurement = errors.New("received measurement where timestamp was not set, can't store")
)

type Measurement struct {
	UplinkMessageID                 string                    `json:"uplink_message_id"`
	OrganisationID                  int                       `json:"organisation_id"`
	OrganisationName                string                    `json:"organisation_name"`
	OrganisationAddress             string                    `json:"organisation_address"`
	OrganisationZipcode             string                    `json:"organisation_zipcode"`
	OrganisationCity                string                    `json:"organisation_city"`
	OrganisationChamberOfCommerceID string                    `json:"organisation_chamber_of_commerce_id"`
	OrganisationHeadquarterID       string                    `json:"organisation_headquarter_id"`
	OrganisationArchiveTime         int                       `json:"organisation_archive_time"`
	OrganisationState               int                       `json:"organisation_state"` // TODO: Use enumerator
	DeviceID                        int64                     `json:"device_id"`
	DeviceCode                      string                    `json:"device_code"`
	DeviceDescription               string                    `json:"device_description"`
	DeviceLatitude                  *float64                  `json:"device_latitude"`
	DeviceLongitude                 *float64                  `json:"device_longitude"`
	DeviceAltitude                  *float64                  `json:"device_altitude"`
	DeviceLocationDescription       string                    `json:"device_location_description"`
	DeviceProperties                json.RawMessage           `json:"device_properties"`
	DeviceState                     deviceservice.DeviceState `json:"device_state"`
	SensorID                        int64                     `json:"sensor_id"`
	SensorCode                      string                    `json:"sensor_code"`
	SensorDescription               string                    `json:"sensor_description"`
	SensorExternalID                string                    `json:"sensor_external_id"`
	SensorProperties                json.RawMessage           `json:"sensor_properties"`
	SensorBrand                     string                    `json:"sensor_brand"`
	SensorArchiveTime               *int                      `json:"sensor_archive_time"`
	DatastreamID                    uuid.UUID                 `json:"datastream_id"`
	DatastreamDescription           string                    `json:"datastream_description"`
	DatastreamObservedProperty      string                    `json:"datastream_observed_property"`
	DatastreamUnitOfMeasurement     string                    `json:"datastream_unit_of_measurement"`
	MeasurementTimestamp            time.Time                 `json:"measurement_timestamp"`
	MeasurementValue                float64                   `json:"measurement_value"`
	MeasurementLatitude             *float64                  `json:"measurement_latitude"`
	MeasurementLongitude            *float64                  `json:"measurement_longitude"`
	MeasurementAltitude             *float64                  `json:"measurement_altitude"`
	MeasurementProperties           map[string]any            `json:"measurement_properties"`
}

func (m *Measurement) Validate() error {
	if m.DeviceID == 0 {
		return ErrMissingDeviceInMeasurement
	}
	if m.MeasurementTimestamp.IsZero() {
		return ErrMissingTimestampInMeasurement
	}
	// TODO: Add validation
	return nil
}

// QueryFilters represents the available filters for querying measurements
type QueryFilters struct {
	DeviceIDs   []string
	SensorCodes []string
	Datastream  string
}

// Query contains query information for a list of measurements
type Query struct {
	Start   time.Time
	End     time.Time
	Filters QueryFilters
}

// Pagination represents the pagination information for the measurements query.
type Pagination struct {
	Limit     int
	Skip      int
	Timestamp time.Time
}

// iService is an interface for the service's exported interface, it can be used as a developer reference
type iService interface {
	StoreMeasurement(Measurement) error
	StorePipelineMessage(context.Context, pipeline.Message) error
	QueryMeasurements(Query, Pagination) ([]Measurement, *Pagination, error)
}

// Ensure Service implements iService
var _ iService = (*Service)(nil)

// Store stores measurement data
type Store interface {
	DatastreamFinderCreater

	Insert(Measurement) error
	Query(Query, Pagination) ([]Measurement, *Pagination, error)
	ListDatastreams() ([]Datastream, error)
}

// Service is the measurement service which stores measurement data.
type Service struct {
	store Store
}

func New(store Store) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) StorePipelineMessage(ctx context.Context, msg pipeline.Message) error {
	// TODO: get organisation from context

	// Validate incoming message for completeness
	if msg.Device == nil {
		return ErrMissingDeviceInMeasurement
	}
	if len(msg.Measurements) == 0 {
		log.Printf("[warn] got pipeline message (%v) but it has no measurements\n", msg.ID)
		return nil
	}

	for _, measurement := range msg.Measurements {
		err := s.storePipelineMeasurement(msg, measurement)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) storePipelineMeasurement(msg pipeline.Message, m pipeline.Measurement) error {
	// Get sensor that produced this measurement
	dev := (*deviceservice.Device)(msg.Device)
	sensor, err := dev.GetSensorByExternalID(m.SensorExternalID)
	if err != nil {
		return fmt.Errorf("GetSensorByExternalID error for device: %d, sensor eID: %s, error: %w", dev.ID, m.SensorExternalID, err)
	}

	// Find or create datastream
	ds, err := FindOrCreateDatastream(sensor.ID, m.ObservedProperty, m.UnitOfMeasurement, s.store)
	if err != nil {
		return err
	}

	measurement := Measurement{
		UplinkMessageID: msg.ID,
		// TODO: Organisation...
		DeviceID:                  msg.Device.ID,
		DeviceCode:                msg.Device.Code,
		DeviceDescription:         msg.Device.Description,
		DeviceLatitude:            msg.Device.Latitude,
		DeviceLongitude:           msg.Device.Longitude,
		DeviceAltitude:            msg.Device.Altitude,
		DeviceLocationDescription: msg.Device.LocationDescription,
		DeviceProperties:          msg.Device.Properties,
		DeviceState:               msg.Device.State,

		SensorID:          sensor.ID,
		SensorCode:        sensor.Code,
		SensorDescription: sensor.Description,
		SensorExternalID:  sensor.ExternalID,
		SensorProperties:  sensor.Properties,
		SensorBrand:       sensor.Brand,
		SensorArchiveTime: sensor.ArchiveTime,

		DatastreamID:                ds.ID,
		DatastreamDescription:       ds.Description,
		DatastreamObservedProperty:  ds.ObservedProperty,
		DatastreamUnitOfMeasurement: ds.UnitOfMeasurement,

		MeasurementTimestamp:  time.UnixMilli(m.Timestamp),
		MeasurementValue:      m.Value,
		MeasurementLatitude:   msg.Device.Latitude,
		MeasurementLongitude:  msg.Device.Longitude,
		MeasurementAltitude:   msg.Device.Altitude,
		MeasurementProperties: m.Properties,
	}

	// Measurement location is either explicitly set or falls back to device location
	if m.Latitude != nil && m.Longitude != nil {
		measurement.MeasurementLatitude = m.Latitude
		measurement.MeasurementLongitude = m.Longitude
		measurement.MeasurementAltitude = m.Altitude
	}

	return s.StoreMeasurement(measurement)
}

func (s *Service) StoreMeasurement(m Measurement) error {
	if err := m.Validate(); err != nil {
		return fmt.Errorf("validation failed for measurement: %w", err)
	}

	return s.store.Insert(m)
}

func (s *Service) QueryMeasurements(q Query, p Pagination) ([]Measurement, *Pagination, error) {
	measurements, nextPage, err := s.store.Query(q, p)
	if err != nil {
		return nil, nil, err
	}
	return measurements, nextPage, nil
}

func (s *Service) ListDatastreams(ctx context.Context) ([]Datastream, error) {
	return s.store.ListDatastreams()
}
