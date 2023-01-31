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
