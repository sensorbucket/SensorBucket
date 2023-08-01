package devices_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sensorbucket.nl/sensorbucket/services/core/devices"
)

func ptr[T any](a T) *T {
	return &a
}

func TestServiceDeviceUpdates(t *testing.T) {
	updateDTO := devices.UpdateDeviceOpts{
		Description:         ptr("description_b"),
		Latitude:            ptr(float64(30)),
		Longitude:           ptr(float64(40)),
		Altitude:            ptr(float64(50)),
		State:               ptr(devices.DeviceDisabled),
		LocationDescription: ptr("location_description_b"),
		Properties:          []byte(`{"meta":true}`),
	}
	originalDevice := devices.Device{
		Code:                "1234",
		Description:         "description_a",
		Organisation:        "organisation_a",
		Sensors:             []devices.Sensor{},
		Properties:          []byte("{}"),
		Latitude:            ptr(float64(10)),
		Longitude:           ptr(float64(20)),
		Altitude:            ptr(float64(30)),
		State:               devices.DeviceEnabled,
		LocationDescription: "location_description_a",
	}
	var newDevice devices.Device
	store := &DeviceStoreMock{SaveFunc: func(dev *devices.Device) error {
		newDevice = *dev
		return nil
	}}

	svc := devices.New(store, nil)

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
	newDTO := devices.NewDeviceOpts{
		Code:                "1234",
		Description:         "description_a",
		Organisation:        "organisation_a",
		Properties:          []byte("{}"),
		Latitude:            ptr(float64(10)),
		Longitude:           ptr(float64(20)),
		Altitude:            ptr(float64(30)),
		State:               devices.DeviceEnabled,
		LocationDescription: "location_description_a",
	}
	var storedDev *devices.Device
	store := &DeviceStoreMock{SaveFunc: func(dev *devices.Device) error {
		storedDev = dev
		return nil
	}}
	svc := devices.New(store, nil)

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
	dev := devices.Device{
		Code:                "1234",
		Description:         "description_a",
		Organisation:        "organisation_a",
		Sensors:             []devices.Sensor{},
		Properties:          []byte("{}"),
		Latitude:            ptr(float64(10)),
		Longitude:           ptr(float64(20)),
		LocationDescription: "location_description_a",
	}
	sensorDTO := devices.NewSensorDTO{
		Code:        "sensorcode",
		Brand:       "sensorbrand",
		Description: "sensordescription",
		ExternalID:  "sensorexternalid",
		Properties:  json.RawMessage("{}"),
		ArchiveTime: ptr(1000),
	}
	store := &DeviceStoreMock{
		SaveFunc: func(dev *devices.Device) error {
			return nil
		},
	}
	svc := devices.New(store, nil)

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

func TestServiceShouldAddSensorToSensorGroup(t *testing.T) {
	ctx := context.Background()
	var sensorGroupID int64 = 5
	var sensorID int64 = 10
	deviceStore := &DeviceStoreMock{
		GetSensorFunc: func(id int64) (*devices.Sensor, error) {
			if id != sensorID {
				return nil, devices.ErrSensorNotFound
			}
			return &devices.Sensor{
				ID:   sensorID,
				Code: "Sensor",
			}, nil
		},
	}
	sensorGroupStore := &SensorGroupStoreMock{
		GetFunc: func(id int64) (*devices.SensorGroup, error) {
			if id != sensorGroupID {
				return nil, devices.ErrSensorGroupNotFound
			}
			return &devices.SensorGroup{
				ID:      sensorGroupID,
				Name:    "testgroup",
				Sensors: []int64{},
			}, nil
		},
		SaveFunc: func(group *devices.SensorGroup) error {
			return nil
		},
	}
	svc := devices.New(deviceStore, sensorGroupStore)

	// Act
	err := svc.AddSensorToSensorGroup(ctx, sensorGroupID, sensorID)

	// Assert
	assert.NoError(t, err)
	if assert.Len(t, sensorGroupStore.SaveCalls(), 1, "Should save new group") {
		assert.Equal(t, []int64{sensorID}, sensorGroupStore.SaveCalls()[0].Group.Sensors, "Should have added sensor to group")
	}
}

func TestServiceShouldDeleteSensorFromSensorGroup(t *testing.T) {
	ctx := context.Background()
	var sensorGroupID int64 = 5
	var sensorID int64 = 10
	deviceStore := &DeviceStoreMock{
		GetSensorFunc: func(id int64) (*devices.Sensor, error) {
			if id != sensorID {
				return nil, devices.ErrSensorNotFound
			}
			return &devices.Sensor{
				ID:   sensorID,
				Code: "Sensor",
			}, nil
		},
	}
	sensorGroupStore := &SensorGroupStoreMock{
		GetFunc: func(id int64) (*devices.SensorGroup, error) {
			if id != sensorGroupID {
				return nil, devices.ErrSensorGroupNotFound
			}
			return &devices.SensorGroup{
				ID:      sensorGroupID,
				Name:    "testgroup",
				Sensors: []int64{sensorID},
			}, nil
		},
		SaveFunc: func(group *devices.SensorGroup) error {
			return nil
		},
	}
	svc := devices.New(deviceStore, sensorGroupStore)

	// Act
	err := svc.DeleteSensorFromSensorGroup(ctx, sensorGroupID, sensorID)

	// Assert
	assert.NoError(t, err)
	if assert.Len(t, sensorGroupStore.SaveCalls(), 1, "Should save new group") {
		assert.Equal(t, []int64{}, sensorGroupStore.SaveCalls()[0].Group.Sensors, "Should have removed sensor to group")
	}
}
