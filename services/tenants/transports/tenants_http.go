package tenantstransports

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"sensorbucket.nl/sensorbucket/internal/httpfilter"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/tenants/tenants"
)

func NewTenantsHTTP(r chi.Router, tenantSvc tenantService, url string) *TenantsHTTPTransport {
	t := &TenantsHTTPTransport{
		router:    r,
		tenantSvc: tenantSvc,
		url:       url,
	}
	t.setupRoutes(t.router)
	return t
}

type TenantsHTTPTransport struct {
	router    chi.Router
	tenantSvc tenantService
	url       string
}

func (t *TenantsHTTPTransport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.router.ServeHTTP(w, r)
}

func (t *TenantsHTTPTransport) setupRoutes(r chi.Router) {
	r.Get("/tenants/list", t.httpGetTenants())
	r.Post("/tenants", t.httpCreateTenant())
	r.Delete("/tenants/{tenant_id}", t.httpDeleteTenant())
}

func (t *TenantsHTTPTransport) httpCreateTenant() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dto := tenants.TenantDTO{}
		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(&dto)
		if err != nil {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "Invalid JSON body",
			})
			return
		}
		if validationErrors := ensureValuesNotEmptyOrZero(
			map[string]interface{}{
				"name":                   dto.Name,
				"address":                dto.Address,
				"zip_code":               dto.ZipCode,
				"city":                   dto.City,
				"chamber_of_commerce_id": dto.ChamberOfCommerceID,
				"headquarter_id":         dto.HeadquarterID,
			},
		); len(validationErrors) > 0 {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "model not valid",
				Data:    validationErrors,
			})
			return
		}
		created, err := t.tenantSvc.CreateNewTenant(dto)
		if err != nil {
			if errors.Is(err, tenants.ErrParentTenantNotFound) {
				web.HTTPResponse(w, http.StatusNotFound, web.APIResponseAny{
					Message: "Parent tenant could not be found",
				})
				return
			} else {
				web.HTTPError(w, err)
				return
			}
		}
		web.HTTPResponse(w, http.StatusCreated, web.APIResponseAny{
			Message: "Created new tenant",
			Data:    created,
		})
	}
}

func (t *TenantsHTTPTransport) httpGetTenants() http.HandlerFunc {
	type Params struct {
		tenants.Filter     `pagination:",squash"`
		pagination.Request `pagination:",squash"`
	}
	return func(rw http.ResponseWriter, r *http.Request) {
		params, err := httpfilter.Parse[Params](r)
		if err != nil {
			web.HTTPError(rw, web.NewError(http.StatusBadRequest, "invalid params", ""))
			return
		}
		page, err := t.tenantSvc.ListTenants(params.Filter, params.Request)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}
		web.HTTPResponse(rw, http.StatusOK, pagination.CreateResponse(r, t.url, *page))
	}
}

func (t *TenantsHTTPTransport) httpDeleteTenant() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantIDStr := chi.URLParam(r, "tenant_id")
		tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
		if err != nil {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "tenant_id must be a number",
			})
			return
		}
		if err := t.tenantSvc.ArchiveTenant(tenantID); err != nil {
			if errors.Is(err, tenants.ErrTenantNotFound) {
				web.HTTPResponse(w, http.StatusNotFound, web.APIResponseAny{
					Message: "Tenant does not exist",
				})
				return
			} else {
				web.HTTPError(w, err)
				return
			}
		}
		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
			Message: "Deleted tenant",
		})
	}
}

func ensureValuesNotEmptyOrZero(values map[string]interface{}) []string {
	validationErrorMsg := func(name string) string {
		return fmt.Sprintf("%s must be set", name)
	}
	validationErrors := []string{}
	for name, val := range values {
		switch val.(type) {
		case string:
			if val == "" {
				validationErrors = append(validationErrors, validationErrorMsg(name))
			}
		case int64:
			if val == int64(0) {
				validationErrors = append(validationErrors, validationErrorMsg(name))
			}
		default:
			validationErrors = append(validationErrors, fmt.Sprintf("%s is of invalid type", name))
		}
	}
	return validationErrors
}

type tenantService interface {
	CreateNewTenant(tenant tenants.TenantDTO) (tenants.TenantDTO, error)
	ArchiveTenant(tenantID int64) error
	ListTenants(filter tenants.Filter, p pagination.Request) (*pagination.Page[tenants.TenantDTO], error)
}
