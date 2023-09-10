package devices

//go:generate moq -pkg devices_test -out mock_test.go . DeviceStore SensorGroupStore

import (
	"context"
	"encoding/json"
	"fmt"

	"sensorbucket.nl/sensorbucket/internal/pagination"
)

type DeviceStore interface {
	List(DeviceFilter, pagination.Request) (*pagination.Page[Device], error)
	ListInBoundingBox(DeviceFilter, pagination.Request) (*pagination.Page[Device], error)
	ListInRange(DeviceFilter, pagination.Request) (*pagination.Page[Device], error)
	ListSensors(pagination.Request) (*pagination.Page[Sensor], error)
	Find(id int64) (*Device, error)
	Save(dev *Device) error
	Delete(dev *Device) error
	GetSensor(id int64) (*Sensor, error)
}

type SensorGroupStore interface {
	Save(group *SensorGroup) error
	Delete(id int64) error
	List(p pagination.Request) (*pagination.Page[SensorGroup], error)
	Get(id int64) (*SensorGroup, error)
}

type Service struct {
	store            DeviceStore
	sensorGroupStore SensorGroupStore
}

func New(store DeviceStore, sensorGroupStore SensorGroupStore) *Service {
	return &Service{
		store:            store,
		sensorGroupStore: sensorGroupStore,
	}
}

type DeviceFilter struct {
	BoundingBoxFilter
	RangeFilter
	ID         []int64
	Sensor     []int64
	Properties json.RawMessage `json:"properties"`
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
	if filter.HasBoundingBox() {
		return s.store.ListInBoundingBox(filter, p)
	}
	if filter.HasRange() {
		return s.store.ListInRange(filter, p)
	}
	return s.store.List(filter, p)
}

func (s *Service) CreateDevice(ctx context.Context, dto NewDeviceOpts) (*Device, error) {
	dev, err := NewDevice(dto)
	if err != nil {
		return nil, err
	}
	if err := s.store.Save(dev); err != nil {
		return nil, err
	}
	return dev, nil
}

func (s *Service) GetDevice(ctx context.Context, id int64) (*Device, error) {
	dev, err := s.store.Find(id)
	if err != nil {
		return nil, err
	}
	return dev, nil
}

type NewSensorDTO struct {
	Code        string          `json:"code"`
	Brand       string          `json:"brand"`
	GoalID      int64           `json:"goal_id"`
	TypeID      int64           `json:"type_id"`
	Description string          `json:"description"`
	ExternalID  string          `json:"external_id"`
	Properties  json.RawMessage `json:"properties"`
	ArchiveTime *int            `json:"archive_time"`
	IsFallback  bool            `json:"is_fallback"`
}

func (s *Service) AddSensor(ctx context.Context, dev *Device, dto NewSensorDTO) error {
	opts := NewSensorOpts{
		Code:        dto.Code,
		Brand:       dto.Brand,
		Description: dto.Description,
		ExternalID:  dto.ExternalID,
		Properties:  dto.Properties,
		ArchiveTime: dto.ArchiveTime,
		IsFallback:  dto.IsFallback,
	}
	if err := dev.AddSensor(opts); err != nil {
		return err
	}
	if err := s.store.Save(dev); err != nil {
		return err
	}
	return nil
}

func (s *Service) DeleteSensor(ctx context.Context, dev *Device, sensor *Sensor) error {
	if err := dev.DeleteSensorByID(sensor.ID); err != nil {
		return err
	}
	if err := s.store.Save(dev); err != nil {
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
	if err := s.store.Save(dev); err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteDevice(ctx context.Context, dev *Device) error {
	if err := s.store.Delete(dev); err != nil {
		return err
	}
	return nil
}

func (s *Service) ListSensors(ctx context.Context, p pagination.Request) (*pagination.Page[Sensor], error) {
	return s.store.ListSensors(p)
}

func (s *Service) GetSensor(ctx context.Context, id int64) (*Sensor, error) {
	return s.store.GetSensor(id)
}

func (s *Service) CreateSensorGroup(ctx context.Context, name, description string) (*SensorGroup, error) {
	group, err := NewSensorGroup(name, description)
	if err != nil {
		return nil, fmt.Errorf("create sensor group failed: %w", err)
	}
	if err := s.sensorGroupStore.Save(group); err != nil {
		return nil, fmt.Errorf("could not store sensor group: %w", err)
	}
	return group, nil
}

func (s *Service) ListSensorGroups(ctx context.Context, p pagination.Request) (*pagination.Page[SensorGroup], error) {
	return s.sensorGroupStore.List(p)
}

func (s *Service) GetSensorGroup(ctx context.Context, id int64) (*SensorGroup, error) {
	return s.sensorGroupStore.Get(id)
}

func (s *Service) AddSensorToSensorGroup(ctx context.Context, groupID, sensorID int64) error {
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
	return s.sensorGroupStore.Delete(group.ID)
}

type UpdateSensorGroupOpts struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func (s *Service) UpdateSensorGroup(ctx context.Context, group *SensorGroup, opts UpdateSensorGroupOpts) error {
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
