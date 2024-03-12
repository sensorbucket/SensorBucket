package tenantstransports_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/services/tenants/tenants"
	tenantstransports "sensorbucket.nl/sensorbucket/services/tenants/transports"
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
			svc := TenantServiceMock{}
			transport := testTenantsTransport(&svc)
			req, _ := http.NewRequest("POST", "/tenants", strings.NewReader(cfg.input))
			req.Header.Set("content-type", "application/json")

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
	svc := TenantServiceMock{}
	transport := testTenantsTransport(&svc)
	req, _ := http.NewRequest("POST", "/tenants", strings.NewReader("invalid json!!"))
	req.Header.Set("content-type", "application/json")

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"message":"Malformed JSON","code":"MALFORMED_JSON"}`+"\n", rr.Body.String())
	assert.Len(t, svc.CreateNewTenantCalls(), 0)
}

func TestCreateTenantParentTenantDoesNotExist(t *testing.T) {
	// Arrange
	svc := TenantServiceMock{
		CreateNewTenantFunc: func(ctx context.Context, tenant tenants.CreateTenantDTO) (tenants.CreateTenantDTO, error) {
			assert.NotNil(t, tenant.ParentID)
			assert.Equal(t, int64(345), *tenant.ParentID)
			return tenants.CreateTenantDTO{}, tenants.ErrTenantNotFound
		},
	}
	transport := testTenantsTransport(&svc)
	req, _ := http.NewRequest("POST", "/tenants", strings.NewReader(`{
		"name": "some-org",
		"chamber_of_commerce_id": "asdasdasd",
		"address": "sadasdasd",
		"zip_code": "sad",
		"parent_tenant_id":345,
		"city": "Breda",
		"headquarter_id": "sadasda"
	}`))
	req.Header.Set("content-type", "application/json")

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Equal(t, `{"message":"Tenant could not be found","code":"TENANT_NOT_FOUND"}`+"\n", rr.Body.String())
	assert.Len(t, svc.CreateNewTenantCalls(), 1)
}

func TestCreateTenantTenantIsCreated(t *testing.T) {
	// Arrange
	svc := TenantServiceMock{
		CreateNewTenantFunc: func(ctx context.Context, tenant tenants.CreateTenantDTO) (tenants.CreateTenantDTO, error) {
			assert.NotNil(t, tenant.ParentID)
			assert.Equal(t, int64(345), *tenant.ParentID)
			t := tenants.CreateTenantDTO{
				ID: 124,
			}
			return t, nil
		},
	}
	transport := testTenantsTransport(&svc)
	req, _ := http.NewRequest("POST", "/tenants", strings.NewReader(`{
		"name": "some-org",
		"chamber_of_commerce_id": "asdasdasd",
		"address": "sadasdasd",
		"zip_code": "sad",
		"parent_tenant_id":345,
		"city": "Breda",
		"headquarter_id": "sadasda"
	}`))
	req.Header.Set("content-type", "application/json")

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, `{"message":"Created new tenant","data":{"id":124,"name":"","address":"","zip_code":"","city":"","chamber_of_commerce_id":null,"headquarter_id":null,"archive_time":null,"logo":null,"parent_tenant_id":null}}`+"\n", rr.Body.String(), "\n")
	assert.Len(t, svc.CreateNewTenantCalls(), 1)
}

func TestDeleteTenantTenantIDIsNotAnInt(t *testing.T) {
	// Arrange
	svc := TenantServiceMock{}
	transport := testTenantsTransport(&svc)
	req, _ := http.NewRequest("DELETE", "/tenants/asdasd", nil)

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"message":"tenant_id must be a number"}`+"\n", rr.Body.String(), "\n")
	assert.Len(t, svc.ArchiveTenantCalls(), 0)
}

func TestDeleteTenantTenantDoesNotExist(t *testing.T) {
	// Arrange
	svc := TenantServiceMock{
		ArchiveTenantFunc: func(ctx context.Context, tenantID int64) error {
			assert.Equal(t, int64(12345), tenantID)
			return tenants.ErrTenantNotFound
		},
	}
	transport := testTenantsTransport(&svc)
	req, _ := http.NewRequest("DELETE", "/tenants/12345", nil)

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Equal(t, `{"message":"Tenant could not be found","code":"TENANT_NOT_FOUND"}`+"\n", rr.Body.String(), "\n")
	assert.Len(t, svc.ArchiveTenantCalls(), 1)
}

func TestDeleteTenantErrorOccursWhileDeleting(t *testing.T) {
	// Arrange
	svc := TenantServiceMock{
		ArchiveTenantFunc: func(ctx context.Context, tenantID int64) error {
			assert.Equal(t, int64(12345), tenantID)
			return fmt.Errorf("weird db error!!")
		},
	}
	transport := testTenantsTransport(&svc)
	req, _ := http.NewRequest("DELETE", "/tenants/12345", nil)

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, `{"message":"Internal server error"}`+"\n", rr.Body.String(), "\n")
	assert.Len(t, svc.ArchiveTenantCalls(), 1)
}

func TestDeleteTenantTenantIsDeleted(t *testing.T) {
	// Arrange
	svc := TenantServiceMock{
		ArchiveTenantFunc: func(ctx context.Context, tenantID int64) error {
			assert.Equal(t, int64(12345), tenantID)
			return nil
		},
	}
	transport := testTenantsTransport(&svc)
	req, _ := http.NewRequest("DELETE", "/tenants/12345", nil)

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"message":"Deleted tenant"}`+"\n", rr.Body.String(), "\n")
	assert.Len(t, svc.ArchiveTenantCalls(), 1)
}

func TestListTenantsInvalidParams(t *testing.T) {
	// Arrange
	svc := TenantServiceMock{}
	transport := testTenantsTransport(&svc)
	req, _ := http.NewRequest("GET", "/tenants?state=asdasd", nil)

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"message":"invalid params"}`+"\n", rr.Body.String(), "\n")
	assert.Len(t, svc.ListTenantsCalls(), 0)
}

func TestListTenantsErrorOccursWhileRetrievingData(t *testing.T) {
	// Arrange
	svc := TenantServiceMock{
		ListTenantsFunc: func(ctx context.Context, filter tenants.Filter, p pagination.Request) (*pagination.Page[tenants.CreateTenantDTO], error) {
			return nil, fmt.Errorf("weird error occurred!")
		},
	}
	transport := testTenantsTransport(&svc)
	req, _ := http.NewRequest("GET", "/tenants", nil)

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, `{"message":"Internal server error"}`+"\n", rr.Body.String(), "\n")
	assert.Len(t, svc.ListTenantsCalls(), 1)
}

func TestListTenantsReturnsListOfTenants(t *testing.T) {
	// Arrange
	svc := TenantServiceMock{
		ListTenantsFunc: func(ctx context.Context, filter tenants.Filter, p pagination.Request) (*pagination.Page[tenants.CreateTenantDTO], error) {
			return &pagination.Page[tenants.CreateTenantDTO]{
				Cursor: "asdasdads",
				Data: []tenants.CreateTenantDTO{
					{
						Name: "blabla",
					},
					{
						Name: "ewrtras",
					},
				},
			}, nil
		},
	}
	transport := testTenantsTransport(&svc)
	req, _ := http.NewRequest("GET", "/tenants", nil)

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"links":{"previous":"","next":"/tenants?cursor=asdasdads"},"page_size":2,"total_count":0,"data":[{"id":0,"name":"blabla","address":"","zip_code":"","city":"","chamber_of_commerce_id":null,"headquarter_id":null,"archive_time":null,"logo":null,"parent_tenant_id":null},{"id":0,"name":"ewrtras","address":"","zip_code":"","city":"","chamber_of_commerce_id":null,"headquarter_id":null,"archive_time":null,"logo":null,"parent_tenant_id":null}]}`+"\n", rr.Body.String(), "\n")
	assert.Len(t, svc.ListTenantsCalls(), 1)
}

func TestAddTenantMember(t *testing.T) {
	// Arrange
	svc := TenantServiceMock{
		AddTenantMemberFunc: func(ctx context.Context, tenantID int64, userID string, permissions auth.Permissions) error {
			return nil
		},
	}
	transport := testTenantsTransport(&svc)
	req, _ := http.NewRequest("POST", "/tenants/15/members", strings.NewReader(`
        {"user_id": "useridhere", "permissions": ["WRITE_TENANTS"]}
    `))
	req.Header.Set("content-type", "application/json")

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Contains(t, rr.Body.String(), "User added to tenant")
	assert.Len(t, svc.AddTenantMemberCalls(), 1)
	call := svc.AddTenantMemberCalls()[0]
	assert.Equal(t, call.UserID, "useridhere")
	assert.EqualValues(t, call.TenantID, 15)
	assert.Equal(t, call.Permissions, auth.Permissions{auth.WRITE_TENANTS})
}

func TestAddTenantMemberInvalidTenantID(t *testing.T) {
	// Arrange
	svc := TenantServiceMock{
		AddTenantMemberFunc: func(ctx context.Context, tenantID int64, userID string, permissions auth.Permissions) error {
			return nil
		},
	}
	transport := testTenantsTransport(&svc)
	req, _ := http.NewRequest("POST", "/tenants/asdasd/members", strings.NewReader(`
        {"user_id": "", "permissions": []}
    `))
	req.Header.Set("content-type", "application/json")

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "must be a number")
	assert.Len(t, svc.AddTenantMemberCalls(), 0)
}

func TestUpdateTenantMember(t *testing.T) {
	// Arrange
	svc := TenantServiceMock{
		UpdateTenantMemberFunc: func(ctx context.Context, tenantID int64, userID string, permissions auth.Permissions) error {
			return nil
		},
	}
	transport := testTenantsTransport(&svc)
	req, _ := http.NewRequest("PATCH", "/tenants/15/members/useridhere", strings.NewReader(`
        {"permissions": ["WRITE_TENANTS"]}
    `))
	req.Header.Set("content-type", "application/json")

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "permissions updated")
	require.Len(t, svc.UpdateTenantMemberCalls(), 1)
	call := svc.UpdateTenantMemberCalls()[0]
	assert.Equal(t, call.UserID, "useridhere")
	assert.EqualValues(t, call.TenantID, 15)
	assert.Equal(t, call.Permissions, auth.Permissions{auth.WRITE_TENANTS})
}

func TestDeleteTenantMember(t *testing.T) {
	// Arrange
	svc := TenantServiceMock{
		RemoveTenantMemberFunc: func(ctx context.Context, tenantID int64, userID string) error {
			return nil
		},
	}
	transport := testTenantsTransport(&svc)
	req, _ := http.NewRequest("DELETE", "/tenants/15/members/useridhere", nil)

	// Act
	rr := httptest.NewRecorder()
	transport.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "User removed from tenant")
	require.Len(t, svc.RemoveTenantMemberCalls(), 1)
	call := svc.RemoveTenantMemberCalls()[0]
	assert.Equal(t, call.UserID, "useridhere")
	assert.EqualValues(t, call.TenantID, 15)
}

func testTenantsTransport(svc tenantstransports.TenantService) *tenantstransports.TenantsHTTPTransport {
	return tenantstransports.NewTenantsHTTP(chi.NewRouter(), svc, "")
}
