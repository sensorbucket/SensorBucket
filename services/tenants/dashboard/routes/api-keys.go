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
	r.Get("/create", h.createApiKeyView())
	r.Post("/create", h.createApiKey())
	r.Delete("/revoke/{api_key_id}", h.revokeApiKey())
}

func (h *ApiKeysPageHandler) apiKeysListPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, resp, err := h.client.ApiKeysApi.ListApiKeys(r.Context()).Execute()
		if err != nil {
			log.Printf("api key list page: %v\n", err)
			layout.SnackbarSomethingWentWrong(w)
			return
		}
		if resp == nil || resp.StatusCode != http.StatusOK {
			layout.SnackbarSomethingWentWrong(w)
			return
		}
		if list == nil {
			layout.SnackbarSomethingWentWrong(w)
			return
		}
		keys := map[views.TenantInfo][]views.ApiKey{}
		for _, key := range list.Data {
			k := views.TenantInfo{
				ID:   int(key.TenantId),
				Name: key.TenantName,
			}
			if _, ok := keys[k]; !ok {
				keys[k] = []views.ApiKey{}
			}
			keys[k] = append(keys[k], views.ApiKey{
				ID:             int(key.Id),
				TenantName:     key.TenantName,
				Name:           key.Name,
				ExpirationDate: key.ExpirationDate,
				Created:        key.Created,
			})
		}

		page := &views.ApiKeysPage{ApiKeys: keys}
		if list.Links.GetNext() != "" {
			page.ApiKeysNextPage = views.U("/api-keys/table?cursor=" + getCursor(list.Links.GetNext()))
		}
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
		parsedTime, err := time.Parse("2006-01-02", expiry)
		if err != nil {
			fmt.Println("CANT PARSE TIME")
			layout.SnackbarSomethingWentWrong(w)
			return
		}
		id, err := strconv.ParseInt(tenantId, 10, 32)
		if err != nil {
			fmt.Println("WRONG INT")
			layout.SnackbarSomethingWentWrong(w)
			return
		}
		var dto api.CreateApiKeyRequest
		dto.SetName(name)
		dto.SetTenantId(id)
		dto.SetExpirationDate(parsedTime)
		res, err := h.client.ApiKeysApi.CreateApiKey(r.Context()).CreateApiKeyRequest(api.CreateApiKeyRequest{}).Execute()
		if res.StatusCode != http.StatusOK {
			fmt.Println("STATUS", res.StatusCode)
			layout.SnackbarSomethingWentWrong(w)
			return
		}

		fmt.Println(name, expiry, tenantId)
		fmt.Println("FORM", r.Form)

		w.Write([]byte("stuff"))
		layout.SnackbarSaveSuccessful(w)
	}
}

func (h *ApiKeysPageHandler) createApiKeyView() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := &views.ApiKeysCreatePage{
			Tenants: []views.TenantInfo{
				{
					ID:   12,
					Name: "stuff",
				},
			},
		}
		if layout.IsHX(r) {
			page.WriteBody(w)
			return
		}
		layout.WriteIndex(w, page)
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
