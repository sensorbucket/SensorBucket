package service

//go:generate moq -pkg service_test -out mock_test.go . Store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	deviceservice "sensorbucket.nl/sensorbucket/services/device/service"
)

var (
	ErrMissingDeviceInMeasurement    = errors.New("received measurement where device was not set, can't store")
	ErrMissingTimestampInMeasurement = errors.New("received measurement where timestamp was not set, can't store")
)

type Measurement struct {
	UplinkMessageID           string          `json:"uplink_message_id"`
	OrganisationID            int             `json:"organisation_id"`
	OrganisationName          string          `json:"organisation_name"`
	OrganisationAddress       string          `json:"organisation_address"`
	OrganisationZipcode       string          `json:"organisation_zipcode"`
	OrganisationCity          string          `json:"organisation_city"`
	OrganisationCoC           string          `json:"organisation_coc"`
	OrganisationLocationCoC   string          `json:"orgnisation_location_coc"`
	DeviceID                  int64           `json:"device_id"`
	DeviceCode                string          `json:"device_code"`
	DeviceDescription         string          `json:"device_description"`
	DeviceLatitude            *float64        `json:"device_latitude"`
	DeviceLongitude           *float64        `json:"device_longitude"`
	DeviceLocationDescription string          `json:"device_location_description"`
	DeviceConfiguration       json.RawMessage `json:"device_configuration"`
	SensorID                  int64           `json:"sensor_id"`
	SensorCode                string          `json:"sensor_code"`
	SensorTypeID              int64           `json:"sensor_type_id"`
	SensorTypeDescription     string          `json:"sensor_type_description"`
	SensorGoalID              int64           `json:"sensor_goal_id"`
	SensorGoalName            string          `json:"sensor_goal_name"`
	SensorDescription         string          `json:"sensor_description"`
	SensorExternalID          string          `json:"sensor_external_id"`
	SensorConfig              json.RawMessage `json:"sensor_config"`
	SensorBrand               string          `json:"sensor_brand"`
	MeasurementType           string          `json:"measurement_type"`
	MeasurementUnit           string          `json:"measurement_unit"`
	MeasurementTimestamp      time.Time       `json:"measurement_timestamp"`
	MeasurementValue          float64         `json:"measurement_value"`
	MeasurementValuePrefix    string          `json:"measurement_value_prefix"`
	MeasurementValueFactor    int             `json:"measurement_value_factor"`
	MeasurementLatitude       *float64        `json:"measurement_latitude"`
	MeasurementLongitude      *float64        `json:"measurement_longitude"`
	MeasurementMetadata       map[string]any  `json:"measurement_metadata"`
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
	DeviceIDs        []string
	SensorCodes      []string
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
	Insert(Measurement) error
	Query(Query, Pagination) ([]Measurement, *Pagination, error)
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
	measurement := Measurement{
		UplinkMessageID: msg.ID,
		// TODO: Organisation...
		DeviceID:                  msg.Device.ID,
		DeviceCode:                msg.Device.Code,
		DeviceDescription:         msg.Device.Description,
		DeviceLatitude:            msg.Device.Latitude,
		DeviceLongitude:           msg.Device.Longitude,
		DeviceLocationDescription: msg.Device.LocationDescription,
		DeviceConfiguration:       msg.Device.Configuration,

		MeasurementType:      m.MeasurementType,
		MeasurementUnit:      m.MeasurementUnit,
		MeasurementTimestamp: time.UnixMilli(m.Timestamp),
		MeasurementValue:     m.MeasurementValue,
		// Prefix?!?
		MeasurementValueFactor: m.MeasurementValueFactor,
		MeasurementLatitude:    msg.Device.Latitude,
		MeasurementLongitude:   msg.Device.Longitude,
		MeasurementMetadata:    m.MeasurementMetadata,
	}

	// Measurement location is either explicitly set or falls back to device location
	if m.MeasurementLatitude != nil && m.MeasurementLongitude != nil {
		measurement.MeasurementLatitude = m.MeasurementLatitude
		measurement.MeasurementLongitude = m.MeasurementLongitude
	}

	// Get sensor
	dev := (*deviceservice.Device)(msg.Device)
	sensor, err := dev.GetSensorByExternalID(m.SensorExternalID)
	if err != nil {
		return err
	}

	measurement.SensorID = sensor.ID
	measurement.SensorCode = sensor.Code
	measurement.SensorTypeID = sensor.Type
	measurement.SensorTypeDescription = ""
	measurement.SensorGoalID = sensor.Goal
	measurement.SensorGoalName = ""
	measurement.SensorDescription = sensor.Description
	measurement.SensorExternalID = sensor.ExternalID
	measurement.SensorConfig = sensor.Configuration
	measurement.SensorBrand = sensor.Brand

	return s.StoreMeasurement(measurement)
}

func (s *Service) StoreMeasurement(m Measurement) error {
	log.Printf("Inserting measurements: %+v\n", m)
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
