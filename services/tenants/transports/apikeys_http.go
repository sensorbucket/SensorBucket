package tenantstransports

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"sensorbucket.nl/sensorbucket/internal/httpfilter"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/services/tenants/apikeys"
)

func NewAPIKeysHTTP(r chi.Router, apiKeySvc ApiKeyService, url string) *APIKeysHTTPTransport {
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
	apiKeySvc ApiKeyService
	url       string
}

func (t *APIKeysHTTPTransport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.router.ServeHTTP(w, r)
}

func (t *APIKeysHTTPTransport) setupRoutes(r chi.Router) {
	r.Get("/api-keys", t.httpListApiKeys())
	r.Get("/api-keys/{api_key_id}", t.httpGetApiKey())
	r.Delete("/api-keys/{api_key_id}", t.httpRevokeApiKey())
	r.Post("/api-keys", t.httpCreateApiKey())
	r.Handle("/api-keys/authenticate", t.httpAuthenticateApiKey())
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
		if err = t.apiKeySvc.RevokeApiKey(r.Context(), apiKeyId); err != nil {
			log.Printf("error revoking api key: %s\n", err.Error())
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
		Name           string           `json:"name"`
		TenantID       int64            `json:"tenant_id"`
		ExpirationDate *time.Time       `json:"expiration_date"`
		Permissions    auth.Permissions `json:"permissions"`
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

		if len(params.Permissions) == 0 {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "at least one permission is required",
			})
			return
		}

		apiKey, err := t.apiKeySvc.GenerateNewApiKey(
			r.Context(),
			params.Name,
			params.TenantID,
			params.Permissions,
			params.ExpirationDate,
		)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusCreated, Result{
			ApiKey: apiKey,
		})
	}
}

func (t *APIKeysHTTPTransport) httpAuthenticateApiKey() http.HandlerFunc {
	type AuthenticationSession struct {
		Subject      string         `json:"subject,omitempty"`
		Extra        map[string]any `json:"extra,omitempty"`
		Header       any            `json:"header,omitempty"`
		MatchContext any            `json:"match_context,omitempty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := auth.StripBearer(r.Header.Get("Authorization"))
		if !ok {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "Authorization header must be set",
			})
			return
		}
		keyInfo, err := t.apiKeySvc.AuthenticateApiKey(r.Context(), token)
		if err == nil {
			session := AuthenticationSession{
				Subject: "",
				Extra: map[string]any{
					"tid":   keyInfo.TenantID,
					"perms": keyInfo.Permissions,
				},
			}
			if keyInfo.Expiration != nil {
				session.Extra["exp"] = *keyInfo.Expiration
			}
			web.HTTPResponse(w, http.StatusOK, session)
			return
		} else if errors.Is(err, apikeys.ErrInvalidEncoding) {
			fmt.Printf("API authenticate denied: %s\n", err)
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "Invalid input",
			})
			return
		} else if errors.Is(err, apikeys.ErrKeyNotFound) {
			fmt.Printf("API authenticate denied: %s\n", err)
			web.HTTPResponse(w, http.StatusUnauthorized, web.APIResponseAny{})
			return
		} else {
			fmt.Printf("API authenticate denied: %s\n", err)
			web.HTTPError(w, err)
			return
		}
	}
}

func (t *APIKeysHTTPTransport) httpGetApiKey() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		keyIDStr := chi.URLParam(r, "api_key_id")
		keyID, err := strconv.ParseInt(keyIDStr, 10, 64)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}
		key, err := t.apiKeySvc.GetAPIKey(r.Context(), keyID)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, web.APIResponseAny{
			Data: key,
		})
	}
}

func (t *APIKeysHTTPTransport) httpListApiKeys() http.HandlerFunc {
	type Params struct {
		apikeys.APIKeyFilter `pagination:",squash"`
		pagination.Request   `pagination:",squash"`
	}
	return func(rw http.ResponseWriter, r *http.Request) {
		params, err := httpfilter.Parse[Params](r)
		if err != nil {
			web.HTTPError(rw, web.NewError(http.StatusBadRequest, "invalid params", ""))
			return
		}
		page, err := t.apiKeySvc.ListAPIKeys(r.Context(), params.APIKeyFilter, params.Request)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, pagination.CreateResponse(r, t.url, *page))
	}
}

type ApiKeyService interface {
	AuthenticateApiKey(
		ctx context.Context,
		base64IdAndKeyCombination string,
	) (apikeys.ApiKeyAuthenticationDTO, error)
	GenerateNewApiKey(
		ctx context.Context,
		name string,
		tenantId int64,
		permissions auth.Permissions,
		expiry *time.Time,
	) (string, error)
	RevokeApiKey(ctx context.Context, id int64) error
	ListAPIKeys(
		ctx context.Context,
		filter apikeys.APIKeyFilter,
		p pagination.Request,
	) (*pagination.Page[apikeys.ApiKeyDTO], error)
	GetAPIKey(ctx context.Context, id int64) (*apikeys.HashedApiKey, error)
}
