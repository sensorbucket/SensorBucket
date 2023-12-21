package tenantstransports

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/tenants/apikeys"
)

func NewAPIKeysHTTP(apiKeySvc apiKeyService, url string) *APIKeysHTTPTransport {
	t := &APIKeysHTTPTransport{
		router:    chi.NewRouter(),
		apiKeySvc: apiKeySvc,
		url:       url,
	}
	t.setupRoutes(t.router)
	return t
}

type APIKeysHTTPTransport struct {
	router    chi.Router
	apiKeySvc apiKeyService
	url       string
}

func (t *APIKeysHTTPTransport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.router.ServeHTTP(w, r)
}

func (t *APIKeysHTTPTransport) setupRoutes(r chi.Router) {
	r.Delete("/api-keys/revoke", t.httpRevokeApiKey())
	r.Post("/api-keys/new", t.httpCreateApiKey())
	r.Get("/api-keys/validate", t.httpValidateApiKey())
}

func (t *APIKeysHTTPTransport) httpRevokeApiKey() http.HandlerFunc {
	type Params struct {
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
		if params.ApiKey == "" {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "API key cannot be empty",
			})
			return
		}
		if err = t.apiKeySvc.RevokeApiKey(params.ApiKey); err != nil {
			if errors.Is(err, apikeys.ErrInvalidEncoding) {
				web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
					Message: "Invalid input",
				})
				return
			} else {
				web.HTTPError(w, err)
				return
			}
		}
		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
			Message: "API key has been revoked",
			Data: struct {
				ApiKey string `json:"api_key"`
			}{
				ApiKey: params.ApiKey,
			},
		})
	}
}

func (t *APIKeysHTTPTransport) httpCreateApiKey() http.HandlerFunc {
	type Params struct {
		TenantID       int64      `json:"organisation_id"`
		ExpirationDate *time.Time `json:"expiration_date"`
	}
	type Result struct {
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

		apiKey, err := t.apiKeySvc.GenerateNewApiKey(params.TenantID, params.ExpirationDate)
		if err != nil {
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
			ApiKey: apiKey,
		})
	}
}

func (t *APIKeysHTTPTransport) httpValidateApiKey() http.HandlerFunc {
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
			if errors.Is(err, apikeys.ErrInvalidEncoding) {
				web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
					Message: "Invalid input",
				})
				return
			} else {
				web.HTTPError(w, err)
				return
			}
		}
		if valid {
			web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
				Message: "API Key is valid",
			})
			return
		}
		web.HTTPResponse(w, http.StatusUnauthorized, web.APIResponseAny{})
	}
}

type apiKeyService interface {
	ValidateApiKey(base64IdAndKeyCombination string) (bool, error)
	GenerateNewApiKey(tenantId int64, expiry *time.Time) (string, error)
	RevokeApiKey(base64IdAndKeyCombination string) error
}
