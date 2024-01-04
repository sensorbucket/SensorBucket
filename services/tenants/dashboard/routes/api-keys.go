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
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/api"
	"sensorbucket.nl/sensorbucket/pkg/layout"
	"sensorbucket.nl/sensorbucket/services/tenants/dashboard/views"
)

type ApiKeysPageHandler struct {
	router chi.Router
	client *api.APIClient
}

func CreateApiKeysPageHandler(client *api.APIClient) *ApiKeysPageHandler {
	handler := &ApiKeysPageHandler{
		router: chi.NewRouter(),
		client: client,
	}
	handler.SetupRoutes(handler.router)
	return handler
}

func (h ApiKeysPageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *ApiKeysPageHandler) SetupRoutes(r chi.Router) {
	r.Get("/", h.apiKeysListPage())
	r.Get("/table", h.apiKeysGetTableRows())
	r.Get("/create", h.createApiKeyView())
	r.Post("/create", h.createApiKey())
	r.Delete("/revoke/{api_key_id}", h.revokeApiKey())
}

func (h *ApiKeysPageHandler) apiKeysGetTableRows() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := h.client.ApiKeysApi.ListApiKeys(r.Context())
		if r.URL.Query().Has("cursor") {
			req = req.Cursor(r.URL.Query().Get("cursor"))
		} else {
			fmt.Println("no qyer")
			// If no cursor is given, the initial state must be derived from the url params
			req = req.Limit(15)
			tenantId := r.URL.Query().Get("tenant_id")
			id, err := strconv.ParseInt(tenantId, 10, 32)
			if err != nil {
				layout.SnackbarSomethingWentWrong(w)
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
			nextPage = views.U("/api-keys/table?cursor=" + getCursor(res.Links.GetNext()))
		}

		fmt.Println("retrieved keys", len(res.Data))
		viewKeys := []views.ApiKey{}
		for _, key := range res.Data {
			viewKeys = append(viewKeys, views.ApiKey{
				ID:       int(key.Id),
				TenantID: int(key.TenantId),
				Name:     key.Name,
			})
		}

		views.WriteRenderApiKeyRows(w, viewKeys, nextPage)
	}
}

func (h *ApiKeysPageHandler) apiKeysListPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// The initial page starts with an overview of different tenants
		// req := h.client.ApiKeysApi.ListApiKeys(r.Context())
		// list, resp, err := req.Execute()
		// if err != nil {
		// 	log.Printf("api key list page: %v\n", err)
		// 	layout.SnackbarSomethingWentWrong(w)
		// 	return
		// }
		// if resp == nil || resp.StatusCode != http.StatusOK {
		// 	layout.SnackbarSomethingWentWrong(w)
		// 	return
		// }
		// if list == nil {
		// 	layout.SnackbarSomethingWentWrong(w)
		// 	return
		// }
		// keys := apiKeysGrouped(list.Data)
		page := &views.ApiKeysPage{Tenants: tenants()}
		// if list.Links.GetNext() != "" {
		// 	page.ApiKeysNextPage = views.U("/api-keys/table?cursor=" + getCursor(list.Links.GetNext()))
		// }
		if layout.IsHX(r) {
			page.WriteBody(w)
			return
		}
		layout.WriteIndex(w, page)
	}
}

func (h *ApiKeysPageHandler) revokeApiKey() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKeyId := chi.URLParam(r, "api_key_id")
		if apiKeyId == "" {
			layout.WithSnackbarError(w, "api_key_id must be given", http.StatusBadRequest)
			return
		}
		id, err := strconv.ParseInt(apiKeyId, 10, 64)
		if err != nil {
			layout.WithSnackbarError(w, "api_key_id must be a number", http.StatusBadRequest)
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

func (h *ApiKeysPageHandler) createApiKey() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			layout.SnackbarSomethingWentWrong(w)
			return
		}
		name := r.FormValue("api-key-name")
		expiry := r.FormValue("api-key-expiry")
		tenantId := r.FormValue("api-key-tenant")
		id, err := strconv.ParseInt(tenantId, 10, 32)
		if err != nil {
			layout.SnackbarSomethingWentWrong(w)
			return
		}

		var dto api.CreateApiKeyRequest
		dto.SetName(name)
		dto.SetTenantId(id)

		// Expiry is an optional value
		if expiry != "" {
			parsedTime, err := time.Parse("2006-01-02", expiry)
			if err != nil {
				layout.SnackbarSomethingWentWrong(w)
				return
			}

			dto.SetExpirationDate(parsedTime)
		}
		apiKey, res, err := h.client.ApiKeysApi.CreateApiKey(r.Context()).CreateApiKeyRequest(dto).Execute()
		if res.StatusCode != http.StatusCreated {
			layout.SnackbarSomethingWentWrong(w)
			return
		}

		// Show the API key to the user
		layout.WithSnackbarSuccess(w, "Created API Key")
		w.Write([]byte(apiKey.ApiKey))

	}
}

func (h *ApiKeysPageHandler) createApiKeyView() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := &views.ApiKeysCreatePage{
			// TODO: when tenants ticket is merged
			Tenants: tenants(),
		}
		if layout.IsHX(r) {
			page.WriteBody(w)
			return
		}
		layout.WriteIndex(w, page)
	}
}

func tenants() []views.TenantInfo {
	return []views.TenantInfo{
		{
			ID:   12,
			Name: "Pollex B.V.",
		},
		{
			ID:   5,
			Name: "Provincie Zeeland",
		},
		{
			ID:   345,
			Name: "Aannemer 1",
		},
	}
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
