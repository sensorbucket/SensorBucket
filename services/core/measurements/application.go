package measurements

import (
	"context"
	"time"

	"github.com/google/uuid"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/auth"
)

// Store stores measurement data
type Store interface {
	Query(Filter, pagination.Request) (*pagination.Page[Measurement], error)
	ListDatastreams(DatastreamFilter, pagination.Request) (*pagination.Page[Datastream], error)
	GetDatastream(id uuid.UUID, filter DatastreamFilter) (*Datastream, error)
}

type MeasurementStoreBuilder interface {
	Begin() (MeasurementStorer, error)
}
type MeasurementStorer interface {
	GetDatastream(tenantID, sensorID int64, observedProperty, unitOfMeasurement string) (*Datastream, error)
	AddMeasurements([]Measurement) error
	Finish() error
}

// Service is the measurement service which stores measurement data.
type Service struct {
	store             Store
	measurementStore  MeasurementStoreBuilder
	systemArchiveTime int
	keyClient         auth.JWKSClient
}

func New(store Store, measurementStore MeasurementStoreBuilder, systemArchiveTime int, keyClient auth.JWKSClient) *Service {
	return &Service{
		store:             store,
		measurementStore:  measurementStore,
		systemArchiveTime: systemArchiveTime,
		keyClient:         keyClient,
	}
}

func (s *Service) ProcessPipelineMessage(msg *PipelineMessage) error {
	// Only error when internal error and not a business error
	_, err := msg.Authorize(s.keyClient)
	if err != nil {
		return err
	}
	if err := msg.Validate(); err != nil {
		return err
	}

	storer, err := s.measurementStore.Begin()
	if err != nil {
		return err
	}

	measurements, err := buildMeasurements(msg, storer, s.systemArchiveTime)
	if err != nil {
		return err
	}

	if err := storer.AddMeasurements(measurements); err != nil {
		return err
	}

	if err := storer.Finish(); err != nil {
		return err
	}
	return nil
}

// Filter contains query information for a list of measurements
type Filter struct {
	Start       time.Time `url:",required"`
	End         time.Time `url:",required"`
	DeviceIDs   []string
	SensorCodes []string
	Datastream  []string
	TenantID    []int64
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

	page, err := s.store.Query(f, r)
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

	return s.store.ListDatastreams(filter, r)
}

func (s *Service) GetDatastream(ctx context.Context, id uuid.UUID) (*Datastream, error) {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.READ_MEASUREMENTS}); err != nil {
		return nil, err
	}
	tenantID, err := auth.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	return s.store.GetDatastream(id, DatastreamFilter{TenantID: []int64{tenantID}})
}
