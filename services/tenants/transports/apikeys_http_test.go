package tenantstransports_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/services/tenants/apikeys"
	tenantstransports "sensorbucket.nl/sensorbucket/services/tenants/transports"
)

func TestNewApiKeyInvalidJsonBody(t *testing.T) {
	// Arrange
	svc := ApiKeyServiceMock{}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("POST", "/api-keys", strings.NewReader(`blablabla`))

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"message":"Invalid content type","code":"INVALID_CONTENT_TYPE"}`+"\n", rr.Body.String())
	assert.Len(t, svc.GenerateNewApiKeyCalls(), 0)
}

func TestNewApiKeyNoName(t *testing.T) {
	// Arrange
	svc := ApiKeyServiceMock{}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("POST", "/api-keys", strings.NewReader(`{"name": "", "tenant_id": 905}`))
	req.Header.Add("content-type", "application/json")

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"message":"name cannot be empty"}`+"\n", rr.Body.String())
	assert.Len(t, svc.GenerateNewApiKeyCalls(), 0)
}

func TestNewApiKeyNoOrganisationID(t *testing.T) {
	// Arrange
	svc := ApiKeyServiceMock{}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("POST", "/api-keys", strings.NewReader(`{"name": "wasdasdas", "tenant_id": 0}`))
	req.Header.Add("content-type", "application/json")

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"message":"tenant_id must be higher than 0"}`+"\n", rr.Body.String())
	assert.Len(t, svc.GenerateNewApiKeyCalls(), 0)
}

func TestNewApiKeyPermissionsNotGiven(t *testing.T) {
	// Arrange
	svc := ApiKeyServiceMock{}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("POST", "/api-keys",
		strings.NewReader(fmt.Sprintf(`{"name": "wasdasdas", "tenant_id": 12, "expiration_date": "%s"}`, time.Now().Add(time.Hour*24).Format(time.RFC3339))))
	req.Header.Add("content-type", "application/json")

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"message":"at least one permission is required"}`+"\n", rr.Body.String())
	assert.Len(t, svc.GenerateNewApiKeyCalls(), 0)
}

func TestNewApiKeyPermissionsEmptyList(t *testing.T) {
	// Arrange
	svc := ApiKeyServiceMock{}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("POST", "/api-keys",
		strings.NewReader(fmt.Sprintf(`{"name": "wasdasdas", "tenant_id": 12, "permissions":[], "expiration_date": "%s"}`, time.Now().Add(time.Hour*24).Format(time.RFC3339))))
	req.Header.Add("content-type", "application/json")

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"message":"at least one permission is required"}`+"\n", rr.Body.String())
	assert.Len(t, svc.GenerateNewApiKeyCalls(), 0)
}

func TestNewApiKeyExpirationDateNotInTheFuture(t *testing.T) {
	// Arrange
	svc := &ApiKeyServiceMock{}
	transport := testTransport(svc)
	req, _ := http.NewRequest("POST", "/api-keys",
		strings.NewReader(fmt.Sprintf(`{"name": "wasdasdas", "tenant_id": 12, "permissions":["WRITE_DEVICES"], "expiration_date": "%s"}`, time.Now().Add(-time.Hour*24).Format(time.RFC3339))))
	req.Header.Add("content-type", "application/json")

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"message":"expiration_date must be set in the future"}`+"\n", rr.Body.String())
	assert.Len(t, svc.GenerateNewApiKeyCalls(), 0)
}

func TestNewApiKeyTenantIsNotFound(t *testing.T) {
	// Arrange
	svc := ApiKeyServiceMock{
		GenerateNewApiKeyFunc: func(ctx context.Context, _ string, tenantId int64, permissions auth.Permissions, expiry *time.Time) (string, error) {
			assert.Equal(t, int64(905), tenantId)
			assert.Equal(t, auth.Permissions{auth.Permission("WRITE_DEVICES")}, permissions)
			assert.Nil(t, expiry)
			return "", apikeys.ErrTenantIsNotValid
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("POST", "/api-keys", strings.NewReader(`{"name": "whatever", "permissions":["WRITE_DEVICES"], "tenant_id": 905}`))
	req.Header.Add("content-type", "application/json")

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Equal(t, `{"message":"Organisation does not exist or has been archived"}`+"\n", rr.Body.String())
	assert.Len(t, svc.GenerateNewApiKeyCalls(), 1)
}

func TestNewApiKeyErrorOccurs(t *testing.T) {
	// Arrange
	svc := ApiKeyServiceMock{
		GenerateNewApiKeyFunc: func(ctx context.Context, _ string, tenantId int64, permissions auth.Permissions, expiry *time.Time) (string, error) {
			assert.Equal(t, int64(905), tenantId)
			assert.Equal(t, auth.Permissions{auth.Permission("WRITE_DEVICES")}, permissions)
			assert.Nil(t, expiry)
			return "", fmt.Errorf("weird error!")
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("POST", "/api-keys", strings.NewReader(`{"name": "whatever", "permissions":["WRITE_DEVICES"], "tenant_id": 905}`))
	req.Header.Add("content-type", "application/json")

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, `{"message":"Internal server error"}`+"\n", rr.Body.String())
	assert.Len(t, svc.GenerateNewApiKeyCalls(), 1)
}

func TestNewApiKeyIsCreatedWithExpirationDate(t *testing.T) {
	// Arrange
	exp := time.Now().UTC().Add(time.Hour * 24 * 5)
	svc := ApiKeyServiceMock{
		GenerateNewApiKeyFunc: func(ctx context.Context, _ string, tenantId int64, permissions auth.Permissions, expiry *time.Time) (string, error) {
			assert.Equal(t, int64(905), tenantId)
			assert.Equal(t, auth.Permissions{auth.Permission("WRITE_DEVICES")}, permissions)
			assert.NotNil(t, expiry)
			assert.Equal(t, exp, *expiry)
			return "newapikey", nil
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("POST", "/api-keys", strings.NewReader(fmt.Sprintf(`{"name": "whatever", "permissions":["WRITE_DEVICES"], "tenant_id": 905, "expiration_date": "%s"}`, exp.Format("2006-01-02T15:04:05.999999999Z"))))
	req.Header.Add("content-type", "application/json")

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
	svc := ApiKeyServiceMock{
		GenerateNewApiKeyFunc: func(ctx context.Context, _ string, tenantId int64, permissions auth.Permissions, expiry *time.Time) (string, error) {
			assert.Equal(t, int64(905), tenantId)
			assert.Equal(t, auth.Permissions{auth.Permission("WRITE_DEVICES")}, permissions)
			assert.Nil(t, expiry)
			return "newapikey", nil
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("POST", "/api-keys", strings.NewReader(`{"name": "whatever", "permissions":["WRITE_DEVICES"], "tenant_id": 905}`))
	req.Header.Add("content-type", "application/json")

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, `{"api_key":"newapikey"}`+"\n", rr.Body.String())
	assert.Len(t, svc.GenerateNewApiKeyCalls(), 1)
}

func TestRevokeApiKeyInvalidApiKeyId(t *testing.T) {
	// Arrange
	svc := ApiKeyServiceMock{}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("DELETE", "/api-keys/blablalb", nil)

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"message":"api_key_id must be a valid int"}`+"\n", rr.Body.String())
	assert.Len(t, svc.RevokeApiKeyCalls(), 0)
}

func TestRevokeApiKeyRevokesApiKey(t *testing.T) {
	// Arrange
	svc := ApiKeyServiceMock{
		RevokeApiKeyFunc: func(ctx context.Context, id int64) error {
			assert.Equal(t, int64(123), id)
			return nil
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("DELETE", "/api-keys/123", nil)

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"message":"API key has been revoked"}`+"\n", rr.Body.String())
	assert.Len(t, svc.RevokeApiKeyCalls(), 1)
}

func TestRevokeApiKeyRevokeFails(t *testing.T) {
	// Arrange
	svc := ApiKeyServiceMock{
		RevokeApiKeyFunc: func(ctx context.Context, id int64) error {
			assert.Equal(t, int64(12343), id)
			return fmt.Errorf("weird error")
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("DELETE", "/api-keys/12343", nil)

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, `{"message":"Internal server error"}`+"\n", rr.Body.String())
	assert.Len(t, svc.RevokeApiKeyCalls(), 1)
}

func TestRevokeApiKeyKeyDoesNotExist(t *testing.T) {
	// Arrange
	svc := ApiKeyServiceMock{
		RevokeApiKeyFunc: func(ctx context.Context, id int64) error {
			assert.Equal(t, int64(12343), id)
			return apikeys.ErrKeyNotFound
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("DELETE", "/api-keys/12343", nil)

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Equal(t, `{"message":"Key does not exist"}`+"\n", rr.Body.String())
	assert.Len(t, svc.RevokeApiKeyCalls(), 1)
}

func TestAuthenticateNoAuthorizationHeaderInRequest(t *testing.T) {
	// Arrange
	svc := ApiKeyServiceMock{}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("GET", "/api-keys/authenticate", nil)

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"message":"Authorization header must be set"}`+"\n", rr.Body.String())
}

func TestAuthenticateAuthorizationHeaderIncorrectFormat(t *testing.T) {
	// Arrange
	svc := ApiKeyServiceMock{}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("GET", "/api-keys/authenticate", nil)

	// Act
	req.Header["Authorization"] = []string{"wrong format!!"}
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"message":"Authorization header must be set"}`+"\n", rr.Body.String())
}

func TestAuthenticateErrorOccursWhileValidatingApiKey(t *testing.T) {
	// Arrange
	svc := ApiKeyServiceMock{
		AuthenticateApiKeyFunc: func(ctx context.Context, base64IdAndKeyCombination string) (apikeys.ApiKeyAuthenticationDTO, error) {
			assert.Equal(t, "MjMxNDMyNDM6bXl2YWxpZGFwaWtleQ==", base64IdAndKeyCombination)
			return apikeys.ApiKeyAuthenticationDTO{}, fmt.Errorf("database error!")
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("GET", "/api-keys/authenticate", nil)

	// Act
	req.Header["Authorization"] = []string{"Bearer MjMxNDMyNDM6bXl2YWxpZGFwaWtleQ=="}
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, `{"message":"Internal server error"}`+"\n", rr.Body.String())
	assert.Len(t, svc.AuthenticateApiKeyCalls(), 1)
}

func TestAuthenticateApiKeyIsNotFound(t *testing.T) {
	// Arrange
	svc := ApiKeyServiceMock{
		AuthenticateApiKeyFunc: func(ctx context.Context, base64IdAndKeyCombination string) (apikeys.ApiKeyAuthenticationDTO, error) {
			assert.Equal(t, "MjMxNDMyNDM6bXl2YWxpZGFwaWtleQ==", base64IdAndKeyCombination)
			return apikeys.ApiKeyAuthenticationDTO{}, apikeys.ErrKeyNotFound
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("GET", "/api-keys/authenticate", nil)

	// Act
	req.Header["Authorization"] = []string{"Bearer MjMxNDMyNDM6bXl2YWxpZGFwaWtleQ=="}
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Equal(t, `{}`+"\n", rr.Body.String())
	assert.Len(t, svc.AuthenticateApiKeyCalls(), 1)
}

func TestAuthenticateApiKeyInvalidEncodingErrorOccurs(t *testing.T) {
	// Arrange
	svc := ApiKeyServiceMock{
		AuthenticateApiKeyFunc: func(ctx context.Context, base64IdAndKeyCombination string) (apikeys.ApiKeyAuthenticationDTO, error) {
			assert.Equal(t, "blablabla", base64IdAndKeyCombination)
			return apikeys.ApiKeyAuthenticationDTO{}, apikeys.ErrInvalidEncoding
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("GET", "/api-keys/authenticate", nil)

	// Act
	req.Header["Authorization"] = []string{"Bearer blablabla"}
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"message":"Invalid input"}`+"\n", rr.Body.String())
	assert.Len(t, svc.AuthenticateApiKeyCalls(), 1)
}

func TestAuthenticateApiKeyIsValidNoExpirationDate(t *testing.T) {
	// Arrange
	svc := ApiKeyServiceMock{
		AuthenticateApiKeyFunc: func(ctx context.Context, base64IdAndKeyCombination string) (apikeys.ApiKeyAuthenticationDTO, error) {
			assert.Equal(t, "MjMxNDMyNDM6bXl2YWxpZGFwaWtleQ==", base64IdAndKeyCombination)
			return apikeys.ApiKeyAuthenticationDTO{
				TenantID:    431,
				Expiration:  nil,
				Permissions: auth.Permissions{},
			}, nil
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("GET", "/api-keys/authenticate", nil)

	// Act
	req.Header["Authorization"] = []string{"Bearer MjMxNDMyNDM6bXl2YWxpZGFwaWtleQ=="}
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"extra":{"perms":[],"tid":431}}`+"\n", rr.Body.String())
	assert.Len(t, svc.AuthenticateApiKeyCalls(), 1)
}

func TestAuthenticateApiKeyWithPermissions(t *testing.T) {
	// Arrange
	svc := ApiKeyServiceMock{
		AuthenticateApiKeyFunc: func(ctx context.Context, base64IdAndKeyCombination string) (apikeys.ApiKeyAuthenticationDTO, error) {
			assert.Equal(t, "MjMxNDMyNDM6bXl2YWxpZGFwaWtleQ==", base64IdAndKeyCombination)
			return apikeys.ApiKeyAuthenticationDTO{
				TenantID:    431,
				Expiration:  nil,
				Permissions: auth.Permissions{"READ_DEVICES", "WRITE_DEVICES", "READ_API_KEYS", "WRITE_API_KEYS"},
			}, nil
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("GET", "/api-keys/authenticate", nil)

	// Act
	req.Header["Authorization"] = []string{"Bearer MjMxNDMyNDM6bXl2YWxpZGFwaWtleQ=="}
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"extra":{"perms":["READ_DEVICES","WRITE_DEVICES","READ_API_KEYS","WRITE_API_KEYS"],"tid":431}}`+"\n", rr.Body.String())
	assert.Len(t, svc.AuthenticateApiKeyCalls(), 1)
}

func TestAuthenticateApiKeyIsValidWithExpirationDate(t *testing.T) {
	// Arrange
	exp := time.Now().Add(time.Minute).Unix()
	svc := ApiKeyServiceMock{
		AuthenticateApiKeyFunc: func(ctx context.Context, base64IdAndKeyCombination string) (apikeys.ApiKeyAuthenticationDTO, error) {
			assert.Equal(t, "MjMxNDMyNDM6bXl2YWxpZGFwaWtleQ==", base64IdAndKeyCombination)
			return apikeys.ApiKeyAuthenticationDTO{
				TenantID:    431,
				Expiration:  &exp,
				Permissions: auth.Permissions{},
			}, nil
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("GET", "/api-keys/authenticate", nil)

	// Act
	req.Header["Authorization"] = []string{"Bearer MjMxNDMyNDM6bXl2YWxpZGFwaWtleQ=="}
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, fmt.Sprintf(`{"extra":{"exp":%d,"perms":[],"tid":431}}`+"\n", exp), rr.Body.String())
	assert.Len(t, svc.AuthenticateApiKeyCalls(), 1)
}

func TestListApiKeysReturnsPaginatedList(t *testing.T) {
	// Arrange
	svc := ApiKeyServiceMock{
		ListAPIKeysFunc: func(ctx context.Context, filter apikeys.Filter, p pagination.Request) (*pagination.Page[apikeys.ApiKeyDTO], error) {
			return &pagination.Page[apikeys.ApiKeyDTO]{
				Cursor: "encoded_cursor",
				Data: []apikeys.ApiKeyDTO{
					{
						Name:        "api-key-1",
						Permissions: auth.Permissions{"READ_API_KEYS"},
					},
					{
						Name:        "api-key-2",
						Permissions: auth.Permissions{"READ_DEVICES", "WRITE_DEVICES"},
					},
				},
			}, nil
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("GET", "/api-keys", nil)

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t,
		`{"links":{"previous":"","next":"/api-keys?cursor=encoded_cursor"},"page_size":2,"total_count":0,"data":[{"id":0,"name":"api-key-1","tenant_id":0,"tenant_name":"","expiration_date":null,"created":"0001-01-01T00:00:00Z","permissions":["READ_API_KEYS"]},{"id":0,"name":"api-key-2","tenant_id":0,"tenant_name":"","expiration_date":null,"created":"0001-01-01T00:00:00Z","permissions":["READ_DEVICES","WRITE_DEVICES"]}]}`+"\n", rr.Body.String())
	assert.Len(t, svc.ListAPIKeysCalls(), 1)
}

func TestListApiKeysInvalidParams(t *testing.T) {
	// Arrange
	svc := ApiKeyServiceMock{}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("GET", "/api-keys?tenant_id=blablalq", nil)

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"message":"invalid params"}`+"\n", rr.Body.String())
	assert.Len(t, svc.ListAPIKeysCalls(), 0)
}

func TestListApiKeysErrorsOccursWhileRetrievingData(t *testing.T) {
	// Arrange
	svc := ApiKeyServiceMock{
		ListAPIKeysFunc: func(ctx context.Context, filter apikeys.Filter, p pagination.Request) (*pagination.Page[apikeys.ApiKeyDTO], error) {
			return nil, fmt.Errorf("weird database error!")
		},
	}
	transport := testTransport(&svc)
	req, _ := http.NewRequest("GET", "/api-keys", nil)

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t,
		`{"message":"Internal server error"}`+"\n", rr.Body.String())
	assert.Len(t, svc.ListAPIKeysCalls(), 1)
}

func testTransport(svc tenantstransports.ApiKeyService) *tenantstransports.APIKeysHTTPTransport {
	transport := tenantstransports.NewAPIKeysHTTP(chi.NewRouter(), svc, "")
	return transport
}
