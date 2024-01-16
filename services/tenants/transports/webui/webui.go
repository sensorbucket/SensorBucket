package webui

import (
	"embed"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"sensorbucket.nl/sensorbucket/services/tenants/transports/webui/routes"
	"sensorbucket.nl/sensorbucket/services/tenants/transports/webui/views"
)

type WebUI struct {
	router chi.Router
}

func New(baseURLString string) *WebUI {
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
	ui.router.Handle("/static/*", serveStatic())
	ui.router.Mount("/auth", routes.SetupKratosRoutes())

	return ui
}

func (ui WebUI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ui.router.ServeHTTP(w, r)
}

//go:embed static/*
var staticFS embed.FS

func serveStatic() http.Handler {
	return http.FileServer(http.FS(staticFS))
}
