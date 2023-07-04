package devices

//go:generate moq -pkg service_test -out mock_test.go . Store Service

import (
	"context"
	"encoding/json"

	"sensorbucket.nl/sensorbucket/internal/pagination"
)

var _ Service = (*ServiceImpl)(nil)

type Store interface {
	List(DeviceFilter, pagination.Request) (*pagination.Page[Device], error)
	ListInBoundingBox(DeviceFilter, pagination.Request) (*pagination.Page[Device], error)
	ListInRange(DeviceFilter, pagination.Request) (*pagination.Page[Device], error)
	ListSensors(pagination.Request) (*pagination.Page[Sensor], error)
	Find(id int64) (*Device, error)
	Save(dev *Device) error
	Delete(dev *Device) error
}
type Service interface {
	CreateDevice(ctx context.Context, dto NewDeviceOpts) (*Device, error)
	ListDevices(ctx context.Context, filter DeviceFilter, r pagination.Request) (*pagination.Page[Device], error)
	GetDevice(ctx context.Context, id int64) (*Device, error)
	UpdateDevice(ctx context.Context, dev *Device, opt UpdateDeviceOpts) error
	DeleteDevice(ctx context.Context, dev *Device) error

	AddSensor(ctx context.Context, dev *Device, dto NewSensorDTO) error
	ListSensors(ctx context.Context, r pagination.Request) (*pagination.Page[Sensor], error)
	DeleteSensor(ctx context.Context, dev *Device, sensor *Sensor) error
}
type ServiceImpl struct {
	store Store
}

func New(store Store) *ServiceImpl {
	return &ServiceImpl{
		store: store,
	}
}

type DeviceFilter struct {
	BoundingBoxFilter
	RangeFilter
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

func (s *ServiceImpl) ListDevices(ctx context.Context, filter DeviceFilter, p pagination.Request) (*pagination.Page[Device], error) {
	if filter.HasBoundingBox() {
		return s.store.ListInBoundingBox(filter, p)
	}
	if filter.HasRange() {
		return s.store.ListInRange(filter, p)
	}
	return s.store.List(filter, p)
}

func (s *ServiceImpl) CreateDevice(ctx context.Context, dto NewDeviceOpts) (*Device, error) {
	dev, err := NewDevice(dto)
	if err != nil {
		return nil, err
	}
	if err := s.store.Save(dev); err != nil {
		return nil, err
	}
	return dev, nil
}

func (s *ServiceImpl) GetDevice(ctx context.Context, id int64) (*Device, error) {
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
}

func (s *ServiceImpl) AddSensor(ctx context.Context, dev *Device, dto NewSensorDTO) error {
	opts := NewSensorOpts{
		Code:        dto.Code,
		Brand:       dto.Brand,
		Description: dto.Description,
		ExternalID:  dto.ExternalID,
		Properties:  dto.Properties,
		ArchiveTime: dto.ArchiveTime,
	}
	if err := dev.AddSensor(opts); err != nil {
		return err
	}
	if err := s.store.Save(dev); err != nil {
		return err
	}
	return nil
}

func (s *ServiceImpl) DeleteSensor(ctx context.Context, dev *Device, sensor *Sensor) error {
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

func (s *ServiceImpl) UpdateDevice(ctx context.Context, dev *Device, opt UpdateDeviceOpts) error {
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

func (s *ServiceImpl) DeleteDevice(ctx context.Context, dev *Device) error {
	if err := s.store.Delete(dev); err != nil {
		return err
	}
	return nil
}

func (s *ServiceImpl) ListSensors(ctx context.Context, p pagination.Request) (*pagination.Page[Sensor], error) {
	return s.store.ListSensors(p)
}
