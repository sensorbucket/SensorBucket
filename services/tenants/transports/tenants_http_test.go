package tenantstransports

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"sensorbucket.nl/sensorbucket/services/tenants/tenants"
)

func TestCreateTenantModelNotValid(t *testing.T) {
	type scene struct {
		input    string
		expected []string
	}

	contains := func(arr1 []string, arr2 []string) bool {
		for _, val1 := range arr1 {
			found := false
			for _, val2 := range arr2 {
				if val1 == val2 {
					found = true
				}
			}
			if !found {
				return false
			}
		}
		return true
	}

	// Arrange
	scenarios := map[string]scene{
		"name is missing": {
			input: `{
				"name": "",
				"chamber_of_commerce_id": "asdasdasd",
				"address": "sadasdasd",
				"zip_code": "13445",
				"city": "Breda",
				"headquarter_id": "sadasdsadasd"
			}`,
			expected: []string{"name must be set"},
		},
		"chamber_of_commerce_id is missing": {
			input: `{
				"name": "some-name",
				"chamber_of_commerce_id": "",
				"address": "sadasdasd",
				"zip_code": "13445",
				"city": "Breda",
				"headquarter_id": "sadasdsadasd"
			}`,
			expected: []string{"chamber_of_commerce_id must be set"},
		},
		"address is missing": {
			input: `{
				"name": "some-name",
				"chamber_of_commerce_id": "asdasdasd",
				"address": "",
				"zip_code": "13445",
				"city": "Breda",
				"headquarter_id": "sadasdsadasd"
			}`,
			expected: []string{"address must be set"},
		},
		"zip_code is missing": {
			input: `{
				"name": "asdsad",
				"chamber_of_commerce_id": "asdasdasd",
				"address": "sadasdasd",
				"zip_code": "",
				"city": "Breda",
				"headquarter_id": "sadasdsadasd"
			}`,
			expected: []string{"zip_code must be set"},
		},
		"city is missing": {
			input: `{
				"name": "asdsad",
				"chamber_of_commerce_id": "asdasdasd",
				"address": "sadasdasd",
				"zip_code": "13445",
				"city": "",
				"headquarter_id": "sadasdsadasd"
			}`,
			expected: []string{"city must be set"},
		},
		"headquarter_id is missing": {
			input: `{
				"name": "asdasfasd",
				"chamber_of_commerce_id": "asdasdasd",
				"address": "sadasdasd",
				"zip_code": "13445",
				"city": "Breda",
				"headquarter_id": ""
			}`,
			expected: []string{"headquarter_id must be set"},
		},
		"multiple values are missing": {
			input: `{
				"name": "",
				"chamber_of_commerce_id": "asdasdasd",
				"address": "sadasdasd",
				"zip_code": "",
				"city": "Breda",
				"headquarter_id": ""
			}`,
			expected: []string{"headquarter_id must be set", "name must be set", "zip_code must be set"},
		},
	}

	for scene, cfg := range scenarios {
		t.Run(scene, func(t *testing.T) {
			svc := tenantServiceMock{}
			transport := testTenantsTransport(&svc)
			req, _ := http.NewRequest("POST", "/tenants/create", strings.NewReader(cfg.input))

			// Act
			rr := httptest.NewRecorder()
			transport.ServeHTTP(rr, req)

			// Assert
			assert.Equal(t, http.StatusBadRequest, rr.Code)
			type expectedResp struct {
				Message string
				Data    []string
			}
			e := expectedResp{}
			_ = json.NewDecoder(rr.Body).Decode(&e)
			assert.Equal(t, "model not valid", e.Message)
			assert.True(t, contains(e.Data, cfg.expected))
		})
	}
}

func TestCreateTenantInvalidJSONBody(t *testing.T) {
	// Arrange
	svc := tenantServiceMock{}
	transport := testTenantsTransport(&svc)
	req, _ := http.NewRequest("POST", "/tenants/create", strings.NewReader("invalid json!!"))

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"message":"Invalid JSON body"}`+"\n", rr.Body.String())
	assert.Len(t, svc.CreateNewTenantCalls(), 0)
}

func TestCreateTenantParentTenantDoesNotExist(t *testing.T) {
	// Arrange
	svc := tenantServiceMock{
		CreateNewTenantFunc: func(tenant tenants.TenantDTO) (*tenants.TenantDTO, error) {
			assert.NotNil(t, tenant.ParentID)
			assert.Equal(t, int64(345), *tenant.ParentID)
			return nil, tenants.ErrParentTenantNotFound
		},
	}
	transport := testTenantsTransport(&svc)
	req, _ := http.NewRequest("POST", "/tenants/create", strings.NewReader(`{
		"name": "some-org",
		"chamber_of_commerce_id": "asdasdasd",
		"address": "sadasdasd",
		"zip_code": "sad",
		"parent_tenant_id":345,
		"city": "Breda",
		"headquarter_id": "sadasda"
	}`))

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Equal(t, `{"message":"Parent tenant could not be found"}`+"\n", rr.Body.String())
	assert.Len(t, svc.CreateNewTenantCalls(), 1)
}

func TestCreateTenantTenantIsCreated(t *testing.T) {
	// Arrange
	svc := tenantServiceMock{
		CreateNewTenantFunc: func(tenant tenants.TenantDTO) (*tenants.TenantDTO, error) {
			assert.NotNil(t, tenant.ParentID)
			assert.Equal(t, int64(345), *tenant.ParentID)
			t := tenants.TenantDTO{}
			return &t, nil
		},
	}
	transport := testTenantsTransport(&svc)
	req, _ := http.NewRequest("POST", "/tenants/create", strings.NewReader(`{
		"name": "some-org",
		"chamber_of_commerce_id": "asdasdasd",
		"address": "sadasdasd",
		"zip_code": "sad",
		"parent_tenant_id":345,
		"city": "Breda",
		"headquarter_id": "sadasda"
	}`))

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, strings.Trim(
		`{
			"message":"Created new tenant", 
			"data": {
				"name":"",
				"address":"",
				"zip_code":"",
				"city":"",
				"chamber_of_commerce_id":"",
				"headquarter_id":"",
				"archive_time":null,
				"logo":null,
				"parent_tenant_id":null
		}}`+"\n", rr.Body.String()), "\n")
	assert.Len(t, svc.CreateNewTenantCalls(), 1)
}

func testTenantsTransport(svc tenantService) *TenantsHTTPTransport {
	transport := &TenantsHTTPTransport{
		tenantSvc: svc,
		router:    chi.NewMux(),
	}
	transport.setupRoutes(transport.router)
	return transport
}