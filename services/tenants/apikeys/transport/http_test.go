package apikeystransport

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestValidateNoAuthorizationHeaderInRequest(t *testing.T) {
	// Arrange
	svc := apiKeyServiceMock{}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("GET", "/api-keys/validate", nil)

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"message":"Authorization header must be set"}`+"\n", rr.Body.String())
}

func TestValidateAuthorizationHeaderIncorrectFormat(t *testing.T) {
	// Arrange
	svc := apiKeyServiceMock{}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("GET", "/api-keys/validate", nil)

	// Act
	req.Header["Authorization"] = []string{"wrong format!!"}
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"message":"Authorization header must be set"}`+"\n", rr.Body.String())
}

func TestValidateErrorOccursWhileValidatingApiKey(t *testing.T) {
	// Arrange
	svc := apiKeyServiceMock{
		ValidateApiKeyFunc: func(base64IdAndKeyCombination string) (bool, error) {
			assert.Equal(t, "MjMxNDMyNDM6bXl2YWxpZGFwaWtleQ==", base64IdAndKeyCombination)
			return false, fmt.Errorf("database error!")
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("GET", "/api-keys/validate", nil)

	// Act
	req.Header["Authorization"] = []string{"Bearer MjMxNDMyNDM6bXl2YWxpZGFwaWtleQ=="}
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, `{"message":"Internal server error"}`+"\n", rr.Body.String())
	assert.Len(t, svc.ValidateApiKeyCalls(), 1)
}

func TestValidateApiKeyIsValid(t *testing.T) {
	// Arrange
	svc := apiKeyServiceMock{
		ValidateApiKeyFunc: func(base64IdAndKeyCombination string) (bool, error) {
			assert.Equal(t, "MjMxNDMyNDM6bXl2YWxpZGFwaWtleQ==", base64IdAndKeyCombination)
			return true, nil
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("GET", "/api-keys/validate", nil)

	// Act
	req.Header["Authorization"] = []string{"Bearer MjMxNDMyNDM6bXl2YWxpZGFwaWtleQ=="}
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"message":"API Key is valid","data":""}`+"\n", rr.Body.String())
	assert.Len(t, svc.ValidateApiKeyCalls(), 1)
}

func testTransport(svc apiKeyService) *HTTPTransport {
	transport := &HTTPTransport{
		apiKeySvc: svc,
		router:    chi.NewMux(),
	}
	transport.setupRoutes(transport.router)
	return transport
}
