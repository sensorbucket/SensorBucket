package measurements

//go:generate moq -pkg measurements_test -out mock_test.go . Store

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/internal/cleanupper"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	"sensorbucket.nl/sensorbucket/services/core/devices"
)

// Store stores measurement data
type Store interface {
	Query(context.Context, Filter, pagination.Request) (*pagination.Page[Measurement], error)
	ListDatastreams(context.Context, DatastreamFilter, pagination.Request) (*pagination.Page[Datastream], error)
	GetDatastream(ctx context.Context, id uuid.UUID, filter DatastreamFilter) (*Datastream, error)
	FindOrCreateDatastream(ctx context.Context, tenantID, sensorID int64, observedProperty, UnitOfMeasurement string) (*Datastream, error)
	StoreMeasurements(context.Context, []Measurement) error
}

// Service is the measurement service which stores measurement data.
type Service struct {
	store                Store
	systemArchiveTime    int
	keyClient            auth.JWKSClient
	measurementBatchChan chan Measurement
	measurementBatch     []Measurement
}

func New(store Store, systemArchiveTime, batchSize int, keyClient auth.JWKSClient) *Service {
	return &Service{
		store:                store,
		systemArchiveTime:    systemArchiveTime,
		keyClient:            keyClient,
		measurementBatch:     make([]Measurement, 0, batchSize),
		measurementBatchChan: make(chan Measurement, batchSize),
	}
}

func (s *Service) StartMeasurementBatchStorer(interval time.Duration) cleanupper.Shutdown {
	stop := make(chan struct{})
	done := make(chan struct{})
	t := time.NewTicker(interval)

	go func() {
		log.Println("Measurement service batch storer started")
		defer log.Println("Measurement service batch storer stopped!")
	outer:
		for {
			select {
			case <-stop:
				if err := s.CommitBatch(false); err != nil {
					log.Printf("error committing batch: %s\n", err.Error())
				}
				break outer
			case m := <-s.measurementBatchChan:
				s.measurementBatch = append(s.measurementBatch, m)
				if len(s.measurementBatch) == cap(s.measurementBatch) {
					if err := s.CommitBatch(false); err != nil {
						log.Printf("error committing batch: %s\n", err.Error())
					}
				}
			case <-t.C:
				if err := s.CommitBatch(false); err != nil {
					log.Printf("error committing batch: %s\n", err.Error())
				}
			}
		}
		close(done)
	}()

	return func(ctx context.Context) error {
		close(stop)
		<-done
		return nil
	}
}

func (s *Service) CommitBatch(collect bool) error {
	if len(s.measurementBatch) == 0 {
		if !collect || len(s.measurementBatchChan) == 0 {
			return nil
		}
		count := len(s.measurementBatchChan)
		for i := 0; i < count; i++ {
			s.measurementBatch = append(s.measurementBatch, <-s.measurementBatchChan)
		}
	}
	log.Printf("Committing %d measurements\n", len(s.measurementBatch))
	err := s.store.StoreMeasurements(context.Background(), s.measurementBatch)
	if err != nil {
		return fmt.Errorf("committing measurements failed: %w", err)
	}
	s.measurementBatch = s.measurementBatch[:0]
	return nil
}

func (s *Service) ProcessPipelineMessage(pmsg pipeline.Message) error {
	msg := PipelineMessage(pmsg)

	// Only error when internal error and not a business error
	ctx, err := msg.Authorize(s.keyClient)
	if err != nil {
		return err
	}
	if err := msg.Validate(); err != nil {
		return err
	}

	dev := (*devices.Device)(msg.Device)
	baseMeasurement := Measurement{
		UplinkMessageID:           msg.TracingID,
		OrganisationID:            int(msg.TenantID),
		DeviceID:                  msg.Device.ID,
		DeviceCode:                msg.Device.Code,
		DeviceDescription:         msg.Device.Description,
		DeviceLatitude:            msg.Device.Latitude,
		DeviceLongitude:           msg.Device.Longitude,
		DeviceAltitude:            msg.Device.Altitude,
		DeviceLocationDescription: msg.Device.LocationDescription,
		DeviceProperties:          msg.Device.Properties,
		DeviceState:               msg.Device.State,
		MeasurementLatitude:       msg.Device.Latitude,
		MeasurementLongitude:      msg.Device.Longitude,
		MeasurementAltitude:       msg.Device.Altitude,
		CreatedAt:                 time.Now(),
	}

	for _, m := range msg.Measurements {
		sensor, err := dev.GetSensorByExternalIDOrFallback(m.SensorExternalID)
		if err != nil {
			return fmt.Errorf("cannot get sensor: %w", err)
		}
		if sensor.ExternalID != m.SensorExternalID {
			m.ObservedProperty = m.SensorExternalID + "_" + m.ObservedProperty
		}

		archiveTimeDays, _ := lo.Coalesce(sensor.ArchiveTime, &s.systemArchiveTime) // msg.Organisation.ArchiveTime)

		ds, err := s.store.FindOrCreateDatastream(ctx, msg.TenantID, sensor.ID, m.ObservedProperty, m.UnitOfMeasurement)
		if err != nil {
			return err
		}

		measurement := baseMeasurement
		measurement.SensorID = sensor.ID
		measurement.SensorCode = sensor.Code
		measurement.SensorDescription = sensor.Description
		measurement.SensorExternalID = sensor.ExternalID
		measurement.SensorProperties = sensor.Properties
		measurement.SensorBrand = sensor.Brand
		measurement.SensorArchiveTime = sensor.ArchiveTime
		measurement.SensorIsFallback = sensor.IsFallback
		measurement.DatastreamID = ds.ID
		measurement.DatastreamDescription = ds.Description
		measurement.DatastreamObservedProperty = ds.ObservedProperty
		measurement.DatastreamUnitOfMeasurement = ds.UnitOfMeasurement
		measurement.MeasurementTimestamp = time.UnixMilli(m.Timestamp)
		measurement.MeasurementValue = m.Value
		measurement.MeasurementProperties = m.Properties
		measurement.MeasurementExpiration = time.UnixMilli(msg.ReceivedAt).Add(time.Duration(*archiveTimeDays) * 24 * time.Hour)

		// Fetch FoI info
		if sensor.FeatureOfInterest != nil {
			measurement.FeatureOfInterestID = &sensor.FeatureOfInterest.ID
			measurement.FeatureOfInterestName = &sensor.FeatureOfInterest.Name
			measurement.FeatureOfInterestDescription = &sensor.FeatureOfInterest.Description
			measurement.FeatureOfInterestEncodingType = &sensor.FeatureOfInterest.EncodingType
			measurement.FeatureOfInterestFeature = sensor.FeatureOfInterest.Feature
			measurement.FeatureOfInterestProperties = sensor.FeatureOfInterest.Properties
		}

		// Measurement location is either explicitly set or falls back to device location
		if m.Latitude != nil && m.Longitude != nil {
			measurement.MeasurementLatitude = m.Latitude
			measurement.MeasurementLongitude = m.Longitude
			measurement.MeasurementAltitude = m.Altitude
		}

		s.measurementBatchChan <- measurement
	}

	return nil
}

// Filter contains query information for a list of measurements
type Filter struct {
	Start               time.Time `url:"start,required"`
	End                 time.Time `url:"end,required"`
	SensorCodes         []string  `url:"sensor_codes"`
	Datastream          []string  `url:"datastream"`
	TenantID            []int64   `url:"tenant_id"`
	FeatureOfInterestID []int64   `url:"feature_of_interest_id"`
	ObservedProperty    []string  `url:"observed_property"`
}

func (s *Service) QueryMeasurements(ctx context.Context, f Filter, r pagination.Request) (*pagination.Page[Measurement], error) {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.READ_MEASUREMENTS}); err != nil {
		return nil, err
	}
	tenantID, err := auth.GetTenant(ctx)
	if err != nil {
		return nil, err
	}
	f.TenantID = []int64{tenantID}

	page, err := s.store.Query(ctx, f, r)
	if err != nil {
		return nil, err
	}
	return page, nil
}

type DatastreamFilter struct {
	Sensor           []int
	ObservedProperty []string
	TenantID         []int64
}

func (s *Service) ListDatastreams(ctx context.Context, filter DatastreamFilter, r pagination.Request) (*pagination.Page[Datastream], error) {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.READ_MEASUREMENTS}); err != nil {
		return nil, err
	}
	tenantID, err := auth.GetTenant(ctx)
	if err != nil {
		return nil, err
	}
	filter.TenantID = []int64{tenantID}

	return s.store.ListDatastreams(ctx, filter, r)
}

func (s *Service) GetDatastream(ctx context.Context, id uuid.UUID) (*Datastream, error) {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.READ_MEASUREMENTS}); err != nil {
		return nil, err
	}
	tenantID, err := auth.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	return s.store.GetDatastream(ctx, id, DatastreamFilter{TenantID: []int64{tenantID}})
}
