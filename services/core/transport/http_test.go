package coretransport_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

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
