package apikeystransport

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/tenants/apikeys"
)

func NewHTTP(apiKeySvc apiKeyService, url string) *HTTPTransport {
	t := &HTTPTransport{
		router:    chi.NewRouter(),
		apiKeySvc: apiKeySvc,
		url:       url,
	}
	t.setupRoutes(t.router)
	return t
}

type HTTPTransport struct {
	router    chi.Router
	apiKeySvc apiKeyService
	url       string
}

func (t *HTTPTransport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.router.ServeHTTP(w, r)
}

func (t *HTTPTransport) setupRoutes(r chi.Router) {
	r.Delete("/api-keys/revoke/{api_key_id}", t.httpRevokeApiKey())
	r.Post("/api-keys/new", t.httpCreateApiKey())
	r.Get("/api-keys/validate", t.httpValidateApiKey())
}

func (t *HTTPTransport) httpRevokeApiKey() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKeyIdStr := chi.URLParam(r, "api_key_id")
		if apiKeyIdStr == "" {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "api_key_id must be set",
			})
			return
		}
		apiKeyId, err := strconv.ParseInt(apiKeyIdStr, 10, 32)
		if err != nil {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "api_key_id must be a valid int",
			})
			return
		}
		err = t.apiKeySvc.RevokeApiKey(apiKeyId)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
			Message: "API key has been revoked",
			Data: struct {
				ApiKeyId int64 `json:"api_key_id"`
			}{
				ApiKeyId: apiKeyId,
			},
		})
	}
}

func (t *HTTPTransport) httpCreateApiKey() http.HandlerFunc {
	type Params struct {
		TenantID int64      `json:"organisation_id"`
		Expiry   *time.Time `json:"expiry"`
	}
	type Result struct {
		ID     int64  `json:"id"`
		ApiKey string `json:"api_key"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		params := Params{}
		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "Invalid data",
			})
			return
		}

		apiKeyId, apiKey, err := t.apiKeySvc.GenerateNewApiKey(params.TenantID, params.Expiry)
		if err == nil {
			if errors.Is(err, apikeys.ErrTenantIsNotValid) {
				web.HTTPResponse(w, http.StatusNotFound, web.APIResponseAny{
					Message: "Organisation does not exist or has been archived",
				})
				return
			} else {
				web.HTTPError(w, err)
				return
			}
		}
		web.HTTPResponse(w, http.StatusCreated, Result{
			ID:     apiKeyId,
			ApiKey: apiKey,
		})
	}
}

func (t *HTTPTransport) httpValidateApiKey() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "Authorization header must be set",
			})
			return
		}
		idAndKeyCombination := strings.TrimPrefix(authHeader, "Bearer ")
		valid, err := t.apiKeySvc.ValidateApiKey(idAndKeyCombination)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		if valid {
			web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
				Message: "API Key is valid",
				Data:    "", // TODO: rw.Write([]byte(fmt.Sprintf(`{"sub": "org id"}`)))??
			})
			return
		}
		web.HTTPResponse(w, http.StatusUnauthorized, web.APIResponseAny{})
	}
}

type apiKeyService interface {
	ValidateApiKey(base64IdAndKeyCombination string) (bool, error)
	GenerateNewApiKey(tenantId int64, expiry *time.Time) (int64, string, error)
	RevokeApiKey(apiKeyId int64) error
}
