package service_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/services/device/service"
)

func TestHTTPListDeviceUsesRegularList(t *testing.T) {
	isCalled := false
	var argFilter service.DeviceFilter
	svc := &ServiceMock{
		ListDevicesFunc: func(ctx context.Context, filter service.DeviceFilter, r pagination.Request) (*pagination.Page[service.Device], error) {
			argFilter = filter
			isCalled = true
			return &pagination.Page[service.Device]{Data: []service.Device{}}, nil
		},
	}
	transport := service.NewHTTPTransport(svc, "http://testurl")
	req := httptest.NewRequest("GET", "/devices", nil)
	rw := httptest.NewRecorder()

	transport.ServeHTTP(rw, req)

	assert.True(t, isCalled)
	assert.Empty(t, argFilter.Properties)
}
