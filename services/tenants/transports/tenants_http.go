package tenantstransports

import (
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
	r.Post("/tenants/member-permissions", t.httpAddMemberPermission())
	r.Delete("/tenants/members-permissions", t.httpRevokeMemberPermission())
	r.Delete("/tenants/{tenant_id}", t.httpDeleteTenant())
}

func (t *TenantsHTTPTransport) httpCreateTenant() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dto := tenants.TenantDTO{}
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
				Message: "Model not valid",
				Data:    validationErrors,
			})
			return
		}
		if len(dto.Permissions) == 0 {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "Tenant should have at least 1 permission",
			})
			return
		}
		created, err := t.tenantSvc.CreateNewTenant(dto)
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

func (t *TenantsHTTPTransport) httpRevokeMemberPermission() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dto := tenants.MemberPermissionsMutationDTO{}
		if err := web.DecodeJSON(r, &dto); err != nil {
			web.HTTPError(w, err)
			return
		}
		if validationErrors := ensureValuesNotEmpty(
			map[string]int64{
				"tenant_id": dto.TenantID,
				"user_id":   dto.UserID,
			},
		); len(validationErrors) > 0 {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "model not valid",
				Data:    validationErrors,
			})
			return
		}
		if len(dto.Permissions) == 0 {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "at least 1 permission must be set",
			})
			return
		}

		err := t.tenantSvc.DeleteMemberPermissions(dto)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusCreated, web.APIResponseAny{
			Message: "Revoked permissions",
		})
	}
}

// TODO: check if user_id exists by asking ory kratos?
func (t *TenantsHTTPTransport) httpAddMemberPermission() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dto := tenants.MemberPermissionsMutationDTO{}
		if err := web.DecodeJSON(r, &dto); err != nil {
			web.HTTPError(w, err)
			return
		}
		if validationErrors := ensureValuesNotEmpty(
			map[string]int64{
				"tenant_id": dto.TenantID,
				"user_id":   dto.UserID,
			},
		); len(validationErrors) > 0 {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "model not valid",
				Data:    validationErrors,
			})
			return
		}
		if len(dto.Permissions) == 0 {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "at least 1 permission must be set",
			})
			return
		}
		fmt.Println(dto)
		created, err := t.tenantSvc.AddMemberPermissions(dto)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusCreated, web.APIResponseAny{
			Message: "Added permissions",
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
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
			Message: "Deleted tenant",
		})
	}
}

func ensureValuesNotEmpty[T comparable](values map[string]T) []string {
	validationErrorMsg := func(name string) string {
		return fmt.Sprintf("%s must be set", name)
	}
	validationErrors := []string{}
	for name, val := range values {
		empty := *new(T)
		if val == empty {
			validationErrors = append(validationErrors, validationErrorMsg(name))
		}
	}
	return validationErrors
}

type tenantService interface {
	CreateNewTenant(tenant tenants.TenantDTO) (tenants.TenantDTO, error)
	AddMemberPermissions(memberPermissions tenants.MemberPermissionsMutationDTO) (tenants.MemberPermissionsAddedDTO, error)
	DeleteMemberPermissions(memberPermissions tenants.MemberPermissionsMutationDTO) error
	ArchiveTenant(tenantID int64) error
	ListTenants(filter tenants.Filter, p pagination.Request) (*pagination.Page[tenants.TenantDTO], error)
}
