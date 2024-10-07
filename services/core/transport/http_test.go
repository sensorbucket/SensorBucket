package coretransport_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/authtest"
	"sensorbucket.nl/sensorbucket/services/core/devices"
	"sensorbucket.nl/sensorbucket/services/core/measurements"
	coretransport "sensorbucket.nl/sensorbucket/services/core/transport"
)

func TestShouldCheckAuthentication(t *testing.T) {
	req, _ := http.NewRequest("GET", "/00000000-0000-0000-0000-000000000000", nil)
	res := httptest.NewRecorder()

	// Services can be nil since it shouldn't even reach them!
	transport := coretransport.New("", nil, nil, nil, nil)

	transport.ServeHTTP(res, req)

	assert.Equal(t, http.StatusUnauthorized, res.Result().StatusCode)
}

func TestMeasurementResponseShouldIncludeMeasurementID(t *testing.T) {
	measurement := measurements.Measurement{
		ID:                              100, // This field is not in the JSON, so we'll leave it as 0
		UplinkMessageID:                 "ff0bde71-e88e-406e-824e-2a0bb5346778",
		OrganisationID:                  3,
		OrganisationName:                "Provincie Zeeland",
		OrganisationAddress:             "ABCD",
		OrganisationZipcode:             "1234AB",
		OrganisationCity:                "STAD",
		OrganisationChamberOfCommerceID: "123456789",
		OrganisationHeadquarterID:       "123456789",
		OrganisationArchiveTime:         10,
		OrganisationState:               1,
		DeviceID:                        24,
		DeviceCode:                      "D49C10000000405C",
		DeviceDescription:               "Provincie Zeeland SensorBox for NO2, OX, PM, RV",
		DeviceLatitude:                  lo.ToPtr(float64(12.4)),
		DeviceLongitude:                 lo.ToPtr(float64(13.5)),
		DeviceAltitude:                  lo.ToPtr(float64(14.6)),
		DeviceLocationDescription:       "Some location",
		DeviceProperties:                json.RawMessage(`{"dev_eui":"123123123"}`),
		DeviceState:                     devices.DeviceState(1),
		SensorID:                        87,
		SensorCode:                      "PM",
		SensorDescription:               "Particulate Matter",
		SensorExternalID:                "sps30",
		SensorProperties:                json.RawMessage(`{}`),
		SensorBrand:                     "Sensirion",
		SensorIsFallback:                false,
		SensorArchiveTime:               nil,
		DatastreamID:                    uuid.MustParse("cbf909c4-6a56-40e1-b7ce-cfd6e7bff573"),
		DatastreamDescription:           "",
		DatastreamObservedProperty:      "pm_typical_size",
		DatastreamUnitOfMeasurement:     "nm",
		MeasurementTimestamp:            mustParseTime("2024-10-07T09:23:56Z"),
		MeasurementValue:                0.751,
		MeasurementLatitude:             lo.ToPtr(float64(0)),
		MeasurementLongitude:            lo.ToPtr(float64(0)),
		MeasurementAltitude:             nil,
		MeasurementProperties:           nil,
		MeasurementExpiration:           mustParseTime("2024-10-14T00:00:00Z"),
		CreatedAt:                       mustParseTime("2024-10-07T09:23:57Z"),
	}
	measurementService := &MeasurementServiceMock{
		QueryMeasurementsFunc: func(contextMoqParam context.Context, filter measurements.Filter, request pagination.Request) (*pagination.Page[measurements.Measurement], error) {
			return &pagination.Page[measurements.Measurement]{
				Data: []measurements.Measurement{
					measurement,
				},
			}, nil
		},
	}
	transport := coretransport.New("", authtest.JWKS(), nil, measurementService, nil)

	req, _ := http.NewRequest("GET", "/measurements", nil)
	authtest.AuthenticateRequest(req)
	res := httptest.NewRecorder()

	transport.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Result().StatusCode)
	var responseBody pagination.Page[measurements.Measurement]
	require.NoError(t, json.NewDecoder(res.Body).Decode(&responseBody))
	responseMeasurement := responseBody.Data[0]
	assert.EqualValues(t, measurement, responseMeasurement)
}

func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}
