package service_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	deviceservice "sensorbucket.nl/sensorbucket/services/device/service"
	"sensorbucket.nl/sensorbucket/services/measurements/service"
)

func ptr[T any](v T) *T {
	return &v
}

var (
	prefabSensor1       deviceservice.Sensor
	prefabDevice1       pipeline.Device
	prefabTimestamp     time.Time
	prefabMeasurement1  pipeline.Measurement
	prefabMessage       pipeline.Message
	expectedMeasurement service.Measurement
)

func resetPrefabs() {
	prefabSensor1 = deviceservice.Sensor{
		ID:            1,
		Code:          "abcd",
		Description:   "",
		Brand:         "",
		ArchiveTime:   100,
		Type:          1,
		Goal:          1,
		Configuration: json.RawMessage{},
		ExternalID:    "1",
	}
	prefabDevice1 = pipeline.Device{
		ID:                  1,
		Code:                "",
		Description:         "",
		Organisation:        "",
		Sensors:             []deviceservice.Sensor{prefabSensor1},
		Latitude:            ptr(float64(10)),
		Longitude:           ptr(float64(20)),
		LocationDescription: "",
		Configuration:       json.RawMessage([]byte(`{"hello":"world"}`)),
	}
	prefabTimestamp = time.Now()
	prefabMeasurement1 = pipeline.Measurement{
		Timestamp:              prefabTimestamp.UnixMilli(),
		SensorExternalID:       prefabSensor1.ExternalID,
		MeasurementValue:       5,
		MeasurementValueFactor: 0,
		MeasurementType:        "",
		MeasurementUnit:        "",
		MeasurementLatitude:    ptr(float64(30)),
		MeasurementLongitude:   ptr(float64(40)),
		MeasurementMetadata:    map[string]any{},
	}
	prefabMessage = pipeline.Message{
		ID:            uuid.NewString(),
		PipelineID:    uuid.NewString(),
		PipelineSteps: []string{},
		Timestamp:     prefabTimestamp.UnixMilli(),
		Device:        &prefabDevice1,
		Measurements: []pipeline.Measurement{
			prefabMeasurement1,
		},
	}
	expectedMeasurement = service.Measurement{
		UplinkMessageID:           prefabMessage.ID,
		OrganisationName:          "",
		OrganisationAddress:       "",
		OrganisationZipcode:       "",
		OrganisationCity:          "",
		OrganisationCoC:           "",
		OrganisationLocationCoC:   "",
		DeviceID:                  prefabMessage.Device.ID,
		DeviceCode:                prefabMessage.Device.Code,
		DeviceDescription:         prefabMessage.Device.Description,
		DeviceLatitude:            prefabMessage.Device.Latitude,
		DeviceLongitude:           prefabMessage.Device.Longitude,
		DeviceLocationDescription: prefabMessage.Device.LocationDescription,
		DeviceConfiguration:       prefabMessage.Device.Configuration,
		SensorID:                  prefabSensor1.ID,
		SensorCode:                prefabSensor1.Code,
		SensorTypeID:              prefabSensor1.Type,
		SensorTypeDescription:     "",
		SensorGoalID:              prefabSensor1.Goal,
		SensorGoalName:            "",
		SensorDescription:         prefabSensor1.Description,
		SensorExternalID:          prefabSensor1.ExternalID,
		SensorConfig:              prefabSensor1.Configuration,
		SensorBrand:               prefabSensor1.Brand,
		MeasurementType:           prefabMeasurement1.MeasurementType,
		MeasurementUnit:           prefabMeasurement1.MeasurementUnit,
		MeasurementTimestamp:      time.UnixMilli(prefabMeasurement1.Timestamp),
		MeasurementValue:          prefabMeasurement1.MeasurementValue,
		MeasurementValuePrefix:    "",
		MeasurementValueFactor:    prefabMeasurement1.MeasurementValueFactor,
		MeasurementLatitude:       prefabMeasurement1.MeasurementLatitude,
		MeasurementLongitude:      prefabMeasurement1.MeasurementLongitude,
		MeasurementMetadata:       prefabMeasurement1.MeasurementMetadata,
	}
}

func TestShouldConvertPipelineMessageToMeasurements(t *testing.T) {
	for _, tc := range []struct {
		Name          string
		Setup         func()
		ExpectedError error
	}{
		{
			Name:  "Default case",
			Setup: func() {},
		},
		{
			Name: "Should fallback to device location if no measurement location is set",
			Setup: func() {
				prefabMessage.Measurements[0].MeasurementLatitude = nil
				prefabMessage.Measurements[0].MeasurementLongitude = nil
				expectedMeasurement.MeasurementLatitude = prefabDevice1.Latitude
				expectedMeasurement.MeasurementLongitude = prefabDevice1.Longitude
			},
		},
		{
			Name: "Should throw if no device is set",
			Setup: func() {
				prefabMessage.Device = nil
			},
			ExpectedError: service.ErrMissingDeviceInMeasurement,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			resetPrefabs()
			var storeInsertCallCount int
			storeInsertArgs := []service.Measurement{}
			store := &StoreMock{
				InsertFunc: func(measurement service.Measurement) error {
					storeInsertCallCount++
					storeInsertArgs = append(storeInsertArgs, measurement)
					return nil
				},
			}
			svc := service.New(store)

			tc.Setup()
			err := svc.StorePipelineMessage(context.Background(), prefabMessage)
			if tc.ExpectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.ExpectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(prefabMessage.Measurements), storeInsertCallCount)
				assert.EqualValues(t, expectedMeasurement, storeInsertArgs[0])
			}
		})
	}
}
