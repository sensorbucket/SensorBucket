package service_test

import (
	"context"
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
