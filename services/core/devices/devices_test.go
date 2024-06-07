package devices_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sensorbucket.nl/sensorbucket/pkg/authtest"
	"sensorbucket.nl/sensorbucket/services/core/devices"
)

func TestAddSensorToDevice(t *testing.T) {
	setupDevice := func() *devices.Device {
		device, err := devices.NewDevice(authtest.DefaultTenantID, devices.NewDeviceOpts{
			Code:        "testdevice",
			Description: "",
			Properties:  json.RawMessage("{}"),
			State:       devices.DeviceEnabled,
		})
		require.NoError(t, err)
		err = device.AddSensor(devices.NewSensorOpts{
			Code:        "existing",
			Description: "",
			ExternalID:  "existing",
			Properties:  json.RawMessage("{}"),
			IsFallback:  true,
		})
		require.NoError(t, err)
		return device
	}

	testCases := []struct {
		desc        string
		sensor      devices.NewSensorOpts
		expectedErr error
	}{
		{
			desc: "Add single non-conflicting sensor",
			sensor: devices.NewSensorOpts{
				Code:        "new_sensor",
				Brand:       "",
				Description: "",
				ExternalID:  "",
				Properties:  json.RawMessage("{}"),
				IsFallback:  false,
			},
			expectedErr: nil,
		},
		{
			desc: "Duplicate sensor code",
			sensor: devices.NewSensorOpts{
				Code:        "existing",
				Brand:       "",
				Description: "",
				ExternalID:  "",
				Properties:  json.RawMessage("{}"),
				IsFallback:  false,
			},
			expectedErr: devices.ErrDuplicateSensorCode,
		},
		{
			desc: "Duplicate sensor external ID",
			sensor: devices.NewSensorOpts{
				Code:        "new_sensor",
				Brand:       "",
				Description: "",
				ExternalID:  "existing",
				Properties:  json.RawMessage("{}"),
				IsFallback:  false,
			},
			expectedErr: devices.ErrDuplicateSensorExternalID,
		},
		{
			desc: "Two default sensors, one device",
			sensor: devices.NewSensorOpts{
				Code:        "new_sensor",
				Brand:       "",
				Description: "",
				ExternalID:  "new",
				Properties:  json.RawMessage("{}"),
				IsFallback:  true,
			},
			expectedErr: devices.ErrDuplicateFallbackSensor,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			device := setupDevice()
			err := device.AddSensor(tC.sensor)
			if tC.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tC.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeviceShouldProvideSensorByEIDOrFallback(t *testing.T) {
	dev, err := devices.NewDevice(authtest.DefaultTenantID, devices.NewDeviceOpts{
		Code: "test_device",
	})
	require.NoError(t, err)

	// Fetch non existing sensor
	sensor, err := dev.GetSensorByExternalIDOrFallback("sensor_eid")
	assert.Nil(t, sensor)
	assert.ErrorIs(t, err, devices.ErrSensorNotFound)

	// Add fallback
	opt := devices.NewSensorOpts{
		Code:       "test_fallback_sensor",
		IsFallback: true,
	}
	err = dev.AddSensor(opt)
	assert.NoError(t, err)

	// Fetch fallback
	sensor, err = dev.GetSensorByExternalIDOrFallback("sensor_eid")
	assert.NotNil(t, sensor)
	assert.NoError(t, err)
	assert.Equal(t, opt.Code, sensor.Code)

	// Add match sensor
	opt = devices.NewSensorOpts{
		Code:       "test_sensor",
		ExternalID: "sensor_eid",
	}
	err = dev.AddSensor(opt)
	assert.NoError(t, err)

	// Fetch fallback
	sensor, err = dev.GetSensorByExternalIDOrFallback("sensor_eid")
	assert.NotNil(t, sensor)
	assert.NoError(t, err)
	assert.Equal(t, opt.Code, sensor.Code)
}
