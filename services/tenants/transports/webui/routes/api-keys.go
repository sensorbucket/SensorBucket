package routes

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ory/nosurf"
	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/internal/flash_messages"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/pkg/layout"
	"sensorbucket.nl/sensorbucket/services/tenants/apikeys"
	"sensorbucket.nl/sensorbucket/services/tenants/tenants"
	"sensorbucket.nl/sensorbucket/services/tenants/transports/webui/views"
)

type APIKeysPageHandler struct {
	router  chi.Router
	apiKeys *apikeys.Service
	tenants *tenants.TenantService
}

func SetupAPIKeyRoutes(apiKeys *apikeys.Service, tenants *tenants.TenantService) *APIKeysPageHandler {
	handler := &APIKeysPageHandler{
		router:  chi.NewRouter(),
		apiKeys: apiKeys,
		tenants: tenants,
	}
	handler.router.With(flash_messages.ExtractFlashMessage).Get("/", handler.apiKeysListPage())
	handler.router.With(flash_messages.ExtractFlashMessage).Get("/create", handler.createAPIKeyView())
	handler.router.Get("/table", handler.apiKeysGetTableRows())
	handler.router.Post("/create", handler.createAPIKey())

	handler.router.Route("/revoke/{api_key_id}", func(r chi.Router) {
		r.Get("/", handler.revokeAPIKey())
		r.Delete("/", handler.revokeAPIKey())
		// Post also here since html forms can only GET/POST...
		r.Post("/", handler.revokeAPIKey())
	})
	return handler
}

func (h APIKeysPageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *APIKeysPageHandler) apiKeysGetTableRows() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var tenantIDS []int64
		query := r.URL.Query()
		for _, tenantIDString := range query["tenant_id"] {
			tenantID, err := strconv.ParseInt(tenantIDString, 10, 64)
			if err != nil {
				flash_messages.WriteErrorFlashMessage(w, r, "tenant_id must be a valid number")
				return
			}
			tenantIDS = append(tenantIDS, tenantID)
		}
		apiKeyPage, err := h.apiKeys.ListAPIKeys(
			r.Context(),
			apikeys.Filter{
				TenantID: tenantIDS,
			},
			pagination.Request{
				Cursor: r.URL.Query().Get("cursor"),
				Limit:  15,
			},
		)
		if err != nil {
			if apiErr, ok := flash_messages.IsAPIError(err); ok && apiErr.Message != nil {
				log.Printf("list api key for table result api unexpected status code: %s\n", err)
				flash_messages.WriteErrorFlashMessage(w, r, *apiErr.Message)
			} else {
				log.Printf("list api key for table result api error: %s\n", err)
				flash_messages.WriteErrorFlashMessage(w, r, "An unexpected error occurred in the API")
			}
			return
		}

		nextPage := ""
		if apiKeyPage.Cursor != "" {
			nextPage = views.U("/api-keys/table?cursor=%s", apiKeyPage.Cursor)
		}

		viewKeys := []views.APIKey{}
		for _, key := range apiKeyPage.Data {
			viewKeys = append(viewKeys, views.APIKey{
				ID:             int(key.ID),
				TenantID:       int(key.TenantID),
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
		page := &views.APIKeysPage{
			Base: views.Base{
				CSRFToken: nosurf.Token(r),
			},
		}
		defer func() {
			flash_messages.AddContextFlashMessages(r, &page.FlashMessagesContainer)

			// Check for any flash messages that are in the context
			if layout.IsHX(r) {
				page.WriteBody(w)
				return
			}
			views.WriteWideLayout(w, page)
		}()

		// The initial page starts with an overview of different tenants
		tenantPage, err := h.tenants.ListTenants(
			r.Context(),
			tenants.Filter{
				State: []tenants.State{1},
			},
			pagination.Request{},
		)
		if err != nil {
			if apiErr, ok := flash_messages.IsAPIError(err); ok && apiErr.Message != nil {
				log.Printf("list api key result api unexpected status code: %s\n", err)
				flash_messages.AddErrorFlashMessageToPage(r, &page.FlashMessagesContainer, *apiErr.Message)
			} else {
				log.Printf("list api key result api error: %s\n", err)
				flash_messages.AddErrorFlashMessageToPage(r, &page.FlashMessagesContainer, "An unexpected error occurred in the API")
			}
			return
		}

		// Convert the tenants so they can be displayed in the view
		page.Tenants = toViewTenants(tenantPage.Data)
	}
}

func (h *APIKeysPageHandler) revokeAPIKey() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKeyIDStr := chi.URLParam(r, "api_key_id")
		if apiKeyIDStr == "" {
			flash_messages.AddErrorFlashMessage(w, r, "api_key_id must be given")
			w.WriteHeader(http.StatusBadRequest)
			http.Redirect(w, r, views.U("/api-keys"), http.StatusSeeOther)
			return
		}
		apiKeyID, err := strconv.ParseInt(apiKeyIDStr, 10, 64)
		if err != nil {
			flash_messages.AddErrorFlashMessage(w, r, "api_key_id must be a number")
			w.WriteHeader(http.StatusBadRequest)
			http.Redirect(w, r, views.U("/api-keys"), http.StatusSeeOther)
			return
		}

		if r.Method == http.MethodGet {
			key, err := h.apiKeys.GetAPIKey(r.Context(), apiKeyID)
			if err != nil {
				flash_messages.AddErrorFlashMessage(w, r, "Could not get requested API Key")
				w.Header().Set("HX-Redirect", "/tenants/api-keys")
				http.Redirect(w, r, views.U("/api-keys"), http.StatusSeeOther)
				return
			}
			tenant, err := h.tenants.GetTenantByID(r.Context(), key.TenantID)
			if err != nil {
				flash_messages.AddErrorFlashMessage(w, r, "Could not get requested API Key")
				w.Header().Set("HX-Redirect", "/tenants/api-keys")
				http.Redirect(w, r, views.U("/api-keys"), http.StatusSeeOther)
				return
			}
			views.WriteLayout(w, &views.APIKeyDeletePage{
				Base: views.Base{
					CSRFToken: nosurf.Token(r),
				},
				KeyName:   key.Name,
				KeyTenant: tenant.Name,
			})
			return
		}

		err = h.apiKeys.RevokeApiKey(r.Context(), apiKeyID)
		if err != nil {
			if apiErr, ok := flash_messages.IsAPIError(err); ok && apiErr.Message != nil {
				log.Printf("revoke api key result api unexpected status code: %s\n", err)
				flash_messages.AddErrorFlashMessage(w, r, *apiErr.Message)
			} else {
				log.Printf("revoke api key result api error: %s\n", err)
				flash_messages.AddErrorFlashMessage(w, r, "An unexpected error occurred in the API")
			}
			w.Header().Set("HX-Redirect", "/tenants/api-keys")
			http.Redirect(w, r, views.U("/api-keys"), http.StatusSeeOther)
			return
		}

		// Success, set the flash message and redirect
		w.Header().Set("HX-Redirect", "/tenants/api-keys")
		flash_messages.AddSuccessFlashMessage(w, r, "Succesfully deleted API key")
		http.Redirect(w, r, views.U("/api-keys"), http.StatusSeeOther)
	}
}

func (h *APIKeysPageHandler) createAPIKey() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			flash_messages.AddErrorFlashMessage(w, r, "Form invalid")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		name := r.FormValue("api-key-name")
		expiryString := r.FormValue("api-key-expiry")
		tenantId := r.FormValue("api-key-tenant")
		permissionStrings, ok := r.Form["api-key-permissions"]
		if name == "" || tenantId == "" || !ok || len(permissionStrings) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id, err := strconv.ParseInt(tenantId, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Expiry is an optional value
		var expiry *time.Time
		if expiryString != "" {
			parsedTime, err := time.Parse("2006-01-02", expiryString)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			expiry = &parsedTime
		}

		// Validate permissions
		permissions, err := auth.StringsToPermissions(permissionStrings)
		if err != nil {
			log.Printf("create api key, invalid permission: %s\n", err)
			flash_messages.AddErrorFlashMessage(w, r, err.Error())
			return
		}

		apiKey, err := h.apiKeys.GenerateNewApiKey(r.Context(), name, id, permissions, expiry)
		if err != nil {
			if apiErr, ok := flash_messages.IsAPIError(err); ok && apiErr.Message != nil {
				log.Printf("create api key result api unexpected status code: %s\n", err)
				flash_messages.AddErrorFlashMessage(w, r, *apiErr.Message)
			} else {
				log.Printf("create api key result api error: %s\n", err)
				flash_messages.AddErrorFlashMessage(w, r, "An unexpected error occurred in the API")
			}
			http.Redirect(w, r, "/tenants/api-keys", http.StatusSeeOther)
			return
		}

		// Add the API key as a flash message
		flash_messages.AddWarningFlashMessage(w, r,
			"This is your API key. Please copy your API key immediately as it will not be shown again.",
			apiKey,
			true)

		// Redirect to overview page so that it may be shown
		http.Redirect(w, r, "/tenants/api-keys", http.StatusSeeOther)
	}
}

func (h *APIKeysPageHandler) createAPIKeyView() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the tenants for the select box in the create view
		tenantsPage, err := h.tenants.ListTenants(r.Context(), tenants.Filter{State: []tenants.State{tenants.Active}}, pagination.Request{})
		if err != nil {
			if apiErr, ok := flash_messages.IsAPIError(err); ok && apiErr.Message != nil {
				log.Printf("list api key result for create view api unexpected status code: %s\n", err)
				flash_messages.AddErrorFlashMessage(w, r, *apiErr.Message)
			} else {
				log.Printf("list api key result for create view api error: %s\n", err)
				flash_messages.AddErrorFlashMessage(w, r, "An unexpected error occurred in the API")
			}
			http.Redirect(w, r, "/tenants/api-keys", http.StatusSeeOther)
			return
		}

		perms, err := toViewPermissions()
		if err != nil {
			log.Printf("error creating api key view, converting to view permissions: %s\n", err)
			flash_messages.AddErrorFlashMessage(w, r, "An unexpected error occurred in the API")
			return
		}
		page := &views.APIKeysCreatePage{
			Base: views.Base{
				CSRFToken: nosurf.Token(r),
			},
			Tenants:     toViewTenants(tenantsPage.Data),
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
	allPermissions := auth.AllPermissions()
	authAsStrSlice := []string{}
	for _, p := range allPermissions {
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

func toViewTenants(tenants []tenants.CreateTenantDTO) []views.TenantInfo {
	viewTenants := []views.TenantInfo{}
	for _, tenant := range tenants {
		viewTenants = append(viewTenants, views.TenantInfo{
			ID:   int(tenant.ID),
			Name: tenant.Name,
		})
	}
	return viewTenants
}
