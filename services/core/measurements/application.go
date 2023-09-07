package measurements

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	"sensorbucket.nl/sensorbucket/services/core/devices"
)

// iService is an interface for the service's exported interface, it can be used as a developer reference
type iService interface {
	StoreMeasurement(Measurement) error
	StorePipelineMessage(context.Context, pipeline.Message) error
	QueryMeasurements(Filter, pagination.Request) (*pagination.Page[Measurement], error)
	ListDatastreams(ctx context.Context, filter DatastreamFilter, r pagination.Request) (*pagination.Page[Datastream], error)
	GetDatastream(ctx context.Context, id uuid.UUID) (*Datastream, error)
}

// Ensure Service implements iService
var _ iService = (*Service)(nil)

// Store stores measurement data
type Store interface {
	DatastreamFinderCreater

	Insert(Measurement) error
	Query(Filter, pagination.Request) (*pagination.Page[Measurement], error)
	ListDatastreams(DatastreamFilter, pagination.Request) (*pagination.Page[Datastream], error)
	GetDatastream(id uuid.UUID) (*Datastream, error)
}

// Service is the measurement service which stores measurement data.
type Service struct {
	store             Store
	systemArchiveTime int
	observers         []*observer
}

func New(store Store, systemArchiveTime int) *Service {
	return &Service{
		store:             store,
		systemArchiveTime: systemArchiveTime,
	}
}

func (s *Service) StoreMeasurement(m Measurement) error {
	if err := m.Validate(); err != nil {
		return fmt.Errorf("validation failed for measurement: %w", err)
	}

	return s.store.Insert(m)
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
	// Get sensor that produced this measurement by matching external_ids
	// If no sensor is found, check if there is a fallback sensor
	// If a fallback sensor exists, any observation will be prefixed with the original external_id
	// to avoid observation property collisions
	dev := (*devices.Device)(msg.Device)
	sensor, err := dev.GetSensorByExternalID(m.SensorExternalID)
	if errors.Is(err, devices.ErrSensorNotFound) {
		fmt.Printf("warning: no sensor found for external id '%s' on device id '%d' while storing pipeline measurements\n", m.SensorExternalID, msg.Device.ID)
		m.ObservedProperty = m.SensorExternalID + "_" + m.ObservedProperty
		sensor, err = dev.GetFallbackSensor()
		if err != nil {
			return fmt.Errorf("error getting fallback sensor: %w", err)
		}
	}
	if err != nil {
		return fmt.Errorf("error getting sensor by external ID '%s' for device id '%d': %w", m.SensorExternalID, msg.Device.ID, err)
	}

	// Find or create datastream
	ds, err := FindOrCreateDatastream(sensor.ID, m.ObservedProperty, m.UnitOfMeasurement, s.store)
	if err != nil {
		return err
	}

	// TODO: Get organisation archive time
	// Time is by default in days
	archiveTimeDays, _ := lo.Coalesce(sensor.ArchiveTime, &s.systemArchiveTime) // msg.Organisation.ArchiveTime)

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
		SensorIsFallback:  sensor.IsFallback,

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
		MeasurementExpiration: time.UnixMilli(msg.ReceivedAt).Add(time.Duration(*archiveTimeDays) * 24 * time.Hour),

		CreatedAt: time.Now(),
	}

	// Measurement location is either explicitly set or falls back to device location
	if m.Latitude != nil && m.Longitude != nil {
		measurement.MeasurementLatitude = m.Latitude
		measurement.MeasurementLongitude = m.Longitude
		measurement.MeasurementAltitude = m.Altitude
	}

	// Notify all observers to allow data streaming
	s.notifyObservers(measurement)

	return s.StoreMeasurement(measurement)
}

// Filter contains query information for a list of measurements
type Filter struct {
	Start       time.Time `url:",required"`
	End         time.Time `url:",required"`
	DeviceIDs   []string
	SensorCodes []string
	Datastream  []string
}

func (s *Service) QueryMeasurements(f Filter, r pagination.Request) (*pagination.Page[Measurement], error) {
	page, err := s.store.Query(f, r)
	if err != nil {
		return nil, err
	}
	return page, nil
}

func (s *Service) StreamMeasurements(f Filter) <-chan Measurement {
	// TODO: buffer measurement in slice instead of triggering ws every measurement
	stream := make(chan Measurement)
	s.observers = append(s.observers, &observer{
		filter:   f,
		listener: stream,
	})

	go func() {
		for {
			time.Sleep(time.Second * 5)
			fmt.Println("testing stream")
			for _, obs := range s.observers {
				fmt.Println("sending fake measurement...")
				obs.Notify(Measurement{
					MeasurementValue: 4324.54,
				})
			}
		}
	}()

	return stream
}

type DatastreamFilter struct {
	Sensor []int
}

func (s *Service) ListDatastreams(ctx context.Context, filter DatastreamFilter, r pagination.Request) (*pagination.Page[Datastream], error) {
	return s.store.ListDatastreams(filter, r)
}

func (s *Service) GetDatastream(ctx context.Context, id uuid.UUID) (*Datastream, error) {
	return s.store.GetDatastream(id)
}

func (s *Service) notifyObservers(measurement Measurement) {
	// Notify all observers to allow data streaming
	for _, observer := range s.observers {
		if observer.listener == nil {
			// the observer is marked for deletion
			fmt.Println("c is nil, continueing")
			continue
		}
		observer.Notify(measurement)
	}

	// TODO: will this trigger a concurrent write error when writing measurements to channels?
	s.observers = lo.Filter(s.observers, func(obs *observer, index int) bool {
		return obs.listener != nil
	})
	fmt.Println("resetting list")
}

type observer struct {
	filter   Filter
	measLock sync.Mutex
	listener chan Measurement
}

func (o *observer) Notify(measurement Measurement) {
	// TODO: Check if the measurement matches the given filter
	go func() {
		o.measLock.Lock()
		defer o.measLock.Unlock()

		fmt.Println("putting measurement in channel")
		// TODO: how to handle order of measurements with go routine??
		t := time.NewTicker(time.Second * 3)
		select {
		case o.listener <- measurement:
		case <-t.C:
			log.Println("[Warning] measurement not accepted by data observer within 3 seconds, closing stream")
			close(o.listener)
			o.listener = nil
		}
	}()
}
