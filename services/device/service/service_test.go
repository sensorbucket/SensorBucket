package service_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		Altitude:            ptr(float64(50)),
		State:               ptr(service.DeviceDisabled),
		LocationDescription: ptr("location_description_b"),
		Properties:          []byte(`{"meta":true}`),
	}
	originalDevice := service.Device{
		Code:                "1234",
		Description:         "description_a",
		Organisation:        "organisation_a",
		Sensors:             []service.Sensor{},
		Properties:          []byte("{}"),
		Latitude:            ptr(float64(10)),
		Longitude:           ptr(float64(20)),
		Altitude:            ptr(float64(30)),
		State:               service.DeviceEnabled,
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
	assert.EqualValues(t, newDevice.Altitude, updateDTO.Altitude)
	assert.EqualValues(t, newDevice.State, *updateDTO.State)
	assert.EqualValues(t, newDevice.LocationDescription, *updateDTO.LocationDescription)
	assert.EqualValues(t, newDevice.Properties, updateDTO.Properties)
}

func TestServiceCreateDevice(t *testing.T) {
	newDTO := service.NewDeviceOpts{
		Code:                "1234",
		Description:         "description_a",
		Organisation:        "organisation_a",
		Properties:          []byte("{}"),
		Latitude:            ptr(float64(10)),
		Longitude:           ptr(float64(20)),
		Altitude:            ptr(float64(30)),
		State:               service.DeviceEnabled,
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
	assert.EqualValues(t, newDTO.Altitude, storedDev.Altitude)
	assert.EqualValues(t, newDTO.State, storedDev.State)
	assert.EqualValues(t, newDTO.LocationDescription, storedDev.LocationDescription)
	assert.EqualValues(t, newDTO.Properties, storedDev.Properties)
	assert.Len(t, storedDev.Sensors, 0)
}

func TestServiceShouldAddSensor(t *testing.T) {
	ctx := context.Background()
	dev := service.Device{
		Code:                "1234",
		Description:         "description_a",
		Organisation:        "organisation_a",
		Sensors:             []service.Sensor{},
		Properties:          []byte("{}"),
		Latitude:            ptr(float64(10)),
		Longitude:           ptr(float64(20)),
		LocationDescription: "location_description_a",
	}
	sensorDTO := service.NewSensorDTO{
		Code:        "sensorcode",
		Brand:       "sensorbrand",
		Description: "sensordescription",
		ExternalID:  "sensorexternalid",
		Properties:  json.RawMessage("{}"),
		ArchiveTime: 1000,
	}
	store := &StoreMock{
		SaveFunc: func(dev *service.Device) error {
			return nil
		},
	}
	svc := service.New(store)

	// Act
	err := svc.AddSensor(ctx, &dev, sensorDTO)
	require.NoError(t, err)

	// Assert
	require.Len(t, store.calls.Save, 1)
	calledDev := store.calls.Save[0].Dev
	require.Len(t, calledDev.Sensors, 1)
	dto := calledDev.Sensors[0]
	assert.Equal(t, sensorDTO.Code, dto.Code)
	assert.Equal(t, sensorDTO.Description, dto.Description)
	assert.Equal(t, sensorDTO.Brand, dto.Brand)
	assert.Equal(t, sensorDTO.Description, dto.Description)
	assert.Equal(t, sensorDTO.ExternalID, dto.ExternalID)
	assert.Equal(t, sensorDTO.ArchiveTime, dto.ArchiveTime)
}
