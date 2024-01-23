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

	layout_utils "sensorbucket.nl/sensorbucket/internal/layout-utils"
	"sensorbucket.nl/sensorbucket/pkg/api"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/pkg/layout"
	"sensorbucket.nl/sensorbucket/services/tenants/transports/webui/views"
)

type ctxKey int

const (
	ctxAPIKeyKey ctxKey = iota
	ctxFlashMessagesKey
)

type APIKeysPageHandler struct {
	router chi.Router
	client *api.APIClient
}

func SetupAPIKeyRoutes(client *api.APIClient) *APIKeysPageHandler {
	handler := &APIKeysPageHandler{
		router: chi.NewRouter(),
		client: client,
	}
	handler.router.With(layout_utils.ExtractFlashMessage).Get("/", handler.apiKeysListPage())
	handler.router.With(layout_utils.ExtractFlashMessage).Get("/create", handler.createAPIKeyView())
	handler.router.Delete("/revoke/{api_key_id}", handler.revokeAPIKey())
	handler.router.Get("/table", handler.apiKeysGetTableRows())
	handler.router.Post("/create", handler.createAPIKey())
	return handler
}

func (h APIKeysPageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
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
				layout_utils.WriteErrorFlashMessage(w, r, "tenant_id must be a valid number", views.RenderFlashMessage)
				return
			}
			req = req.TenantId(id)
		}
		res, _, err := req.Execute()
		if err != nil {
			if apiErr, ok := layout_utils.IsAPIError(err); ok && apiErr.Message != nil {
				log.Printf("list api key for table result api unexpected status code: %s\n", err)
				layout_utils.WriteErrorFlashMessage(w, r, *apiErr.Message, views.RenderFlashMessage)
			} else {
				log.Printf("list api key for table result api error: %s\n", err)
				layout_utils.WriteErrorFlashMessage(w, r, "An unexpected error occurred in the API", views.RenderFlashMessage)
			}
			return
		}

		nextPage := ""
		if res.Links.GetNext() != "" {
			nextPage = views.U("/api-keys/table?cursor=" + getCursor(res.Links.GetNext()))
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
		// Create the page and ensure the page is written in any circumstance
		page := &views.APIKeysPage{}
		defer func(p *views.APIKeysPage) {
			// Check for any flash messages that are in the context
			page.FlashMessages, _ = layout_utils.FlashMessagesFromContext(r.Context())
			if layout.IsHX(r) {
				page.WriteBody(w)
				return
			}
			views.WriteWideLayout(w, page)
		}(page)

		// The initial page starts with an overview of different tenants
		req := h.client.TenantsApi.ListTenants(r.Context())
		req = req.State(1) // State Active
		list, resp, err := req.Execute()
		if err != nil {
			if apiErr, ok := layout_utils.IsAPIError(err); ok && apiErr.Message != nil {
				log.Printf("list api key result api unexpected status code: %s\n", err)
				layout_utils.AddErrorFlashMessage(w, r, *apiErr.Message)
			} else {
				log.Printf("list api key result api error: %s\n", err)
				layout_utils.AddErrorFlashMessage(w, r, "An unexpected error occurred in the API")
			}
			http.Redirect(w, r, "/tenants/api-keys", http.StatusSeeOther)
			return
		}
		if resp.StatusCode != http.StatusOK {
			log.Printf("api key api returned unexpected status code\n")
			layout_utils.AddErrorFlashMessage(w, r, "An unexpected error occurred in the API")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Convert the tenants so they can be displayed in the view
		page.Tenants = toViewTenants(list.Data)
	}
}

func (h *APIKeysPageHandler) revokeAPIKey() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKeyId := chi.URLParam(r, "api_key_id")
		if apiKeyId == "" {
			layout_utils.AddErrorFlashMessage(w, r, "api_key_id must be given")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id, err := strconv.ParseInt(apiKeyId, 10, 64)
		if err != nil {
			layout_utils.AddErrorFlashMessage(w, r, "api_key_id must be a number")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		req := h.client.ApiKeysApi.RevokeApiKey(context.Background(), id)
		resp, err := req.Execute()
		if err != nil {
			if apiErr, ok := layout_utils.IsAPIError(err); ok && apiErr.Message != nil {
				log.Printf("revoke api key result api unexpected status code: %s\n", err)
				layout_utils.AddErrorFlashMessage(w, r, *apiErr.Message)
			} else {
				log.Printf("revoke api key result api error: %s\n", err)
				layout_utils.AddErrorFlashMessage(w, r, "An unexpected error occurred in the API")
			}
			http.Redirect(w, r, "/tenants/api-keys", http.StatusSeeOther)
			return
		}
		if resp.StatusCode != http.StatusOK {
			log.Printf("revoke api key result unexpected status code: %d\n", resp.StatusCode)
			layout_utils.WriteErrorFlashMessage(w, r, "An unexpected error occurred", views.RenderFlashMessage)
			w.WriteHeader(resp.StatusCode)
			return
		}

		// Success, set the flash message and redirect
		w.Header().Set("HX-Redirect", "/tenants/api-keys")
		layout_utils.AddSuccessFlashMessage(w, r, "Succesfully deleted API key")
	}
}

func (h *APIKeysPageHandler) createAPIKey() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			layout_utils.AddErrorFlashMessage(w, r, "Form invalid")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		name := r.FormValue("api-key-name")
		expiry := r.FormValue("api-key-expiry")
		tenantId := r.FormValue("api-key-tenant")
		permissions, ok := r.Form["api-key-permissions"]
		if name == "" || tenantId == "" || !ok || len(permissions) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id, err := strconv.ParseInt(tenantId, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var dto api.CreateApiKeyRequest
		dto.SetName(name)
		dto.SetTenantId(id)

		// Expiry is an optional value
		if expiry != "" {
			parsedTime, err := time.Parse("2006-01-02", expiry)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			dto.SetExpirationDate(parsedTime)
		}
		apiKey, resp, err := h.client.ApiKeysApi.CreateApiKey(r.Context()).CreateApiKeyRequest(dto).Execute()
		if err != nil {
			if apiErr, ok := layout_utils.IsAPIError(err); ok && apiErr.Message != nil {
				log.Printf("create api key result api unexpected status code: %s\n", err)
				layout_utils.AddErrorFlashMessage(w, r, *apiErr.Message)
			} else {
				log.Printf("create api key result api error: %s\n", err)
				layout_utils.AddErrorFlashMessage(w, r, "An unexpected error occurred in the API")
			}
			http.Redirect(w, r, "/tenants/api-keys", http.StatusSeeOther)
			return
		}
		if resp.StatusCode != http.StatusCreated {
			log.Printf("revoke api key result unexpected status code: %d\n", resp.StatusCode)
			layout_utils.AddErrorFlashMessage(w, r, "An unexpected error occurred")
			http.Redirect(w, r, "/tenants/api-keys", http.StatusSeeOther)
			return
		}

		// Add the API key as a flash message
		layout_utils.AddWarningFlashMessage(w, r,
			"This is your API key. Please copy your API key immediately as it will not be shown again.",
			apiKey.ApiKey,
			true)

		// Redirect to overview page so that it may be shown
		http.Redirect(w, r, "/tenants/api-keys", http.StatusSeeOther)
	}
}

func (h *APIKeysPageHandler) createAPIKeyView() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the tenants for the select box in the create view
		req := h.client.TenantsApi.ListTenants(r.Context())
		req = req.State(1) // State Active
		list, resp, err := req.Execute()
		if err != nil {
			if apiErr, ok := layout_utils.IsAPIError(err); ok && apiErr.Message != nil {
				log.Printf("list api key result for create view api unexpected status code: %s\n", err)
				layout_utils.AddErrorFlashMessage(w, r, *apiErr.Message)
			} else {
				log.Printf("list api key result for create view api error: %s\n", err)
				layout_utils.AddErrorFlashMessage(w, r, "An unexpected error occurred in the API")
			}
			http.Redirect(w, r, "/tenants/api-keys", http.StatusSeeOther)
			return
		}
		if resp.StatusCode != http.StatusOK {
			log.Printf("api key list failed : %d\n", resp.StatusCode)
			layout_utils.AddErrorFlashMessage(w, r, "An unexpected error occurred in the API")
			http.Redirect(w, r, "/tenants/api-keys", http.StatusSeeOther)
			return
		}
		perms, err := toViewPermissions()
		if err != nil {
			log.Printf("error creating api key view, converting to view permissions: %s\n", err)
			layout_utils.AddErrorFlashMessage(w, r, "An unexpected error occurred in the API")
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
		views.WriteWideLayout(w, page)
	}
}

func toViewPermissions() (map[views.OrderedMapKey][]views.APIKeysPermission, error) {
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
		categorized[views.OrderedMapKey{Index: len(categorized), Value: " Other"}] = lo.Map(missingInView, func(val string, index int) views.APIKeysPermission {
			return views.APIKeysPermission{
				Name:        val,
				Description: "-",
			}
		})
	}
	return categorized, nil
}

func createAPIKeyViewPermissions() map[views.OrderedMapKey][]views.APIKeysPermission {
	// Golang does not ensure iteration order when iterating over a map, therefore use a simple key struct
	// so we can derive the order in which the items need to be displayed
	return map[views.OrderedMapKey][]views.APIKeysPermission{
		{Index: 0, Value: "Devices"}: {
			{
				Name:        auth.READ_DEVICES.String(),
				Description: "Allows the API key to read information regarding devices of the selected organisation.",
			},
			{
				Name:        auth.WRITE_DEVICES.String(),
				Description: "Allows the API key to write information regarding devices of the selected organisation.",
			},
		},
		{Index: 1, Value: "API Keys"}: {
			{
				Name:        auth.READ_API_KEYS.String(),
				Description: "Does not allow reading of actual API keys. Only allowes the API key to read certain information, for example, the expiration date.",
			},
			{
				Name:        auth.WRITE_API_KEYS.String(),
				Description: "Allows the API key to create other API keys for the tenant the API key has access to.",
			},
		},
		{Index: 2, Value: "Organisations"}: {
			{
				Name:        auth.READ_TENANTS.String(),
				Description: "Allows the API key to read information about the tenant they have access to.",
			},
			{
				Name:        auth.WRITE_TENANTS.String(),
				Description: "Allows the API key to create child organisations for the tenant this API key has access to.",
			},
		},
		{Index: 3, Value: "Measurements"}: {
			{
				Name:        auth.READ_MEASUREMENTS.String(),
				Description: "Allows the API key to read measurements that are stored for the tenant this API key has access to.",
			},
			{
				Name:        auth.WRITE_MEASUREMENTS.String(),
				Description: "Allows the API key to write measurements for the tenant this API key has access to.",
			},
		},
		{Index: 4, Value: "Tracing"}: {
			{
				Name:        auth.READ_TRACING.String(),
				Description: "Allows the API key to read tracing messages which give information about the progress of measurements in SensorBucket from receiving them to storage",
			},
		},
		{Index: 5, Value: "User workers"}: {
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

func customError(msg string) layout_utils.FlashMessage {
	return layout_utils.FlashMessage{
		Title:       "Error",
		Description: msg,
		MessageType: layout_utils.Error,
		CopyButton:  false,
	}
}

func genericError() layout_utils.FlashMessage {
	return layout_utils.FlashMessage{
		Title:       "Error",
		Description: "An unexpected error occurred, please try again or contact a system administrator",
		MessageType: layout_utils.Error,
		CopyButton:  false,
	}
}
