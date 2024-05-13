package webui

import (
	"embed"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/services/tenants/apikeys"
	"sensorbucket.nl/sensorbucket/services/tenants/sessions"
	"sensorbucket.nl/sensorbucket/services/tenants/tenants"
	"sensorbucket.nl/sensorbucket/services/tenants/transports/webui/routes"
	"sensorbucket.nl/sensorbucket/services/tenants/transports/webui/views"
)

type WebUI struct {
	router chi.Router
}

func New(
	baseURLString,
	jwksURL string,
	tenantsService *tenants.TenantService,
	apiKeys *apikeys.Service,
	userPreferences *sessions.UserPreferenceService,
) (*WebUI, error) {
	ui := &WebUI{
		router: chi.NewRouter(),
	}

	var baseURL *url.URL
	if baseURLString != "" {
		fmt.Printf("WebUI base path set to: %s\n", baseURLString)
		baseURL, _ = url.Parse(baseURLString)
		views.SetBase(baseURL)
	}
	ui.router.Use(middleware.Logger)
	jwks := auth.NewJWKSHttpClient(jwksURL)
	authMW := auth.Authenticate(jwks)
	ui.router.Handle("/static/*", serveStatic())
	ui.router.Mount("/auth", routes.SetupKratosRoutes())
	ui.router.With(authMW).Mount("/api-keys", routes.SetupAPIKeyRoutes(apiKeys, tenantsService))
	ui.router.With(authMW).Mount("/switch", routes.SetupTenantSwitchingRoutes(tenantsService, userPreferences))

	return ui, nil
}

func (ui WebUI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ui.router.ServeHTTP(w, r)
}

//go:embed static/*
var staticFS embed.FS

func serveStatic() http.Handler {
	return http.FileServer(http.FS(staticFS))
}
