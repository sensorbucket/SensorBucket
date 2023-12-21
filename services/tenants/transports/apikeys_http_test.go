package tenantstransports

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"sensorbucket.nl/sensorbucket/services/tenants/apikeys"
)

func TestNewApiKeyInvalidJsonBody(t *testing.T) {
	// Arrange
	svc := apiKeyServiceMock{}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("POST", "/api-keys/new", strings.NewReader(`blablabla`))

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"message":"Invalid JSON body"}`+"\n", rr.Body.String())
	assert.Len(t, svc.GenerateNewApiKeyCalls(), 0)
}

func TestNewApiKeyTenantIsNotFound(t *testing.T) {
	// Arrange
	svc := apiKeyServiceMock{
		GenerateNewApiKeyFunc: func(name string, tenantId int64, expiry *time.Time) (string, error) {
			assert.Equal(t, int64(905), tenantId)
			assert.Nil(t, expiry)
			return "", apikeys.ErrTenantIsNotValid
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("POST", "/api-keys/new", strings.NewReader(`{"name": "whatever", "organisation_id": 905}`))

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Equal(t, `{"message":"Organisation does not exist or has been archived"}`+"\n", rr.Body.String())
	assert.Len(t, svc.GenerateNewApiKeyCalls(), 1)
}

func TestNewApiKeyIsCreatedWithExpirationDate(t *testing.T) {
	// Arrange
	exp := time.Now().UTC().Add(time.Hour * 24 * 5)
	svc := apiKeyServiceMock{
		GenerateNewApiKeyFunc: func(name string, tenantId int64, expiry *time.Time) (string, error) {
			assert.Equal(t, int64(905), tenantId)
			assert.NotNil(t, expiry)
			assert.Equal(t, exp, *expiry)
			return "newapikey", nil
		},
	}
	transport := testTransport(&svc)
	fmt.Println(exp.String())
	req, _ := http.NewRequest("POST", "/api-keys/new", strings.NewReader(fmt.Sprintf(`{"name": "whatever", "organisation_id": 905, "expiration_date": "%s"}`, exp.Format("2006-01-02T15:04:05.999999999Z"))))

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, `{"api_key":"newapikey"}`+"\n", rr.Body.String())
	assert.Len(t, svc.GenerateNewApiKeyCalls(), 1)
}

func TestNewApiKeyIsCreatedWithoutExpirationDate(t *testing.T) {
	// Arrange
	svc := apiKeyServiceMock{
		GenerateNewApiKeyFunc: func(name string, tenantId int64, expiry *time.Time) (string, error) {
			assert.Equal(t, int64(905), tenantId)
			assert.Nil(t, expiry)
			return "newapikey", nil
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("POST", "/api-keys/new", strings.NewReader(`{"name": "whatever", "organisation_id": 905}`))

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, `{"api_key":"newapikey"}`+"\n", rr.Body.String())
	assert.Len(t, svc.GenerateNewApiKeyCalls(), 1)
}

func TestRevokeApiKeyApiKeyInvalidEncodingErrorOccurs(t *testing.T) {
	// Arrange
	svc := apiKeyServiceMock{
		RevokeApiKeyFunc: func(base64IdAndKeyCombination string) error {
			assert.Equal(t, "blablabla", base64IdAndKeyCombination)
			return apikeys.ErrInvalidEncoding
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("DELETE", "/api-keys/revoke", strings.NewReader(`{"api_key": "blablabla"}`))

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"message":"Invalid input"}`+"\n", rr.Body.String())
	assert.Len(t, svc.RevokeApiKeyCalls(), 1)
}

func TestRevokeApiKeyInvalidJsonBody(t *testing.T) {
	// Arrange
	svc := apiKeyServiceMock{}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("DELETE", "/api-keys/revoke", strings.NewReader(`blablabla`))

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"message":"Invalid JSON body"}`+"\n", rr.Body.String())
	assert.Len(t, svc.RevokeApiKeyCalls(), 0)
}

func TestRevokeApiKeyEmptyApiKey(t *testing.T) {
	// Arrange
	svc := apiKeyServiceMock{}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("DELETE", "/api-keys/revoke", strings.NewReader(`{"api_key": ""}`))

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"message":"API key cannot be empty"}`+"\n", rr.Body.String())
	assert.Len(t, svc.RevokeApiKeyCalls(), 0)
}

func TestRevokeApiKeyRevokesApiKey(t *testing.T) {
	// Arrange
	svc := apiKeyServiceMock{
		RevokeApiKeyFunc: func(base64IdAndKeyCombination string) error {
			assert.Equal(t, "blablablabla", base64IdAndKeyCombination)
			return nil
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("DELETE", "/api-keys/revoke", strings.NewReader(`{"api_key": "blablablabla"}`))

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"message":"API key has been revoked","data":{"api_key":"blablablabla"}}`+"\n", rr.Body.String())
	assert.Len(t, svc.RevokeApiKeyCalls(), 1)
}

func TestRevokeApiKeyRevokeFails(t *testing.T) {
	// Arrange
	svc := apiKeyServiceMock{
		RevokeApiKeyFunc: func(base64IdAndKeyCombination string) error {
			assert.Equal(t, "blablablabla", base64IdAndKeyCombination)
			return fmt.Errorf("weird error")
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("DELETE", "/api-keys/revoke", strings.NewReader(`{"api_key": "blablablabla"}`))

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, `{"message":"Internal server error"}`+"\n", rr.Body.String())
	assert.Len(t, svc.RevokeApiKeyCalls(), 1)
}

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

func TestValidateApiKeyInvalidEncodingErrorOccurs(t *testing.T) {
	// Arrange
	svc := apiKeyServiceMock{
		ValidateApiKeyFunc: func(base64IdAndKeyCombination string) (bool, error) {
			assert.Equal(t, "blablabla", base64IdAndKeyCombination)
			return false, apikeys.ErrInvalidEncoding
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("GET", "/api-keys/validate", nil)

	// Act
	req.Header["Authorization"] = []string{"Bearer blablabla"}
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"message":"Invalid input"}`+"\n", rr.Body.String())
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
	assert.Equal(t, `{"message":"API Key is valid"}`+"\n", rr.Body.String())
	assert.Len(t, svc.ValidateApiKeyCalls(), 1)
}

func testTransport(svc apiKeyService) *APIKeysHTTPTransport {
	transport := &APIKeysHTTPTransport{
		apiKeySvc: svc,
		router:    chi.NewMux(),
	}
	transport.setupRoutes(transport.router)
	return transport
}

func asBase64(val string) string {
	return base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(val))
}
