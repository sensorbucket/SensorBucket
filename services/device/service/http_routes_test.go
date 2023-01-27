package service_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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
