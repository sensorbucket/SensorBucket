package tenantstransport

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"sensorbucket.nl/sensorbucket/services/tenants/tenants"
)

func TestNewApiKeyTenantDoesNotExist(t *testing.T) {
	// Arrange
	svc := serviceMock{
		GetTenantByIdFunc: func(id int64) (tenants.Tenant, error) {
			assert.Equal(t, int64(53424), id)
			return tenants.Tenant{}, tenants.ErrTenantIsNotValid
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("POST", "/api-keys/new", strings.NewReader(`{"organisation_id": 43324}`))

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Equal(t, `{"message":"Organisation does not exist or has been archived"}`+"\n", rr.Body.String())
}

func TestNewApiKeyInvalidParams(t *testing.T) {}

func TestNewApiKeyGeneratesNewApiKey(t *testing.T)             {}
func TestNewApiKeyErrorOccursWhileCreatingApiKey(t *testing.T) {}

func TestValidateNoAuthorizationHeaderInRequest(t *testing.T) {
	// Arrange
	svc := serviceMock{}
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
	svc := serviceMock{}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("GET", "/api-keys/validate", nil)

	// Act
	req.Header["Authorization"] = []string{"Bearer wrong format!!"}
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"message":"Authorization header must have format 'Bearer id:key'"}`+"\n", rr.Body.String())
}

func TestValidateIdDoesNotExist(t *testing.T) {
	// Arrange
	svc := serviceMock{
		GetHashedApiKeyByIdFunc: func(id int64) (tenants.ApiKey, error) {
			return tenants.ApiKey{}, tenants.ErrKeyNotFound
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("GET", "/api-keys/validate", nil)

	// Act
	req.Header["Authorization"] = []string{"Bearer 23143243:myvalidapikey"}
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Equal(t, "{}\n", rr.Body.String())
}

func TestValidatepiKeyIsNotCorrect(t *testing.T) {
	// Arrange
	svc := serviceMock{
		GetHashedApiKeyByIdFunc: func(id int64) (tenants.ApiKey, error) {
			assert.Equal(t, int64(23143243), id)
			return tenants.ApiKey{
				HashedValue: "incorrect hash value",
			}, nil
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("GET", "/api-keys/validate", nil)

	// Act
	req.Header["Authorization"] = []string{"Bearer 23143243:myvalidapikey"}
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Equal(t, "{}\n", rr.Body.String())
}

func TestValidateErrorOccursWhileRetrievingApiKeyById(t *testing.T) {
	// Arrange
	svc := serviceMock{
		GetHashedApiKeyByIdFunc: func(id int64) (tenants.ApiKey, error) {
			assert.Equal(t, int64(23143243), id)
			return tenants.ApiKey{}, fmt.Errorf("database error!")
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("GET", "/api-keys/validate", nil)

	// Act
	req.Header["Authorization"] = []string{"Bearer 23143243:myvalidapikey"}
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, `{"message":"Internal server error"}`+"\n", rr.Body.String())
}

func TestValidateApiKeyIsValid(t *testing.T) {
	// Arrange
	svc := serviceMock{
		GetHashedApiKeyByIdFunc: func(id int64) (tenants.ApiKey, error) {
			assert.Equal(t, int64(23143243), id)
			return tenants.ApiKey{
				HashedValue: "$2a$10$nLKvnGhciEtz8Iyp.iTeWObnwGzoJQ/iEpRF8vSTdbsXWhW2h1KXK",
			}, nil
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("GET", "/api-keys/validate", nil)

	// Act
	req.Header["Authorization"] = []string{"Bearer 23143243:yowMG6WfCErul8rpMX4Xu98PLpVhRois"}
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"message":"API Key is valid","data":""}`+"\n", rr.Body.String())
}

func testTransport(svc service) *HTTPTransport {
	transport := &HTTPTransport{
		svc:    svc,
		router: chi.NewMux(),
	}
	transport.setupRoutes(transport.router)
	return transport
}
