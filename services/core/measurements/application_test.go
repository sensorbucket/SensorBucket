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
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	"sensorbucket.nl/sensorbucket/services/core/devices"
	"sensorbucket.nl/sensorbucket/services/core/measurements"
)

func ptr[T any](v T) *T {
	return &v
}

func TestShouldErrorIfNoDeviceOrNoSensor(t *testing.T) {
	device := &pipeline.Device{
		ID:                  1,
		Code:                "",
		Description:         "",
		Organisation:        "",
		Sensors:             []devices.Sensor{},
		Latitude:            ptr(float64(10)),
		Longitude:           ptr(float64(20)),
		Altitude:            ptr(float64(30)),
		LocationDescription: "",
		State:               devices.DeviceEnabled,
		Properties:          json.RawMessage([]byte(`{"hello":"world"}`)),
	}
	sensor := &devices.Sensor{
		ID:          1,
		Code:        "",
		Description: "",
		Brand:       "",
		ArchiveTime: nil,
		ExternalID:  "",
		Properties:  json.RawMessage("{}"),
	}
	testCases := []struct {
		desc   string
		device *pipeline.Device
		sensor *devices.Sensor
		err    error
	}{
		{
			desc:   "Both set, no error",
			device: device,
			sensor: sensor,
			err:    nil,
		},
		{
			desc:   "Device set, no sensor",
			device: device,
			sensor: nil,
			err:    measurements.ErrInvalidSensorID,
		},
		{
			desc:   "Device not set",
			device: nil,
			sensor: nil,
			err:    measurements.ErrMissingDeviceInMeasurement,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			msg := pipeline.NewMessage(uuid.NewString(), []string{})
			if tC.device != nil {
				msg.Device = tC.device
				if tC.sensor != nil {
					msg.Device.Sensors = append(msg.Device.Sensors, *tC.sensor)
				}
			}
			err := msg.NewMeasurement().SetValue(5, "test_obs", "1").SetSensor("").Add()
			require.NoError(t, err)
			store := &StoreMock{
				FindDatastreamFunc: func(sensorID int64, obs string) (*measurements.Datastream, error) {
					return &measurements.Datastream{}, nil
				},
				InsertFunc: func(measurement measurements.Measurement) error {
					return nil
				},
			}
			svc := measurements.New(store, 0)

			// Act
			err = svc.StorePipelineMessage(context.Background(), *msg)
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
	msg := pipeline.NewMessage(uuid.NewString(), []string{})
	msg.Device = &pipeline.Device{
		ID:           1,
		Code:         "",
		Description:  "",
		Organisation: "",
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
		Latitude:            ptr(float64(10)),
		Longitude:           ptr(float64(20)),
		Altitude:            ptr(float64(30)),
		LocationDescription: "",
		State:               devices.DeviceEnabled,
		Properties:          json.RawMessage([]byte(`{"hello":"world"}`)),
	}
	err := msg.NewMeasurement().SetValue(5, "test_obs", "1").SetSensor("").Add()
	require.NoError(t, err)
	ds := measurements.Datastream{
		ID:                uuid.New(),
		Description:       "",
		SensorID:          msg.Device.Sensors[0].ID,
		ObservedProperty:  msg.Measurements[0].ObservedProperty,
		UnitOfMeasurement: msg.Measurements[0].UnitOfMeasurement,
	}
	store := &StoreMock{
		FindDatastreamFunc: func(sensorID int64, obs string) (*measurements.Datastream, error) {
			return &ds, nil
		},
		InsertFunc: func(measurement measurements.Measurement) error {
			return nil
		},
	}
	svc := measurements.New(store, 0)

	// Act
	err = svc.StorePipelineMessage(context.Background(), *msg)
	require.NoError(t, err)

	// Assert
	require.Len(t, store.calls.Insert, 1, "SQL Insert should've been called")
	measurement := store.calls.Insert[0].Measurement
	assert.Equal(t, msg.ID, measurement.UplinkMessageID)
	//assert.Equal(t, OrganisationName, measurement.OrganisationName)
	//assert.Equal(t, OrganisationAddress, measurement.OrganisationAddress)
	//assert.Equal(t, OrganisationZipcode, measurement.OrganisationZipcode)
	//assert.Equal(t, OrganisationCity, measurement.OrganisationCity)
	//assert.Equal(t, OrganisationChamberOfCommerceID, measurement.OrganisationChamberOfCommerceID)
	//assert.Equal(t, OrganisationHeadquarterID, measurement.OrganisationHeadquarterID)
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
			msg := pipeline.NewMessage(uuid.NewString(), []string{})
			msg.Device = &pipeline.Device{
				ID:           1,
				Code:         "",
				Description:  "",
				Organisation: "",
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
				FindDatastreamFunc: func(sensorID int64, obs string) (*measurements.Datastream, error) {
					return &ds, nil
				},
				InsertFunc: func(measurement measurements.Measurement) error {
					return nil
				},
			}
			svc := measurements.New(store, 0)

			// Act
			require.NoError(t,
				svc.StorePipelineMessage(context.Background(), *msg),
			)

			// Assert
			require.Len(t, store.calls.Insert, 1, "SQL Insert should've been called")
			measurement := store.calls.Insert[0].Measurement
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
		msg := pipeline.NewMessage(uuid.NewString(), []string{})
		msg.ReceivedAt = now.UnixMilli()
		msg.Device = &pipeline.Device{
			ID:           1,
			Code:         "",
			Description:  "",
			Organisation: "",
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
		require.NoError(t, err)
		ds := measurements.Datastream{
			ID:                uuid.New(),
			Description:       "",
			SensorID:          msg.Device.Sensors[0].ID,
			ObservedProperty:  msg.Measurements[0].ObservedProperty,
			UnitOfMeasurement: msg.Measurements[0].UnitOfMeasurement,
		}
		store := &StoreMock{
			FindDatastreamFunc: func(sensorID int64, obs string) (*measurements.Datastream, error) {
				return &ds, nil
			},
			InsertFunc: func(measurement measurements.Measurement) error {
				return nil
			},
		}
		svc := measurements.New(store, sysArchiveTime)

		// Act
		err = svc.StorePipelineMessage(context.Background(), *msg)
		require.NoError(t, err)

		// Assert
		require.Len(t, store.calls.Insert, 1, "SQL Insert should've been called")
		measurement := store.calls.Insert[0].Measurement
		// Check if the difference in seconds is 0, otherwise there might be a subsecond difference
		// due to parsing
		assert.Equal(t,
			float64(0),
			math.Abs(float64(tC.expectedArchiveTime.Unix()-measurement.MeasurementExpiration.Unix())),
			"",
		)
	}
}
