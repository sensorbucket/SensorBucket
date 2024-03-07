package tenantstransports

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"sensorbucket.nl/sensorbucket/internal/httpfilter"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/services/tenants/tenants"
)

func NewTenantsHTTP(r chi.Router, tenantSvc TenantService, url string) *TenantsHTTPTransport {
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
	tenantSvc TenantService
	url       string
}

func (t *TenantsHTTPTransport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.router.ServeHTTP(w, r)
}

func (t *TenantsHTTPTransport) setupRoutes(r chi.Router) {
	r.Get("/tenants", t.httpGetTenants())
	r.Post("/tenants", t.httpCreateTenant())
	r.Delete("/tenants/{tenant_id}", t.httpDeleteTenant())
	r.Post("/tenants/{tenant_id}/members", t.httpAddTenantMember())
	r.Patch("/tenants/{tenant_id}/members/{user_id}", t.httpUpdateTenantMember())
	r.Delete("/tenants/{tenant_id}/members/{user_id}", t.httpDeleteTenantMember())
}

func (t *TenantsHTTPTransport) httpCreateTenant() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var dto tenants.CreateTenantDTO
		if err := web.DecodeJSON(r, &dto); err != nil {
			web.HTTPError(w, err)
			return
		}
		if validationErrors := ensureValuesNotEmpty(
			map[string]string{
				"name":     dto.Name,
				"address":  dto.Address,
				"zip_code": dto.ZipCode,
				"city":     dto.City,
			},
		); len(validationErrors) > 0 {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "model not valid",
				Data:    validationErrors,
			})
			return
		}
		created, err := t.tenantSvc.CreateNewTenant(r.Context(), dto)
		if err != nil {
			web.HTTPError(w, err)
			return
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
		page, err := t.tenantSvc.ListTenants(r.Context(), params.Filter, params.Request)
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
		if err := t.tenantSvc.ArchiveTenant(r.Context(), tenantID); err != nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
			Message: "Deleted tenant",
		})
	}
}

func (transport *TenantsHTTPTransport) httpAddTenantMember() http.HandlerFunc {
	type request struct {
		UserID      string           `json:"user_id"`
		Permissions auth.Permissions `json:"permissions"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		tenantIDStr := chi.URLParam(r, "tenant_id")
		tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
		if err != nil {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "tenant_id must be a number",
			})
			return
		}

		var req request
		if err := web.DecodeJSON(r, &req); err != nil {
			web.HTTPError(w, err)
			return
		}
		if len(req.Permissions) == 0 {
			req.Permissions = make(auth.Permissions, 0)
		}

		if err := transport.tenantSvc.AddTenantMember(r.Context(), tenantID, req.UserID, req.Permissions); err != nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusCreated, web.APIResponseAny{
			Message: "User added to tenant",
		})
	}
}

func (transport *TenantsHTTPTransport) httpUpdateTenantMember() http.HandlerFunc {
	type request struct {
		Permissions auth.Permissions `json:"permissions"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		tenantIDStr := chi.URLParam(r, "tenant_id")
		tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
		if err != nil {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "tenant_id must be a number",
			})
			return
		}
		userID := chi.URLParam(r, "user_id")

		var req request
		if err := web.DecodeJSON(r, &req); err != nil {
			web.HTTPError(w, err)
			return
		}
		if len(req.Permissions) == 0 {
			web.HTTPError(w, web.NewError(http.StatusBadRequest, "Missing 'permissions' field", ""))
		}

		if err := transport.tenantSvc.UpdateTenantMember(r.Context(), tenantID, userID, req.Permissions); err != nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
			Message: "Member permissions updated tenant",
		})
	}
}

func (transport *TenantsHTTPTransport) httpDeleteTenantMember() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantIDStr := chi.URLParam(r, "tenant_id")
		tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
		if err != nil {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "tenant_id must be a number",
			})
			return
		}
		userID := chi.URLParam(r, "user_id")

		if err := transport.tenantSvc.RemoveTenantMember(r.Context(), tenantID, userID); err != nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
			Message: "User removed from tenant",
		})
	}
}

func ensureValuesNotEmpty[T comparable](values map[string]T) []string {
	validationErrorMsg := func(name string) string {
		return fmt.Sprintf("%s must be set", name)
	}
	validationErrors := []string{}
	var empty T
	for name, val := range values {
		if val == empty {
			validationErrors = append(validationErrors, validationErrorMsg(name))
		}
	}
	return validationErrors
}

type TenantService interface {
	CreateNewTenant(ctx context.Context, tenant tenants.CreateTenantDTO) (tenants.CreateTenantDTO, error)
	ArchiveTenant(ctx context.Context, tenantID int64) error
	ListTenants(ctx context.Context, filter tenants.Filter, p pagination.Request) (*pagination.Page[tenants.CreateTenantDTO], error)
	AddTenantMember(ctx context.Context, tenantID int64, userID string, permissions auth.Permissions) error
	RemoveTenantMember(ctx context.Context, tenantID int64, userID string) error
	UpdateTenantMember(ctx context.Context, tenantID int64, userID string, permissions auth.Permissions) error
}
