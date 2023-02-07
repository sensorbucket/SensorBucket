package service_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/device/service"
)

func TestHTTPListDeviceUsesRegularList(t *testing.T) {
	isCalled := false
	var argFilter service.DeviceFilter
	svc := &ServiceMock{
		ListDevicesFunc: func(ctx context.Context, filter service.DeviceFilter) ([]service.Device, error) {
			argFilter = filter
			isCalled = true
			return []service.Device{}, nil
		},
	}
	transport := service.NewHTTPTransport(svc)
	req := httptest.NewRequest("GET", "/devices", nil)
	rw := httptest.NewRecorder()

	transport.ServeHTTP(rw, req)

	assert.True(t, isCalled)
	assert.Empty(t, argFilter.Configuration)
}

func TestHTTPListDeviceUsesBoundingBox(t *testing.T) {
	isCalled := false
	var argFilter service.DeviceFilter
	var argBB service.BoundingBox
	svc := &ServiceMock{
		ListInBoundingBoxFunc: func(ctx context.Context, bb service.BoundingBox, filter service.DeviceFilter) ([]service.Device, error) {
			argFilter = filter
			argBB = bb
			isCalled = true
			return []service.Device{}, nil
		},
	}
	expectedBB := service.BoundingBox{
		North: 1,
		East:  2,
		South: 3,
		West:  4,
	}
	transport := service.NewHTTPTransport(svc)
	url := fmt.Sprintf(
		"/devices?north=%f&west=%f&south=%f&east=%f",
		expectedBB.North, expectedBB.West,
		expectedBB.South, expectedBB.East,
	)
	req := httptest.NewRequest("GET", url, nil)
	rw := httptest.NewRecorder()

	transport.ServeHTTP(rw, req)

	assert.True(t, isCalled)
	assert.Empty(t, argFilter.Configuration)
	assert.Equal(t, argBB, expectedBB)
}

func TestHTTPListDeviceUsesInRange(t *testing.T) {
	isCalled := false
	var argFilter service.DeviceFilter
	var argLR service.LocationRange
	svc := &ServiceMock{
		ListInRangeFunc: func(ctx context.Context, lr service.LocationRange, filter service.DeviceFilter) ([]service.Device, error) {
			argFilter = filter
			argLR = lr
			isCalled = true
			return []service.Device{}, nil
		},
	}
	expectedLR := service.LocationRange{
		Latitude:  1,
		Longitude: 2,
		Distance:  3,
	}
	transport := service.NewHTTPTransport(svc)
	url := fmt.Sprintf(
		"/devices?latitude=%f&longitude=%f&distance=%f",
		expectedLR.Latitude, expectedLR.Longitude, expectedLR.Distance,
	)
	req := httptest.NewRequest("GET", url, nil)
	rw := httptest.NewRecorder()

	transport.ServeHTTP(rw, req)

	assert.True(t, isCalled)
	assert.Empty(t, argFilter.Configuration)
	assert.Equal(t, argLR, expectedLR)
}

func TestHTTPListDeviceUsesInRangeOverBoundingBox(t *testing.T) {
	// This test tests whether the http transport prioritizes in range over bounding box
	// as mentioned in the spec.
	// The ServiceMock does not specify ListInBoundingBoxFunc so if it would be called by
	// the transport, it would  fail the test by default.
	// Also note that in the request parameters for both in range and bounding box are specified.
	isCalled := false
	var argFilter service.DeviceFilter
	var argLR service.LocationRange
	svc := &ServiceMock{
		ListInRangeFunc: func(ctx context.Context, lr service.LocationRange, filter service.DeviceFilter) ([]service.Device, error) {
			argFilter = filter
			argLR = lr
			isCalled = true
			return []service.Device{}, nil
		},
	}
	expectedLR := service.LocationRange{
		Latitude:  1,
		Longitude: 2,
		Distance:  3,
	}
	transport := service.NewHTTPTransport(svc)
	url := fmt.Sprintf(
		"/devices?latitude=%f&longitude=%f&distance=%f&north=1&west=1&east=1&south=1",
		expectedLR.Latitude, expectedLR.Longitude, expectedLR.Distance,
	)
	req := httptest.NewRequest("GET", url, nil)
	rw := httptest.NewRecorder()

	transport.ServeHTTP(rw, req)

	assert.True(t, isCalled)
	assert.Empty(t, argFilter.Configuration)
	assert.Equal(t, argLR, expectedLR)
}

func TestHTTPShouldListSensorTypes(t *testing.T) {
	types := []service.SensorType{
		{
			ID:          1,
			Description: "type_1",
		},
		{
			ID:          2,
			Description: "type_2",
		},
		{
			ID:          3,
			Description: "type_3",
		},
	}
	svc := &ServiceMock{
		ListSensorTypesFunc: func(ctx context.Context) ([]service.SensorType, error) {
			return types, nil
		},
	}
	transport := service.NewHTTPTransport(svc)

	// Act
	r := httptest.NewRequest("GET", "/sensortypes", nil)
	rw := httptest.NewRecorder()
	transport.ServeHTTP(rw, r)

	// assert
	response := rw.Result()
	require.Equal(t, http.StatusOK, response.StatusCode)
	var apiResponse web.APIResponse[[]service.SensorType]
	if err := json.NewDecoder(response.Body).Decode(&apiResponse); err != nil {
		require.NoError(t, err, "could not decode http response")
	}
	assert.Equal(t, types, apiResponse.Data)
}
func TestHTTPShouldListSensorGoals(t *testing.T) {
	goals := []service.SensorGoal{
		{
			ID:          1,
			Name:        "goal_1",
			Description: "goal_1",
		},
		{
			ID:          2,
			Name:        "goal_2",
			Description: "goal_2",
		},
		{
			ID:          3,
			Name:        "goal_3",
			Description: "goal_3",
		},
	}
	svc := &ServiceMock{
		ListSensorGoalsFunc: func(ctx context.Context) ([]service.SensorGoal, error) {
			return goals, nil
		},
	}
	transport := service.NewHTTPTransport(svc)

	// Act
	r := httptest.NewRequest("GET", "/sensorgoals", nil)
	rw := httptest.NewRecorder()
	transport.ServeHTTP(rw, r)

	// assert
	response := rw.Result()
	require.Equal(t, http.StatusOK, response.StatusCode)
	var apiResponse web.APIResponse[[]service.SensorGoal]
	if err := json.NewDecoder(response.Body).Decode(&apiResponse); err != nil {
		require.NoError(t, err, "could not decode http response")
	}
	assert.Equal(t, goals, apiResponse.Data)
}
