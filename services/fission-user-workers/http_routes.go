package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"sensorbucket.nl/sensorbucket/internal/httpfilter"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
)

type HTTPTransport struct {
	server *http.Server
}

func newHTTPTransport(app *Application, addr string) *HTTPTransport {
	r := chi.NewRouter()
	createRoutes(app, r)
	srv := &http.Server{
		Addr:         addr,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
		Handler:      r,
	}
	return &HTTPTransport{
		server: srv,
	}
}

func (t *HTTPTransport) Start() {
	log.Printf("HTTP server listening at: %s\n", HTTP_ADDR)
	t.server.ListenAndServe()
}

func (t *HTTPTransport) Stop(ctx context.Context) {
	log.Println("HTTP Server shutting down...")
	t.server.Shutdown(ctx)
}

func createRoutes(app *Application, r chi.Router) {
	r.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello"))
	})
	r.Get("/workers", func(w http.ResponseWriter, r *http.Request) {
		params, err := httpfilter.Parse[pagination.Request](r)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		page, err := app.ListWorkers(r.Context(), params)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusOK, pagination.CreateResponse(r, "", *page))
	})
	r.Post("/workers", func(w http.ResponseWriter, r *http.Request) {
		var dto CreateWorkerOpts
		if err := web.DecodeJSON(r, &dto); err != nil {
			web.HTTPError(w, err)
			return
		}
		worker, err := app.CreateWorker(r.Context(), dto)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusCreated, web.APIResponseAny{
			Message: "Created worker",
			Data:    worker,
		})
	})
}
