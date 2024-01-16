package routes

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/samber/lo"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/api"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/pkg/layout"
	"sensorbucket.nl/sensorbucket/services/tenants/transports/webui/views"
)

type APIKeysPageHandler struct {
	router chi.Router
	client *api.APIClient
}

func CreateAPIKeysPageHandler(client *api.APIClient) *APIKeysPageHandler {
	handler := &APIKeysPageHandler{
		router: chi.NewRouter(),
		client: client,
	}
	handler.SetupRoutes(handler.router)
	return handler
}

func (h APIKeysPageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *APIKeysPageHandler) SetupRoutes(r chi.Router) {
	r.Get("/", h.apiKeysListPage())
	r.Get("/table", h.apiKeysGetTableRows())
	r.Get("/create", h.createAPIKeyView())
	r.Post("/create", h.createAPIKey())
	r.Delete("/revoke/{api_key_id}", h.revokeAPIKey())
}

func (h *APIKeysPageHandler) apiKeysGetTableRows() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := h.client.ApiKeysApi.ListApiKeys(r.Context())
		if r.URL.Query().Has("cursor") {
			req = req.Cursor(r.URL.Query().Get("cursor"))
		} else {
			// If no cursor is given, the initial state must be derived from the url params
			req = req.Limit(15)
			tenantId := r.URL.Query().Get("tenant_id")
			id, err := strconv.ParseInt(tenantId, 10, 64)
			if err != nil {
				layout.SnackbarBadRequest(w, "tenant_id must be a valid number")
				return
			}
			req = req.TenantId(id)
		}
		res, _, err := req.Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		nextPage := ""
		if res.Links.GetNext() != "" {
			nextPage = layout.U("/api-keys/table?cursor=" + getCursor(res.Links.GetNext()))
		}

		viewKeys := []views.APIKey{}
		for _, key := range res.Data {
			viewKeys = append(viewKeys, views.APIKey{
				ID:             int(key.Id),
				TenantID:       int(key.TenantId),
				Created:        key.Created,
				ExpirationDate: key.ExpirationDate,
				Name:           key.Name,
			})
		}
		views.WriteRenderAPIKeyRows(w, viewKeys, nextPage)
	}
}

func (h *APIKeysPageHandler) apiKeysListPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// The initial page starts with an overview of different tenants
		req := h.client.TenantsApi.ListTenants(r.Context())
		req = req.State(1) // State Active
		list, resp, err := req.Execute()
		if err != nil {
			log.Printf("api key list page: %v\n", err)
			layout.SnackbarSomethingWentWrong(w)
			return
		}
		if resp.StatusCode != http.StatusOK {
			log.Printf("api key list page unexpected status code: %d\n", resp.StatusCode)
			layout.SnackbarSomethingWentWrong(w)
			return
		}

		page := &views.APIKeysPage{Tenants: toViewTenants(list.Data)}
		if apiKey, ok := r.Context().Value("key").(string); ok {
			page.CreatedAPIKey = apiKey
		}

		if layout.IsHX(r) {
			page.WriteBody(w)
			return
		}
		layout.WriteIndex(w, page)
	}
}

func (h *APIKeysPageHandler) revokeAPIKey() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKeyId := chi.URLParam(r, "api_key_id")
		if apiKeyId == "" {
			layout.WithSnackbarError(w, "api_key_id must be given", http.StatusBadRequest)
			return
		}
		id, err := strconv.ParseInt(apiKeyId, 10, 64)
		if err != nil {
			layout.SnackbarBadRequest(w, "api_key_id must be a number")
			return
		}
		req := h.client.ApiKeysApi.RevokeApiKey(context.Background(), id)
		resp, err := req.Execute()
		if err != nil {
			log.Printf("revoke api key execute: %v\n", err)
			layout.SnackbarSomethingWentWrong(w)
			return
		}
		if resp.StatusCode != http.StatusOK {
			log.Printf("revoke api status code not ok: %d\n", resp.StatusCode)
			layout.SnackbarSomethingWentWrong(w)
			return
		}
		layout.SnackbarDeleteSuccessful(w)
	}
}

func (h *APIKeysPageHandler) createAPIKey() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			layout.SnackbarBadRequest(w, "Please select a tenant and enter a name")
			return
		}
		name := r.FormValue("api-key-name")
		expiry := r.FormValue("api-key-expiry")
		tenantId := r.FormValue("api-key-tenant")
		permissions, ok := r.Form["api-key-permissions"]
		if name == "" || tenantId == "" || !ok || len(permissions) == 0 {
			layout.SnackbarBadRequest(w, "Please enter a name, select an organisation and at least 1 permission")
			return
		}

		id, err := strconv.ParseInt(tenantId, 10, 64)
		if err != nil {
			layout.SnackbarBadRequest(w, "tenant_id must be a valid number")
			return
		}

		var dto api.CreateApiKeyRequest
		dto.SetName(name)
		dto.SetTenantId(id)

		// Expiry is an optional value
		if expiry != "" {
			parsedTime, err := time.Parse("2006-01-02", expiry)
			if err != nil {
				layout.SnackbarBadRequest(w, "expiration_date must be a valid time format")
				return
			}

			dto.SetExpirationDate(parsedTime)
		}
		apiKey, res, err := h.client.ApiKeysApi.CreateApiKey(r.Context()).CreateApiKeyRequest(dto).Execute()
		if err != nil {
			log.Printf("[Error] couldnt create api key, err: %s\n", err)
			layout.SnackbarSomethingWentWrong(w)
			return
		}

		if res.StatusCode != http.StatusCreated {
			log.Printf("[Error] couldnt create api key, response: %d\n", res.StatusCode)
			layout.SnackbarSomethingWentWrong(w)
			return
		}

		// Redirect like this instead of using HX-Location or HX-Redirect because we don't want the API key to be
		// sent the the HX-Location/Redirect endpoint in the request for further handling, there might be some browser
		// caching involved which can store the API key. This way we ensure the API key is included in the response for this request
		// along with the api keys list
		w.Header().Set("HX-Replace-Url", "/api-keys")
		layout.WithSnackbarSuccess(w, "Created API Key")
		h.apiKeysListPage().ServeHTTP(
			w,
			r.WithContext(context.WithValue(r.Context(), "key", apiKey.ApiKey)))
	}
}

func (h *APIKeysPageHandler) createAPIKeyView() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Retrieve the tenants for the select box in the create view
		req := h.client.TenantsApi.ListTenants(r.Context())
		req = req.State(1) // State Active
		list, resp, err := req.Execute()
		if err != nil {
			log.Printf("api key list page: %v\n", err)
			layout.SnackbarSomethingWentWrong(w)
			return
		}
		if resp.StatusCode != http.StatusOK {
			log.Printf("api key list failed : %d\n", resp.StatusCode)
			layout.SnackbarSomethingWentWrong(w)
			return
		}

		perms, err := toViewPermissions()
		if err != nil {
			log.Printf("convert to view permissions, err: %s\n", err)
			layout.SnackbarSomethingWentWrong(w)
			return
		}

		page := &views.APIKeysCreatePage{
			Tenants:     toViewTenants(list.Data),
			Permissions: perms,
		}
		if layout.IsHX(r) {
			page.WriteBody(w)
			return
		}
		layout.WriteIndex(w, page)
	}
}

func toViewPermissions() (map[string][]views.APIKeysPermission, error) {
	categorized := createAPIKeyViewPermissions()

	// Retrieve all allowed permissions
	inAuth := auth.AllAllowedPermissions()
	authAsStrSlice := []string{}
	for _, p := range inAuth {
		authAsStrSlice = append(authAsStrSlice, p.String())
	}
	viewAsSlice := lo.Map(lo.Flatten(lo.Values(categorized)), func(view views.APIKeysPermission, index int) string {
		return view.Name
	})

	// Check if there is any difference between the frontend's definition and the auth package definition
	missingInView, missingInAuth := lo.Difference(authAsStrSlice, viewAsSlice)
	if len(missingInAuth) > 0 {

		// Auth package is the single point of truth, if there are permissions missing the frontend package might have
		// outdated permission which the user should not be able to set
		return nil, fmt.Errorf("view permissions contains invalid permissions (%d invalid permissions)", len(missingInAuth))
	}
	if len(missingInView) > 0 {

		// Frontend might not be updated with latest permissions, simply log a warning since we can still show the missing permissions
		// just without a category and description
		log.Printf("[Warning] some permissions are missing in create view (%d missing permissions)\n", len(missingInView))
	}
	if len(missingInView) > 0 {
		categorized["Other"] = lo.Map(missingInView, func(val string, index int) views.APIKeysPermission {
			return views.APIKeysPermission{
				Name:        val,
				Description: "-",
			}
		})
	}
	return categorized, nil
}

func createAPIKeyViewPermissions() map[string][]views.APIKeysPermission {
	return map[string][]views.APIKeysPermission{
		"Devices": {
			{
				Name:        auth.READ_DEVICES.String(),
				Description: "Allows the API key to read information regarding devices of the selected organisation.",
			},
			{
				Name:        auth.WRITE_DEVICES.String(),
				Description: "Allows the API key to write information regarding devices of the selected organisation.",
			},
		},
		"API Keys": {
			{
				Name:        auth.READ_API_KEYS.String(),
				Description: "Does not allow reading of actual API keys. Only allowes the API key to read information certain information, for example, the expiration date.",
			},
			{
				Name:        auth.WRITE_API_KEYS.String(),
				Description: "Allows the API key to create other API keys for the tenant the API key has access to.",
			},
		},
		"Tenants": {
			{
				Name:        auth.READ_TENANTS.String(),
				Description: "Allows the API key to read information about the tenant they have access to.",
			},
			{
				Name:        auth.WRITE_TENANTS.String(),
				Description: "Allows the API key to create child organisations for the tenant this API key has access to.",
			},
		},
		"Measurements": {
			{
				Name:        auth.READ_MEASUREMENTS.String(),
				Description: "Allows the API key to read measurements that are stored for the tenant this API key has access to.",
			},
			{
				Name:        auth.WRITE_MEASUREMENTS.String(),
				Description: "Allows the API key to write measurements for the tenant this API key has access to.",
			},
		},
		"Tracing": {
			{
				Name:        auth.READ_TRACING.String(),
				Description: "Allows the API key to read tracing messages which give information about the progress of measurements in SensorBucket from receiving them to storage",
			},
		},
		"User workers": {
			{
				Name:        auth.READ_USER_WORKERS.String(),
				Description: "Allows the API key to read user worker code for this tenant",
			},
			{
				Name:        auth.WRITE_USER_WORKERS.String(),
				Description: "Allows the API key to create custom user worker code for this tenant",
			},
		},
	}
}

func toViewTenants(tenants []api.Tenant) []views.TenantInfo {
	viewTenants := []views.TenantInfo{}
	for _, tenant := range tenants {
		viewTenants = append(viewTenants, views.TenantInfo{
			ID:   int(tenant.Id),
			Name: tenant.Name,
		})
	}
	return viewTenants
}

func getCursor(next string) string {
	if next == "" {
		return ""
	}
	u, err := url.Parse(next)
	if err != nil {
		return ""
	}
	return u.Query().Get("cursor")
}
