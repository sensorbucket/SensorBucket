package tenantstransports

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/tenants/sessions"
	"sensorbucket.nl/sensorbucket/services/tenants/tenants"
)

type OathkeeperEndpoint struct {
	router        chi.Router
	tenantService *tenants.TenantService
	users         *sessions.UserPreferenceService
}

func NewOathkeeperEndpoint(users *sessions.UserPreferenceService, tenant *tenants.TenantService) *OathkeeperEndpoint {
	ep := &OathkeeperEndpoint{
		router:        chi.NewRouter(),
		tenantService: tenant,
		users:         users,
	}
	ep.setupRoutes()
	return ep
}

func (ep OathkeeperEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ep.router.ServeHTTP(w, r)
}

type AuthenticationSession struct {
	Subject      string         `json:"subject"`
	Extra        map[string]any `json:"extra"`
	Header       any            `json:"header"`
	MatchContext any            `json:"match_context"`
}

func (ep *OathkeeperEndpoint) setupRoutes() {
	r := ep.router
	r.Post("/hydrate", func(w http.ResponseWriter, r *http.Request) {
		var session AuthenticationSession
		if err := web.DecodeJSON(r, &session); err != nil {
			web.HTTPError(w, err)
			return
		}
		defer web.HTTPResponse(w, http.StatusOK, session)

		if session.Subject == "" {
			return
		}
		if session.Extra == nil {
			session.Extra = map[string]any{}
		}
		session.Extra["tid"] = 0
		session.Extra["perms"] = []string{}

		tID, err := ep.users.ActiveTenantID(r.Context(), session.Subject)
		if err != nil {
			fmt.Printf("Hydration error getting active tenant: %s\n", err)
			return
		}
		permissions, err := ep.tenantService.GetMemberPermissions(r.Context(), tID, session.Subject)
		if err != nil {
			fmt.Printf("Hydration error getting member permissions: %s\n", err)
			return
		}

		session.Extra["tid"] = tID
		session.Extra["perms"] = permissions
	})
}
