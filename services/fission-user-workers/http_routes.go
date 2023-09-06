package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type HTTPTransport struct {
	server *http.Server
}

func newHTTPTransport(addr string) *HTTPTransport {
	r := chi.NewRouter()
	createRoutes(r)
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

func createRoutes(r chi.Router) {
	r.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello"))
	})
}
