package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"sensorbucket.nl/sensorbucket/pkg/api"
	"sensorbucket.nl/sensorbucket/services/tenants/dashboard/routes"
)

func main() {
	fmt.Println("Starting Tenants Dashboard")
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	sbURL, err := url.Parse("http://localhost:3000/api")
	if err != nil {
		panic(err)
	}
	cfg := api.NewConfiguration()
	cfg.Scheme = sbURL.Scheme
	cfg.Host = sbURL.Host
	apiClient := api.NewAPIClient(cfg)

	staticPath := "/home/jeffrey/projects/pollex/SensorBucket/pkg/layout/static"
	fmt.Println("Serving static files...")
	router.Use(middleware.GetHead)
	fileServer := http.FileServer(http.Dir(staticPath))
	router.Handle("/static/*", http.StripPrefix("/static", fileServer))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		u := "/overview"
		http.Redirect(w, r, u, http.StatusFound)
	})
	router.Mount("/api-keys", routes.CreateApiKeysPageHandler(apiClient))

	srv := &http.Server{
		Addr:         ":3010",
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
		Handler:      router,
	}
	fmt.Printf("HTTP Server listening on: %s\n", srv.Addr)

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		// errC <- err
		panic(err)
	}

}
