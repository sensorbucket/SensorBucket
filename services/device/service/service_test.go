package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"sensorbucket.nl/sensorbucket/services/device/service"
)

func ptr[T any](a T) *T {
	return &a
}

func TestServiceDeviceUpdates(t *testing.T) {
	updateDTO := service.UpdateDeviceOpts{
		Description:         ptr("description_b"),
		Latitude:            ptr(float64(30)),
		Longitude:           ptr(float64(40)),
		LocationDescription: ptr("location_description_b"),
		Configuration:       []byte(`{"meta":true}`),
	}
	originalDevice := service.Device{
		Code:                "1234",
		Description:         "description_a",
		Organisation:        "organisation_a",
		Sensors:             []service.Sensor{},
		Configuration:       []byte("{}"),
		Latitude:            ptr(float64(10)),
		Longitude:           ptr(float64(20)),
		LocationDescription: "location_description_a",
	}
	var newDevice service.Device
	store := &StoreMock{SaveFunc: func(dev *service.Device) error {
		newDevice = *dev
		return nil
	}}

	svc := service.New(store)

	err := svc.UpdateDevice(context.Background(), &originalDevice, updateDTO)
	assert.NoError(t, err)
	assert.EqualValues(t, newDevice.Description, *updateDTO.Description)
	assert.EqualValues(t, newDevice.Latitude, updateDTO.Latitude)
	assert.EqualValues(t, newDevice.Longitude, updateDTO.Longitude)
	assert.EqualValues(t, newDevice.LocationDescription, *updateDTO.LocationDescription)
	assert.EqualValues(t, newDevice.Configuration, updateDTO.Configuration)
}

func TestServiceCreateDevice(t *testing.T) {
	newDTO := service.NewDeviceOpts{
		Code:                "1234",
		Description:         "description_a",
		Organisation:        "organisation_a",
		Configuration:       []byte("{}"),
		Latitude:            10,
		Longitude:           20,
		LocationDescription: "location_description_a",
	}
	var storedDev *service.Device
	store := &StoreMock{SaveFunc: func(dev *service.Device) error {
		storedDev = dev
		return nil
	}}
	svc := service.New(store)

	_, err := svc.CreateDevice(context.Background(), newDTO)
	assert.NoError(t, err)
	assert.EqualValues(t, newDTO.Code, storedDev.Code)
	assert.EqualValues(t, newDTO.Organisation, storedDev.Organisation)
	assert.EqualValues(t, newDTO.Description, storedDev.Description)
	assert.EqualValues(t, newDTO.Latitude, storedDev.Latitude)
	assert.EqualValues(t, newDTO.Longitude, storedDev.Longitude)
	assert.EqualValues(t, newDTO.LocationDescription, storedDev.LocationDescription)
	assert.EqualValues(t, newDTO.Configuration, storedDev.Configuration)
	assert.Len(t, storedDev.Sensors, 0)
}
