package devices

//go:generate moq -pkg devices_test -out mock_test.go . DeviceStore SensorGroupStore

import (
	"context"
	"encoding/json"
	"fmt"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/services/core/featuresofinterest"
)

type DeviceStore interface {
	List(context.Context, DeviceFilter, pagination.Request) (*pagination.Page[Device], error)
	ListInBoundingBox(context.Context, DeviceFilter, pagination.Request) (*pagination.Page[Device], error)
	ListInRange(context.Context, DeviceFilter, pagination.Request) (*pagination.Page[Device], error)
	ListSensors(context.Context, pagination.Request) (*pagination.Page[Sensor], error)
	Find(ctx context.Context, id int64) (*Device, error)
	Save(ctx context.Context, dev *Device) error
	Delete(ctx context.Context, dev *Device) error
	GetSensor(ctx context.Context, id int64) (*Sensor, error)
}

type SensorGroupStore interface {
	Save(group *SensorGroup) error
	Delete(id int64) error
	List(tenantID int64, p pagination.Request) (*pagination.Page[SensorGroup], error)
	Get(id int64, tenantID int64) (*SensorGroup, error)
}

type Service struct {
	store                    DeviceStore
	sensorGroupStore         SensorGroupStore
	featureOfInterestService *featuresofinterest.Service
}

func New(store DeviceStore, sensorGroupStore SensorGroupStore, featureOfInterestService *featuresofinterest.Service) *Service {
	return &Service{
		store:                    store,
		sensorGroupStore:         sensorGroupStore,
		featureOfInterestService: featureOfInterestService,
	}
}

type DeviceFilter struct {
	BoundingBoxFilter
	RangeFilter
	ID         []int64
	Code       []string
	Sensor     []int64
	Properties json.RawMessage `json:"properties"`
	OwnerID    int64
}
type BoundingBoxFilter struct {
	North *float64 `json:"north"`
	West  *float64 `json:"west"`
	South *float64 `json:"south"`
	East  *float64 `json:"east"`
}
type RangeFilter struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	Distance  *float64 `json:"range"`
}

func (f DeviceFilter) HasBoundingBox() bool {
	return f.North != nil && f.West != nil && f.East != nil && f.South != nil
}

func (f DeviceFilter) HasRange() bool {
	return f.Latitude != nil && f.Longitude != nil && f.Distance != nil
}

func (s *Service) ListDevices(ctx context.Context, filter DeviceFilter, p pagination.Request) (*pagination.Page[Device], error) {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.READ_DEVICES}); err != nil {
		return nil, err
	}

	if filter.HasBoundingBox() {
		return s.store.ListInBoundingBox(ctx, filter, p)
	}
	if filter.HasRange() {
		return s.store.ListInRange(ctx, filter, p)
	}
	return s.store.List(ctx, filter, p)
}

func (s *Service) CreateDevice(ctx context.Context, dto NewDeviceOpts) (*Device, error) {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_DEVICES}); err != nil {
		return nil, err
	}
	tenantID, err := auth.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	dev, err := NewDevice(tenantID, dto)
	if err != nil {
		return nil, err
	}
	if err := s.store.Save(ctx, dev); err != nil {
		return nil, err
	}
	return dev, nil
}

func (s *Service) GetDevice(ctx context.Context, id int64) (*Device, error) {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.READ_DEVICES}); err != nil {
		return nil, err
	}

	dev, err := s.store.Find(ctx, id)
	if err != nil {
		return nil, err
	}
	return dev, nil
}

type NewSensorDTO struct {
	Code                string          `json:"code"`
	Brand               string          `json:"brand"`
	Description         string          `json:"description"`
	ExternalID          string          `json:"external_id"`
	FeatureOfInterestID int64           `json:"feature_of_interest_id"`
	Properties          json.RawMessage `json:"properties"`
	ArchiveTime         *int            `json:"archive_time"`
	IsFallback          bool            `json:"is_fallback"`
}

func (s *Service) AddSensor(ctx context.Context, dev *Device, dto NewSensorDTO) error {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_DEVICES}); err != nil {
		return err
	}

	opts := NewSensorOpts{
		Code:        dto.Code,
		Brand:       dto.Brand,
		Description: dto.Description,
		ExternalID:  dto.ExternalID,
		Properties:  dto.Properties,
		ArchiveTime: dto.ArchiveTime,
		IsFallback:  dto.IsFallback,
	}
	if dto.FeatureOfInterestID > 0 {
		feature, err := s.featureOfInterestService.GetFeatureOfInterest(ctx, dto.FeatureOfInterestID)
		if err != nil {
			return fmt.Errorf("in AddSensor: could not get feature of interest: %w", err)
		}
		opts.FeatureOfInterest = feature
	}

	if err := dev.AddSensor(opts); err != nil {
		return err
	}
	if err := s.store.Save(ctx, dev); err != nil {
		return err
	}
	return nil
}

func (s *Service) DeleteSensor(ctx context.Context, dev *Device, sensor *Sensor) error {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_DEVICES}); err != nil {
		return err
	}

	if err := dev.DeleteSensorByID(sensor.ID); err != nil {
		return err
	}
	if err := s.store.Save(ctx, dev); err != nil {
		return err
	}
	return nil
}

type UpdateDeviceOpts struct {
	Description         *string         `json:"description"`
	Longitude           *float64        `json:"longitude"`
	Latitude            *float64        `json:"latitude"`
	Altitude            *float64        `json:"altitude"`
	LocationDescription *string         `json:"location_description"`
	Properties          json.RawMessage `json:"properties"`
	State               *DeviceState    `json:"state"`
}

func (s *Service) UpdateDevice(ctx context.Context, dev *Device, opt UpdateDeviceOpts) error {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_DEVICES}); err != nil {
		return err
	}

	if opt.Description != nil {
		dev.Description = *opt.Description
	}
	if opt.Latitude != nil {
		dev.Latitude = opt.Latitude
	}
	if opt.Longitude != nil {
		dev.Longitude = opt.Longitude
	}
	if opt.Altitude != nil {
		dev.Altitude = opt.Altitude
	}
	if opt.LocationDescription != nil {
		dev.LocationDescription = *opt.LocationDescription
	}
	if opt.Properties != nil {
		dev.Properties = opt.Properties
	}
	if opt.State != nil {
		dev.State = *opt.State
	}
	if err := s.store.Save(ctx, dev); err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteDevice(ctx context.Context, dev *Device) error {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_DEVICES}); err != nil {
		return err
	}

	if err := s.store.Delete(ctx, dev); err != nil {
		return err
	}
	return nil
}

func (s *Service) ListSensors(ctx context.Context, p pagination.Request) (*pagination.Page[Sensor], error) {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.READ_DEVICES}); err != nil {
		return nil, err
	}

	return s.store.ListSensors(ctx, p)
}

func (s *Service) GetSensor(ctx context.Context, id int64) (*Sensor, error) {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.READ_DEVICES}); err != nil {
		return nil, err
	}

	return s.store.GetSensor(ctx, id)
}

type UpdateSensorOpts struct {
	Description         *string         `json:"description"`
	Brand               *string         `json:"brand"`
	ArchiveTime         *int            `json:"archive_time"`
	ExternalID          *string         `json:"external_id"`
	IsFallback          *bool           `json:"is_fallback"`
	Properties          json.RawMessage `json:"properties"`
	FeatureOfInterestID *int64          `json:"feature_of_interest_id"`
}

func (s *Service) UpdateSensor(ctx context.Context, device *Device, sensor *Sensor, opt UpdateSensorOpts) error {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_DEVICES}); err != nil {
		return err
	}

	if opt.Description != nil {
		sensor.Description = *opt.Description
	}
	if opt.Brand != nil {
		sensor.Brand = *opt.Brand
	}
	if opt.ArchiveTime != nil {
		if *opt.ArchiveTime == 0 {
			sensor.ArchiveTime = nil
		} else {
			sensor.ArchiveTime = opt.ArchiveTime
		}
	}
	if opt.ExternalID != nil {
		sensor.ExternalID = *opt.ExternalID
	}
	if opt.IsFallback != nil {
		sensor.IsFallback = *opt.IsFallback
	}
	if opt.Properties != nil {
		sensor.Properties = opt.Properties
	}
	if opt.FeatureOfInterestID != nil {
		if *opt.FeatureOfInterestID == 0 {
			sensor.FeatureOfInterest = nil
		} else {
			feature, err := s.featureOfInterestService.GetFeatureOfInterest(ctx, *opt.FeatureOfInterestID)
			if err != nil {
				return fmt.Errorf("in UpdateSensor: could not get feature of interest: %w", err)
			}
			sensor.FeatureOfInterest = feature
		}
	}

	if err := device.UpdateSensor(sensor); err != nil {
		return err
	}

	if err := s.store.Save(ctx, device); err != nil {
		return err
	}

	return nil
}

func (s *Service) CreateSensorGroup(ctx context.Context, name, description string) (*SensorGroup, error) {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_DEVICES}); err != nil {
		return nil, err
	}
	tenantID, err := auth.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	group, err := NewSensorGroup(tenantID, name, description)
	if err != nil {
		return nil, fmt.Errorf("create sensor group failed: %w", err)
	}
	if err := s.sensorGroupStore.Save(group); err != nil {
		return nil, fmt.Errorf("could not store sensor group: %w", err)
	}
	return group, nil
}

func (s *Service) ListSensorGroups(ctx context.Context, p pagination.Request) (*pagination.Page[SensorGroup], error) {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.READ_DEVICES}); err != nil {
		return nil, err
	}
	tenantID, err := auth.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	return s.sensorGroupStore.List(tenantID, p)
}

func (s *Service) GetSensorGroup(ctx context.Context, id int64) (*SensorGroup, error) {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.READ_DEVICES}); err != nil {
		return nil, err
	}
	tenantID, err := auth.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	return s.sensorGroupStore.Get(id, tenantID)
}

func (s *Service) AddSensorToSensorGroup(ctx context.Context, groupID, sensorID int64) error {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_DEVICES}); err != nil {
		return err
	}

	group, err := s.GetSensorGroup(ctx, groupID)
	if err != nil {
		return err
	}
	sensor, err := s.GetSensor(ctx, sensorID)
	if err != nil {
		return err
	}

	err = group.Add(sensor)
	if err != nil {
		return fmt.Errorf("could not add sensor to sensor group: %w", err)
	}

	if err := s.sensorGroupStore.Save(group); err != nil {
		return err
	}
	return nil
}

func (s *Service) DeleteSensorFromSensorGroup(ctx context.Context, groupID, sensorID int64) error {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_DEVICES}); err != nil {
		return err
	}

	group, err := s.GetSensorGroup(ctx, groupID)
	if err != nil {
		return err
	}
	sensor, err := s.GetSensor(ctx, sensorID)
	if err != nil {
		return err
	}

	err = group.Remove(sensor.ID)
	if err != nil {
		return fmt.Errorf("could not remove sensor from sensor group: %w", err)
	}

	if err := s.sensorGroupStore.Save(group); err != nil {
		return err
	}
	return nil
}

func (s *Service) DeleteSensorGroup(ctx context.Context, group *SensorGroup) error {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_DEVICES}); err != nil {
		return err
	}

	return s.sensorGroupStore.Delete(group.ID)
}

type UpdateSensorGroupOpts struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func (s *Service) UpdateSensorGroup(ctx context.Context, group *SensorGroup, opts UpdateSensorGroupOpts) error {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.WRITE_DEVICES}); err != nil {
		return err
	}

	if opts.Name != nil {
		err := group.SetName(*opts.Name)
		if err != nil {
			return err
		}
	}

	if opts.Description != nil {
		group.Description = *opts.Description
	}

	err := s.sensorGroupStore.Save(group)
	if err != nil {
		return fmt.Errorf("update sensor group, could not save: %w", err)
	}

	return nil
}
