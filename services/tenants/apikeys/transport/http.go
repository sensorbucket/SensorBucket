package tenantstransport

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/tenants/tenants"
)

func NewHTTP(svc service, url string) *HTTPTransport {
	t := &HTTPTransport{
		router: chi.NewRouter(),
		svc:    svc,
		url:    url,
	}
	t.setupRoutes(t.router)
	return t
}

type HTTPTransport struct {
	router chi.Router
	svc    service
	url    string
}

func (t *HTTPTransport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.router.ServeHTTP(w, r)
}

func (t *HTTPTransport) setupRoutes(r chi.Router) {
	r.Post("/api-keys/new", t.httpCreateApiKey())
	r.Get("/api-keys/validate", t.httpValidateApiKey())
}

func (t *HTTPTransport) httpCreateApiKey() http.HandlerFunc {
	type Params struct {
		TenantID int64      `json:"organisation_id"`
		Expiry   *time.Time `json:"expiry"`
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

		tenant, err := t.svc.GetTenantById(params.TenantID)
		if err != nil {
			if errors.Is(err, tenants.ErrTenantIsNotValid) {
				web.HTTPResponse(w, http.StatusNotFound, web.APIResponseAny{
					Message: "Organisation does not exist or has been archived",
				})
				return
			} else {
				web.HTTPError(w, err)
				return
			}
		}
		apiKey, err := t.svc.GenerateNewApiKey(tenant, params.Expiry)
		if err == nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusCreated, Result{
			ApiKey: apiKey,
		})
	}
}

func (t *HTTPTransport) httpValidateApiKey() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "Authorization header must be set",
			})
			return
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 {
			error
		}

		// The Authorization header should contain the API key in format: 'Bearer <id>:<apiKey>'
		regexPattern := `Bearer (\d+):(\w+)`
		KeyFromString(headerParts[1])
		re := regexp.MustCompile(regexPattern)
		matches := re.FindStringSubmatch(authHeader)
		if len(matches) != 3 {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "Authorization header must have format 'Bearer id:key'",
			})
			return
		}
		apiKey := matches[2]
		if apiKey == "" {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "API key cannot be empty",
			})
			return
		}
		apiKeyId, err := strconv.ParseInt(matches[1], 10, 32)
		if err != nil {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "id in Authorization header should be a valid int",
			})
			return
		}

		// The hashed API key should be retrieved by this id to compare against the given API key by the client
		hashed, err := t.svc.GetHashedApiKeyById(apiKeyId)
		if err != nil {
			if errors.Is(err, tenants.ErrKeyNotFound) {
				// Just send back a unauthorized since we don't want to expose whether an API key id exists or not
				web.HTTPResponse(w, http.StatusUnauthorized, web.APIResponseAny{})
				return
			} else {
				web.HTTPError(w, err)
				return
			}
		}
		authorized := hashed.Compare(apiKey)
		if authorized {
			// API key is succesfully compared and is valid
			web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
				Message: "API Key is valid",
				Data:    "", // TODO: rw.Write([]byte(fmt.Sprintf(`{"sub": "org id"}`)))??
			})
			return
		}
		web.HTTPResponse(w, http.StatusUnauthorized, web.APIResponseAny{})
	}
}

type service interface {
	GetTenantById(id int64) (tenants.Tenant, error)
	GetHashedApiKeyById(id int64) (tenants.ApiKey, error)
	GenerateNewApiKey(owner tenants.Tenant, expiry *time.Time) (string, error)
}
