package routes

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/api"
	"sensorbucket.nl/sensorbucket/pkg/layout"
	"sensorbucket.nl/sensorbucket/services/tenants/dashboard/views"
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
			id, err := strconv.ParseInt(tenantId, 10, 32)
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
		if resp == nil || resp.StatusCode != http.StatusOK {
			layout.SnackbarSomethingWentWrong(w)
			return
		}
		page := &views.APIKeysPage{Tenants: toViewTenants(list.Data)}
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

		id, err := strconv.ParseInt(tenantId, 10, 32)
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

		// Show the API key to the user
		w.Write([]byte(apiKey.ApiKey))
		w.Header().Set("HX-Location", `{"path":"/api-keys", "target":""}`)
		// HX-Location: {"path":"/test2", "target":"#testdiv"}
		// w.Header().Set("HX-Redirect", "/api-keys")/
		layout.WithSnackbarSuccess(w, "Created API Key")

		// w.Write([]byte(apiKey.ApiKey))

		page := &views.APIKeysPage{}
		if layout.IsHX(r) {
			page.WriteBody(w)
			return
		}
		layout.WriteIndex(w, page)

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
			log.Printf("api key list failed : %d", resp.StatusCode)
			layout.SnackbarSomethingWentWrong(w)
			return
		}
		page := &views.APIKeysCreatePage{
			Tenants:     toViewTenants(list.Data),
			Permissions: toViewPermissions(),
		}
		if layout.IsHX(r) {
			page.WriteBody(w)
			return
		}
		layout.WriteIndex(w, page)
	}
}

// TODO: should be replaced with actual auth package values once auth package is done
// https://github.com/sensorbucket/SensorBucket/issues/70
func toViewPermissions() map[string][]views.APIKeysCreatePermission {
	return map[string][]views.APIKeysCreatePermission{
		"Devices": {
			{
				Name:        "READ_DEVICES",
				Description: "Allows the API key to read information regarding devices of the selected organisation",
			},
			{
				Name:        "WRITE_DEVICES",
				Description: "Allows the API key to write information regarding devices of the selected organisation",
			},
		},
		"Uplinks": {
			{
				Name:        "WRITE_UPLINKS",
				Description: "Allows the API key create uplinks",
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
