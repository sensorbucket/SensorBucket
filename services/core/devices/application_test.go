package devices_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sensorbucket.nl/sensorbucket/pkg/authtest"
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
		TenantID:            authtest.DefaultTenantID,
		Sensors:             []devices.Sensor{},
		Properties:          []byte("{}"),
		Latitude:            ptr(float64(10)),
		Longitude:           ptr(float64(20)),
		Altitude:            ptr(float64(30)),
		State:               devices.DeviceEnabled,
		LocationDescription: "location_description_a",
	}
	var newDevice devices.Device
	store := &DeviceStoreMock{SaveFunc: func(ctx context.Context, dev *devices.Device) error {
		newDevice = *dev
		return nil
	}}

	svc := devices.New(store, nil, nil)

	err := svc.UpdateDevice(authtest.GodContext(), &originalDevice, updateDTO)
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
		Properties:          []byte("{}"),
		Latitude:            ptr(float64(10)),
		Longitude:           ptr(float64(20)),
		Altitude:            ptr(float64(30)),
		State:               devices.DeviceEnabled,
		LocationDescription: "location_description_a",
	}
	var storedDev *devices.Device
	store := &DeviceStoreMock{SaveFunc: func(ctx context.Context, dev *devices.Device) error {
		storedDev = dev
		return nil
	}}
	svc := devices.New(store, nil, nil)

	_, err := svc.CreateDevice(authtest.GodContext(), newDTO)
	assert.NoError(t, err)
	assert.EqualValues(t, newDTO.Code, storedDev.Code)
	assert.EqualValues(t, authtest.DefaultTenantID, storedDev.TenantID)
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
	dev := devices.Device{
		Code:                "1234",
		Description:         "description_a",
		TenantID:            authtest.DefaultTenantID,
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
		SaveFunc: func(ctx context.Context, dev *devices.Device) error {
			return nil
		},
	}
	svc := devices.New(store, nil, nil)

	// Act
	err := svc.AddSensor(authtest.GodContext(), &dev, sensorDTO)
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
	var sensorGroupID int64 = 5
	var sensorID int64 = 10
	deviceStore := &DeviceStoreMock{
		GetSensorFunc: func(ctx context.Context, id int64) (*devices.Sensor, error) {
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
		GetFunc: func(id int64, tenantID int64) (*devices.SensorGroup, error) {
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
	svc := devices.New(deviceStore, sensorGroupStore, nil)

	// Act
	err := svc.AddSensorToSensorGroup(authtest.GodContext(), sensorGroupID, sensorID)

	// Assert
	assert.NoError(t, err)
	if assert.Len(t, sensorGroupStore.SaveCalls(), 1, "Should save new group") {
		assert.Equal(t, []int64{sensorID}, sensorGroupStore.SaveCalls()[0].Group.Sensors, "Should have added sensor to group")
	}
}

func TestServiceShouldDeleteSensorFromSensorGroup(t *testing.T) {
	var sensorGroupID int64 = 5
	var sensorID int64 = 10
	deviceStore := &DeviceStoreMock{
		GetSensorFunc: func(ctx context.Context, id int64) (*devices.Sensor, error) {
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
		GetFunc: func(id int64, tenantID int64) (*devices.SensorGroup, error) {
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
	svc := devices.New(deviceStore, sensorGroupStore, nil)

	// Act
	err := svc.DeleteSensorFromSensorGroup(authtest.GodContext(), sensorGroupID, sensorID)

	// Assert
	assert.NoError(t, err)
	if assert.Len(t, sensorGroupStore.SaveCalls(), 1, "Should save new group") {
		assert.Equal(t, []int64{}, sensorGroupStore.SaveCalls()[0].Group.Sensors, "Should have removed sensor to group")
	}
}

func TestServiceShouldDeleteSensorGroup(t *testing.T) {
	sensorGroup := &devices.SensorGroup{
		ID: 5,
	}
	deviceStore := &DeviceStoreMock{}
	sensorGroupStore := &SensorGroupStoreMock{
		DeleteFunc: func(id int64) error {
			if id != sensorGroup.ID {
				return devices.ErrSensorGroupNotFound
			}
			return nil
		},
		GetFunc: func(id int64, tenantID int64) (*devices.SensorGroup, error) {
			if id != sensorGroup.ID {
				return nil, devices.ErrSensorGroupNotFound
			}
			return sensorGroup, nil
		},
		SaveFunc: func(group *devices.SensorGroup) error {
			return nil
		},
	}
	svc := devices.New(deviceStore, sensorGroupStore, nil)

	// Act
	err := svc.DeleteSensorGroup(authtest.GodContext(), sensorGroup)

	// Assert
	assert.NoError(t, err)
	if assert.Len(t, sensorGroupStore.DeleteCalls(), 1, "Should delete group") {
		assert.Equal(t, sensorGroup.ID, sensorGroupStore.DeleteCalls()[0].ID, "Should have removed the correct group")
	}
}

func TestServiceShouldUpdateSensorGroup(t *testing.T) {
	updatedName := "updated_sensorgroup"
	sensorGroup := &devices.SensorGroup{
		ID:   5,
		Name: "sensorgroup",
	}
	deviceStore := &DeviceStoreMock{}
	sensorGroupStore := &SensorGroupStoreMock{
		GetFunc: func(id int64, tenantID int64) (*devices.SensorGroup, error) {
			if id != sensorGroup.ID {
				return nil, devices.ErrSensorGroupNotFound
			}
			return sensorGroup, nil
		},
		SaveFunc: func(group *devices.SensorGroup) error {
			return nil
		},
	}
	svc := devices.New(deviceStore, sensorGroupStore, nil)
	dto := devices.UpdateSensorGroupOpts{
		Name: &updatedName,
	}
	invalidName := "a"
	dtoInvalid := devices.UpdateSensorGroupOpts{
		Name: &invalidName,
	}

	// Act
	err := svc.UpdateSensorGroup(authtest.GodContext(), sensorGroup, dto)
	assert.Error(t, svc.UpdateSensorGroup(authtest.GodContext(), sensorGroup, dtoInvalid))

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, updatedName, sensorGroup.Name)
	assert.NotZero(t, len(sensorGroupStore.SaveCalls()))
}
