package measurements_test

import (
	"context"
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sensorbucket.nl/sensorbucket/pkg/authtest"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	"sensorbucket.nl/sensorbucket/services/core/devices"
	"sensorbucket.nl/sensorbucket/services/core/measurements"
)

func ptr[T any](v T) *T {
	return &v
}

func newPipelineMessage(plID string, steps []string) pipeline.Message {
	return pipeline.Message{
		TracingID:     uuid.NewString(),
		ReceivedAt:    time.Now().UnixMilli(),
		Timestamp:     time.Now().UnixMilli(),
		Payload:       nil,
		PipelineID:    plID,
		PipelineSteps: steps,
		StepIndex:     0,
		Measurements:  []pipeline.Measurement{},
		Metadata:      make(map[string]any),
	}
}

func TestShouldErrorIfNoDeviceOrNoSensor(t *testing.T) {
	createDevice := func() *pipeline.Device {
		return &pipeline.Device{
			ID:                  1,
			Code:                "",
			Description:         "",
			TenantID:            10,
			Sensors:             []devices.Sensor{},
			Latitude:            ptr(float64(10)),
			Longitude:           ptr(float64(20)),
			Altitude:            ptr(float64(30)),
			LocationDescription: "",
			State:               devices.DeviceEnabled,
			Properties:          json.RawMessage([]byte(`{"hello":"world"}`)),
		}
	}
	testCases := []struct {
		desc                        string
		device                      *pipeline.Device
		sensor                      []devices.Sensor
		sensorExternalID            string
		observationProperty         string
		expectedObservationProperty string
		err                         error
	}{
		{
			desc:   "message has both a device set and a sensor match for externalID",
			device: createDevice(),
			sensor: []devices.Sensor{
				{
					ID:          1,
					Code:        "",
					Description: "",
					Brand:       "",
					ArchiveTime: nil,
					ExternalID:  "matching_eid",
					Properties:  json.RawMessage("{}"),
				},
			},
			sensorExternalID: "matching_eid",
			err:              nil,
		},
		{
			desc:             "message has a device set but no sensor match and no fallback",
			device:           createDevice(),
			sensor:           []devices.Sensor{},
			sensorExternalID: "not_existing_eid",
			err:              measurements.ErrInvalidSensorID,
		},
		{
			desc:             "message has no device set",
			device:           nil,
			sensor:           []devices.Sensor{},
			sensorExternalID: "",
			err:              measurements.ErrMissingDeviceInMeasurement,
		},
		{
			desc:   "message has device set, no sensor match but has a fallback sensor",
			device: createDevice(),
			sensor: []devices.Sensor{
				{
					ID:          1,
					Code:        "",
					Description: "",
					Brand:       "",
					ArchiveTime: nil,
					ExternalID:  "matching_eid",
					Properties:  json.RawMessage("{}"),
					IsFallback:  true,
				},
			},
			sensorExternalID:    "eid",
			observationProperty: "obs",
			// Because a fallback is used, the observation property should be prefixed with the eid
			expectedObservationProperty: "eid_obs",
			err:                         nil,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			msg := newPipelineMessage(uuid.NewString(), []string{})
			if tC.device != nil {
				msg.Device = tC.device
				if tC.sensor != nil {
					msg.Device.Sensors = tC.sensor
				}
			}
			if tC.observationProperty == "" {
				tC.observationProperty = "default_obs"
				tC.expectedObservationProperty = "default_obs"
			}

			err := msg.NewMeasurement().SetValue(5, tC.observationProperty, "1").SetSensor(tC.sensorExternalID).Add()
			msg.AccessToken = authtest.CreateToken()
			require.NoError(t, err)

			store := &StoreMock{
				FindOrCreateDatastreamFunc: func(ctx context.Context, tenantID, sensorID int64, observedProperty, UnitOfMeasurement string) (*measurements.Datastream, error) {
					return &measurements.Datastream{}, nil
				},
				StoreMeasurementFunc: func(contextMoqParam context.Context, measurement measurements.Measurement) error { return nil },
			}
			svc := measurements.New(store, 0, 1, authtest.JWKS())

			// Act
			err = svc.ProcessPipelineMessage(msg)
			if tC.err != nil {
				assert.Error(t, tC.err, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// This tests whether the related models are properly copied into the measurements
// It does not test any logic
func TestShouldCopyOverDefaultFields(t *testing.T) {
	msg := newPipelineMessage(uuid.NewString(), []string{})
	msg.Device = &pipeline.Device{
		ID:          1,
		Code:        "123",
		Description: "",
		TenantID:    10,
		Sensors: []devices.Sensor{
			{
				ID:          1,
				Code:        "123",
				Description: "",
				Brand:       "",
				ArchiveTime: nil,
				ExternalID:  "",
				Properties:  json.RawMessage("{}"),
			},
		},
		Latitude:            ptr(float64(10)),
		Longitude:           ptr(float64(20)),
		Altitude:            ptr(float64(30)),
		LocationDescription: "",
		State:               devices.DeviceEnabled,
		Properties:          json.RawMessage([]byte(`{"hello":"world"}`)),
	}
	err := msg.NewMeasurement().SetValue(5, "test_obs", "1").SetSensor("").Add()
	msg.AccessToken = authtest.CreateToken()
	require.NoError(t, err)
	ds := measurements.Datastream{
		ID:                uuid.New(),
		Description:       "",
		SensorID:          msg.Device.Sensors[0].ID,
		ObservedProperty:  msg.Measurements[0].ObservedProperty,
		UnitOfMeasurement: msg.Measurements[0].UnitOfMeasurement,
	}
	store := &StoreMock{
		FindOrCreateDatastreamFunc: func(ctx context.Context, tenantID, sensorID int64, observedProperty, UnitOfMeasurement string) (*measurements.Datastream, error) {
			return &ds, nil
		},
		StoreMeasurementFunc: func(contextMoqParam context.Context, measurement measurements.Measurement) error { return nil },
	}
	svc := measurements.New(store, 0, 1, authtest.JWKS())

	// Act
	err = svc.ProcessPipelineMessage(msg)
	require.NoError(t, err)
	// assert.NoError(t, svc.CommitBatch(true))

	// Assert
	require.Len(t, store.calls.StoreMeasurement, 1, "StoreMeasurements should've been called")
	measurement := store.calls.StoreMeasurement[0].Measurement
	assert.Equal(t, msg.TracingID, measurement.UplinkMessageID)
	// assert.Equal(t, OrganisationName, measurement.OrganisationName)
	// assert.Equal(t, OrganisationAddress, measurement.OrganisationAddress)
	// assert.Equal(t, OrganisationZipcode, measurement.OrganisationZipcode)
	// assert.Equal(t, OrganisationCity, measurement.OrganisationCity)
	// assert.Equal(t, OrganisationChamberOfCommerceID, measurement.OrganisationChamberOfCommerceID)
	// assert.Equal(t, OrganisationHeadquarterID, measurement.OrganisationHeadquarterID)
	assert.Equal(t, msg.Device.ID, measurement.DeviceID)
	assert.Equal(t, msg.Device.Code, measurement.DeviceCode)
	assert.Equal(t, msg.Device.Description, measurement.DeviceDescription)
	assert.Equal(t, msg.Device.Latitude, measurement.DeviceLatitude)
	assert.Equal(t, msg.Device.Longitude, measurement.DeviceLongitude)
	assert.Equal(t, msg.Device.Altitude, measurement.DeviceAltitude)
	assert.Equal(t, msg.Device.LocationDescription, measurement.DeviceLocationDescription)
	assert.Equal(t, msg.Device.State, measurement.DeviceState)
	assert.Equal(t, msg.Device.Properties, measurement.DeviceProperties)
	assert.Equal(t, msg.Device.Sensors[0].ID, measurement.SensorID)
	assert.Equal(t, msg.Device.Sensors[0].Code, measurement.SensorCode)
	assert.Equal(t, msg.Device.Sensors[0].Description, measurement.SensorDescription)
	assert.Equal(t, msg.Device.Sensors[0].ExternalID, measurement.SensorExternalID)
	assert.Equal(t, msg.Device.Sensors[0].Properties, measurement.SensorProperties)
	assert.Equal(t, msg.Device.Sensors[0].Brand, measurement.SensorBrand)
	assert.Equal(t, msg.Device.Sensors[0].ArchiveTime, measurement.SensorArchiveTime)
	assert.Equal(t, msg.Device.Sensors[0].IsFallback, measurement.SensorIsFallback)
	assert.Equal(t, ds.ID, measurement.DatastreamID)
	assert.Equal(t, ds.Description, measurement.DatastreamDescription)
	assert.Equal(t, ds.ObservedProperty, measurement.DatastreamObservedProperty)
	assert.Equal(t, ds.UnitOfMeasurement, measurement.DatastreamUnitOfMeasurement)
}

func TestShouldChooseMeasurementLocationOverDeviceLocation(t *testing.T) {
	testCases := []struct {
		desc                 string
		DeviceLongitude      *float64
		DeviceLatitude       *float64
		DeviceAltitude       *float64
		MeasurementLongitude *float64
		MeasurementLatitude  *float64
		MeasurementAltitude  *float64
		ExpectedLongitude    *float64
		ExpectedLatitude     *float64
		ExpectedAltitude     *float64
	}{
		{
			desc:                 "Device not set, Measurement set",
			MeasurementLatitude:  ptr(float64(10)),
			MeasurementLongitude: ptr(float64(20)),
			MeasurementAltitude:  ptr(float64(30)),
			ExpectedLatitude:     ptr(float64(10)),
			ExpectedLongitude:    ptr(float64(20)),
			ExpectedAltitude:     ptr(float64(30)),
		},
		{
			desc:              "Device set, Measurement not set",
			DeviceLatitude:    ptr(float64(10)),
			DeviceLongitude:   ptr(float64(20)),
			DeviceAltitude:    ptr(float64(30)),
			ExpectedLatitude:  ptr(float64(10)),
			ExpectedLongitude: ptr(float64(20)),
			ExpectedAltitude:  ptr(float64(30)),
		},
		{
			desc:                 "Device set, Measurement set",
			DeviceLatitude:       ptr(float64(90)),
			DeviceLongitude:      ptr(float64(90)),
			DeviceAltitude:       ptr(float64(90)),
			MeasurementLatitude:  ptr(float64(10)),
			MeasurementLongitude: ptr(float64(20)),
			MeasurementAltitude:  ptr(float64(30)),
			ExpectedLatitude:     ptr(float64(10)),
			ExpectedLongitude:    ptr(float64(20)),
			ExpectedAltitude:     ptr(float64(30)),
		},
		{
			desc: "None set",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			msg := newPipelineMessage(uuid.NewString(), []string{})
			msg.Device = &pipeline.Device{
				ID:          1,
				Code:        "",
				Description: "",
				TenantID:    10,
				Sensors: []devices.Sensor{
					{
						ID:          1,
						Code:        "",
						Description: "",
						Brand:       "",
						ArchiveTime: nil,
						ExternalID:  "",
						Properties:  json.RawMessage("{}"),
					},
				},
				Latitude:            tC.DeviceLatitude,
				Longitude:           tC.DeviceLongitude,
				Altitude:            tC.DeviceAltitude,
				LocationDescription: "",
				State:               devices.DeviceEnabled,
				Properties:          json.RawMessage([]byte(`{"hello":"world"}`)),
			}
			builder := msg.NewMeasurement().SetValue(5, "test_obs", "1").SetSensor("")
			msg.AccessToken = authtest.CreateToken()
			if tC.MeasurementLatitude != nil && tC.MeasurementLongitude != nil && tC.MeasurementAltitude != nil {
				builder = builder.SetLocation(
					*tC.MeasurementLatitude,
					*tC.MeasurementLongitude,
					*tC.MeasurementAltitude,
				)
			}
			require.NoError(t, builder.Add())
			ds := measurements.Datastream{
				ID:                uuid.New(),
				Description:       "",
				SensorID:          msg.Device.Sensors[0].ID,
				ObservedProperty:  msg.Measurements[0].ObservedProperty,
				UnitOfMeasurement: msg.Measurements[0].UnitOfMeasurement,
			}
			store := &StoreMock{
				FindOrCreateDatastreamFunc: func(ctx context.Context, tenantID, sensorID int64, observedProperty, UnitOfMeasurement string) (*measurements.Datastream, error) {
					return &ds, nil
				},
				StoreMeasurementFunc: func(contextMoqParam context.Context, measurement measurements.Measurement) error { return nil },
			}
			svc := measurements.New(store, 0, 1, authtest.JWKS())

			// Act
			require.NoError(t,
				svc.ProcessPipelineMessage(msg),
			)
			// assert.NoError(t, svc.CommitBatch(true))

			// Assert
			require.Len(t, store.calls.StoreMeasurement, 1, "StoreMeasurements should've been called")
			measurement := store.calls.StoreMeasurement[0].Measurement
			assert.Equal(t, tC.ExpectedLatitude, measurement.MeasurementLatitude)
			assert.Equal(t, tC.ExpectedLongitude, measurement.MeasurementLongitude)
			assert.Equal(t, tC.ExpectedAltitude, measurement.MeasurementAltitude)
		})
	}
}

func TestShouldSetExpirationDate(t *testing.T) {
	now := time.Now()
	sysArchiveTime := 14
	testCases := []struct {
		desc                    string
		organisationArchiveTime *int
		sensorArchiveTime       *int
		expectedArchiveTime     time.Time
	}{
		{
			desc:                    "No organisationArchiveTime and no sensorArchiveTime, should use system ArchiveTime",
			organisationArchiveTime: nil,
			sensorArchiveTime:       nil,
			expectedArchiveTime:     now.Add(time.Duration(sysArchiveTime) * 24 * time.Hour),
		},
	}
	for _, tC := range testCases {
		msg := newPipelineMessage(uuid.NewString(), []string{})
		msg.ReceivedAt = now.UnixMilli()
		msg.Device = &pipeline.Device{
			ID:          1,
			Code:        "",
			Description: "",
			TenantID:    10,
			// TODO: We kind of need organisationArchiveTime here, but how do we get it?
			// Perhaps split the device endpoint in two:
			//  - one ep for just the device info
			//  - one ep for workers with all relevant data (device, sensors, org)
			Sensors: []devices.Sensor{
				{
					ID:          1,
					Code:        "",
					Description: "",
					Brand:       "",
					ArchiveTime: tC.sensorArchiveTime,
					ExternalID:  "",
					Properties:  json.RawMessage("{}"),
				},
			},
			Latitude:            ptr(float64(10)),
			Longitude:           ptr(float64(20)),
			Altitude:            ptr(float64(30)),
			LocationDescription: "",
			State:               devices.DeviceEnabled,
			Properties:          json.RawMessage([]byte(`{"hello":"world"}`)),
		}
		err := msg.NewMeasurement().SetValue(5, "test_obs", "1").SetSensor("").Add()
		msg.AccessToken = authtest.CreateToken()
		require.NoError(t, err)
		ds := measurements.Datastream{
			ID:                uuid.New(),
			Description:       "",
			SensorID:          msg.Device.Sensors[0].ID,
			ObservedProperty:  msg.Measurements[0].ObservedProperty,
			UnitOfMeasurement: msg.Measurements[0].UnitOfMeasurement,
		}
		store := &StoreMock{
			FindOrCreateDatastreamFunc: func(ctx context.Context, tenantID, sensorID int64, observedProperty, UnitOfMeasurement string) (*measurements.Datastream, error) {
				return &ds, nil
			},
			StoreMeasurementFunc: func(contextMoqParam context.Context, measurement measurements.Measurement) error { return nil },
		}
		svc := measurements.New(store, sysArchiveTime, 1, authtest.JWKS())

		// Act
		err = svc.ProcessPipelineMessage(msg)
		require.NoError(t, err)
		// assert.NoError(t, svc.CommitBatch(true))

		// Assert
		require.Len(t, store.calls.StoreMeasurement, 1, "StoreMeasurements should've been called")
		measurement := store.calls.StoreMeasurement[0].Measurement
		// Check if the difference in seconds is 0, otherwise there might be a subsecond difference
		// due to parsing
		assert.Equal(t,
			float64(0),
			math.Abs(float64(tC.expectedArchiveTime.Unix()-measurement.MeasurementExpiration.Unix())),
			"",
		)
	}
}
