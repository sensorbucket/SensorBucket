package webui

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"sensorbucket.nl/sensorbucket/pkg/api"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/services/tenants/tenants"
	"sensorbucket.nl/sensorbucket/services/tenants/transports/webui/routes"
	"sensorbucket.nl/sensorbucket/services/tenants/transports/webui/views"
)

type WebUI struct {
	router chi.Router
}

func New(baseURLString, sensorbucketAPIEndpoint string, tenantsService *tenants.TenantService) (*WebUI, error) {
	ui := &WebUI{
		router: chi.NewRouter(),
	}

	var baseURL *url.URL
	if baseURLString != "" {
		fmt.Printf("WebUI base path set to: %s\n", baseURLString)
		baseURL, _ = url.Parse(baseURLString)
		views.SetBase(baseURL)
	}

	sbURL, err := url.Parse(sensorbucketAPIEndpoint)
	if err != nil {
		return nil, fmt.Errorf("could not parse SB_API url: %w", err)
	}
	cfg := api.NewConfiguration()
	cfg.Scheme = sbURL.Scheme
	cfg.Host = sbURL.Host
	client := api.NewAPIClient(cfg)

	ui.router.Use(middleware.Logger)
	jwks := auth.NewJWKSHttpClient("http://oathkeeper:4456/.well-known/jwks.json")
	ui.router.Use(auth.Authenticate(jwks))
	// Middleware to pass on basic auth to the client api
	// TODO: This also exists in dashboard/main.go, perhaps make it a package?
	// Also this will become a JWT instead of basic auth!
	ui.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			if ok {
				r = r.WithContext(context.WithValue(
					r.Context(), api.ContextBasicAuth, api.BasicAuth{
						UserName: user,
						Password: pass,
					}))
			}
			next.ServeHTTP(w, r)
		})
	})
	ui.router.Handle("/static/*", serveStatic())
	ui.router.Mount("/auth", routes.SetupKratosRoutes())
	ui.router.Mount("/api-keys", routes.SetupAPIKeyRoutes(client))
	ui.router.Mount("/user", routes.SetupTenantSwitchingRoutes(tenantsService))

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
