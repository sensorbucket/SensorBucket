package tenantstransports

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"sensorbucket.nl/sensorbucket/internal/httpfilter"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/tenants/apikeys"
)

func NewAPIKeysHTTP(r chi.Router, apiKeySvc apiKeyService, url string) *APIKeysHTTPTransport {
	t := &APIKeysHTTPTransport{
		router:    r,
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
	r.Get("/api-keys/list", t.httpListApiKeys())
	r.Delete("/api-keys/{api_key_id}", t.httpRevokeApiKey())
	r.Post("/api-keys", t.httpCreateApiKey())
	r.Get("/api-keys/authenticate", t.httpAuthenticateApiKey())
}

func (t *APIKeysHTTPTransport) httpRevokeApiKey() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKeyIdStr := chi.URLParam(r, "api_key_id")
		if apiKeyIdStr == "" {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "api_key_id cannot be empty",
			})
			return
		}
		apiKeyId, err := strconv.ParseInt(apiKeyIdStr, 10, 64)
		if err != nil {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "api_key_id must be a valid int",
			})
			return
		}
		if err = t.apiKeySvc.RevokeApiKey(apiKeyId); err != nil {
			if errors.Is(err, apikeys.ErrKeyNotFound) {
				web.HTTPResponse(w, http.StatusNotFound, web.APIResponseAny{
					Message: "Key does not exist",
				})
				return
			} else {
				web.HTTPError(w, err)
				return
			}
		}
		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
			Message: "API key has been revoked",
		})
	}
}

func (t *APIKeysHTTPTransport) httpCreateApiKey() http.HandlerFunc {
	type Params struct {
		Name           string     `json:"name"`
		TenantID       int64      `json:"tenant_id"`
		ExpirationDate *time.Time `json:"expiration_date"`
	}
	type Result struct {
		ApiKey string `json:"api_key"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		params := Params{}
		defer r.Body.Close()
		if err := web.DecodeJSON(r, &params); err != nil {
			web.HTTPError(w, err)
			return
		}
		if params.Name == "" {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "name cannot be empty",
			})
			return
		}

		if params.TenantID <= 0 {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "tenant_id must be higher than 0",
			})
			return
		}

		if params.ExpirationDate != nil && !params.ExpirationDate.After(time.Now()) {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "expiration_date must be set in the future",
			})
			return
		}
		apiKey, err := t.apiKeySvc.GenerateNewApiKey(params.Name, params.TenantID, params.ExpirationDate)
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

func (t *APIKeysHTTPTransport) httpAuthenticateApiKey() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "Authorization header must be set",
			})
			return
		}
		idAndKeyCombination := strings.TrimPrefix(authHeader, "Bearer ")
		validResp, err := t.apiKeySvc.AuthenticateApiKey(idAndKeyCombination)
		if err == nil {
			web.HTTPResponse(w, http.StatusOK, validResp)
			return
		} else if errors.Is(err, apikeys.ErrInvalidEncoding) {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "Invalid input",
			})
			return
		} else if errors.Is(err, apikeys.ErrKeyNotFound) {
			web.HTTPResponse(w, http.StatusUnauthorized, web.APIResponseAny{})
			return
		} else {
			web.HTTPError(w, err)
			return
		}
	}
}

func (t *APIKeysHTTPTransport) httpListApiKeys() http.HandlerFunc {
	type Params struct {
		apikeys.Filter     `pagination:",squash"`
		pagination.Request `pagination:",squash"`
	}
	return func(rw http.ResponseWriter, r *http.Request) {
		params, err := httpfilter.Parse[Params](r)
		if err != nil {
			web.HTTPError(rw, web.NewError(http.StatusBadRequest, "invalid params", ""))
			return
		}
		page, err := t.apiKeySvc.ListAPIKeys(params.Filter, params.Request)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, pagination.CreateResponse(r, t.url, *page))
	}
}

type apiKeyService interface {
	AuthenticateApiKey(base64IdAndKeyCombination string) (apikeys.ApiKeyAuthenticationDTO, error)
	GenerateNewApiKey(name string, tenantId int64, expiry *time.Time) (string, error)
	RevokeApiKey(id int64) error
	ListAPIKeys(filter apikeys.Filter, p pagination.Request) (*pagination.Page[apikeys.ApiKeyDTO], error)
}
