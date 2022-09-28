package service

import (
	"context"
	"encoding/json"
)

type Store interface {
	List(DeviceFilter) ([]Device, error)
	ListLocations() ([]Location, error)
	Find(id int) (*Device, error)
	Save(dev *Device) error
	Delete(dev *Device) error
}

type Service struct {
	store Store
}

func New(store Store) *Service {
	return &Service{
		store: store,
	}
}

type DeviceFilter struct {
	Configuration json.RawMessage `json:"configuration"`
}

func (s *Service) ListDevices(ctx context.Context, filter DeviceFilter) ([]Device, error) {
	devices, err := s.store.List(filter)
	return devices, err
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

func (s *Service) GetDevice(ctx context.Context, id int) (*Device, error) {
	dev, err := s.store.Find(id)
	if err != nil {
		return nil, err
	}
	if dev == nil {
		return nil, ErrDeviceNotFound
	}
	return dev, nil
}

func (s *Service) AddSensor(ctx context.Context, dev *Device, dto NewSensorOpts) error {
	if err := dev.AddSensor(dto); err != nil {
		return err
	}
	if err := s.store.Save(dev); err != nil {
		return err
	}
	return nil
}

func (s *Service) DeleteSensor(ctx context.Context, dev *Device, sensor *Sensor) error {
	if err := dev.DeleteSensor(sensor); err != nil {
		return err
	}
	if err := s.store.Save(dev); err != nil {
		return err
	}
	return nil
}

type UpdateDeviceOpts struct {
	Description   *string         `json:"description"`
	Configuration json.RawMessage `json:"configuration"`
}

func (s *Service) UpdateDevice(ctx context.Context, dev *Device, opt UpdateDeviceOpts) error {
	if opt.Description != nil {
		dev.Description = *opt.Description
	}
	if opt.Configuration != nil {
		dev.Configuration = opt.Configuration
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

func (s *Service) ListLocations(ctx context.Context) ([]Location, error) {
	locations, err := s.store.ListLocations()
	if err != nil {
		return nil, err
	}
	return locations, nil
}
