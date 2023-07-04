package devicetransport_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"sensorbucket.nl/sensorbucket/services/core/devices"
	devicetransport "sensorbucket.nl/sensorbucket/services/core/devices/transport"
)

func TestHTTPListDeviceUsesRegularList(t *testing.T) {
	isCalled := false
	var argFilter devices.DeviceFilter
	svc := &ServiceMock{
		ListDevicesFunc: func(ctx context.Context, filter devices.DeviceFilter) ([]devices.Device, error) {
			argFilter = filter
			isCalled = true
			return []devices.Device{}, nil
		},
	}
	transport := devicetransport.NewHTTPTransport(svc)
	req := httptest.NewRequest("GET", "/devices", nil)
	rw := httptest.NewRecorder()

	transport.ServeHTTP(rw, req)

	assert.True(t, isCalled)
	assert.Empty(t, argFilter.Properties)
}
