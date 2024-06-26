package service_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/authtest"
	"sensorbucket.nl/sensorbucket/services/core/processing"
	"sensorbucket.nl/sensorbucket/services/httpimporter/service"
)

// TODO: Assert that httimporter authenticates request

func TestReportInvalidUUID(t *testing.T) {
	ch := make(chan processing.IngressDTO, 1)
	svc := service.New(ch, authtest.JWKS())

	req := httptest.NewRequest("POST", "/THIS_IS_INVALID_UUID", nil)
	req.Header.Set("authorization", "bearer "+authtest.CreateToken())
	rw := httptest.NewRecorder()
	svc.ServeHTTP(rw, req)

	res := rw.Result()
	var apiResponse web.APIError
	if err := json.NewDecoder(res.Body).Decode(&apiResponse); err != nil {
		assert.NoError(t, err, "JSON Decode API (error)response")
	}
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestShouldPublishIngressDTO(t *testing.T) {
	requestData := []byte(`{"hello":"world"}`)
	publ := make(chan processing.IngressDTO, 1)
	svc := service.New(publ, authtest.JWKS())
	plID := uuid.New()

	// Act
	req := httptest.NewRequest("POST", "/"+plID.String(), bytes.NewBuffer(requestData))
	req.Header.Set("authorization", "bearer "+authtest.CreateToken())
	rw := httptest.NewRecorder()
	svc.ServeHTTP(rw, req)

	// Assert
	body, err := io.ReadAll(rw.Body)
	assert.NoErrorf(t, err, "io.ReadAll on reading response body")
	assert.Equal(t, http.StatusAccepted, rw.Result().StatusCode)
	require.Len(t, publ, 1)
	dto := <-publ
	assert.Contains(t, string(body), dto.TracingID.String())
	assert.Equal(t, plID, dto.PipelineID)
}
