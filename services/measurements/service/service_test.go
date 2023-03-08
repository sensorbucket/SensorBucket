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
	prefabDatastream    service.Datastream
	expectedMeasurement service.Measurement
)

func resetPrefabs() {
	prefabSensor1 = deviceservice.Sensor{
		ID:          1,
		Code:        "abcd",
		Description: "",
		Brand:       "",
		ArchiveTime: 100,
		Properties:  json.RawMessage{},
		ExternalID:  "1",
	}
	prefabDatastream = service.Datastream{
		ID:                uuid.New(),
		Description:       "",
		SensorID:          1,
		UnitOfMeasurement: "1/cm3",
		ObservedProperty:  "mc_pm25",
	}
	prefabDevice1 = pipeline.Device{
		ID:                  1,
		Code:                "",
		Description:         "",
		Organisation:        "",
		Sensors:             []deviceservice.Sensor{prefabSensor1},
		Latitude:            ptr(float64(10)),
		Longitude:           ptr(float64(20)),
		Altitude:            ptr(float64(30)),
		LocationDescription: "",
		State:               deviceservice.DeviceEnabled,
		Properties:          json.RawMessage([]byte(`{"hello":"world"}`)),
	}
	prefabTimestamp = time.Now()
	prefabMeasurement1 = pipeline.Measurement{
		Timestamp:         prefabTimestamp.UnixMilli(),
		SensorExternalID:  prefabSensor1.ExternalID,
		Value:             5,
		ObservedProperty:  "",
		UnitOfMeasurement: "",
		Latitude:          ptr(float64(30)),
		Longitude:         ptr(float64(40)),
		Altitude:          ptr(float64(50)),
		Properties:        map[string]any{},
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
		UplinkMessageID:                 prefabMessage.ID,
		OrganisationName:                "",
		OrganisationAddress:             "",
		OrganisationZipcode:             "",
		OrganisationCity:                "",
		OrganisationChamberOfCommerceID: "",
		OrganisationHeadquarterID:       "",
		//OrganisationArchiveTime:         123,
		//OrganisationState:               1,
		DeviceID:                  prefabMessage.Device.ID,
		DeviceCode:                prefabMessage.Device.Code,
		DeviceDescription:         prefabMessage.Device.Description,
		DeviceLatitude:            prefabMessage.Device.Latitude,
		DeviceLongitude:           prefabMessage.Device.Longitude,
		DeviceAltitude:            prefabMessage.Device.Altitude,
		DeviceLocationDescription: prefabMessage.Device.LocationDescription,
		DeviceState:               prefabMessage.Device.State,
		DeviceProperties:          prefabMessage.Device.Properties,

		SensorID:          prefabSensor1.ID,
		SensorCode:        prefabSensor1.Code,
		SensorDescription: prefabSensor1.Description,
		SensorExternalID:  prefabSensor1.ExternalID,
		SensorProperties:  prefabSensor1.Properties,
		SensorBrand:       prefabSensor1.Brand,
		SensorArchiveTime: prefabSensor1.ArchiveTime,

		DatastreamID:                prefabDatastream.ID,
		DatastreamDescription:       prefabDatastream.Description,
		DatastreamObservedProperty:  prefabDatastream.ObservedProperty,
		DatastreamUnitOfMeasurement: prefabDatastream.UnitOfMeasurement,

		MeasurementTimestamp:  time.UnixMilli(prefabMeasurement1.Timestamp),
		MeasurementValue:      prefabMeasurement1.Value,
		MeasurementLatitude:   prefabMeasurement1.Latitude,
		MeasurementLongitude:  prefabMeasurement1.Longitude,
		MeasurementAltitude:   prefabMeasurement1.Altitude,
		MeasurementProperties: prefabMeasurement1.Properties,
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
				prefabMessage.Measurements[0].Latitude = nil
				prefabMessage.Measurements[0].Longitude = nil
				prefabMessage.Measurements[0].Altitude = nil
				expectedMeasurement.MeasurementLatitude = prefabDevice1.Latitude
				expectedMeasurement.MeasurementLongitude = prefabDevice1.Longitude
				expectedMeasurement.MeasurementAltitude = prefabDevice1.Altitude
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
				FindDatastreamFunc: func(sensorID int64, obs string) (*service.Datastream, error) {
					return &prefabDatastream, nil
				},
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
