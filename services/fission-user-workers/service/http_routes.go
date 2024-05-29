package userworkers

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"sensorbucket.nl/sensorbucket/internal/httpfilter"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/auth"
)

type HTTPTransport struct {
	server *http.Server
}

func NewHTTPTransport(app *Application, baseURL, addr string, keySource auth.JWKSClient) *HTTPTransport {
	r := chi.NewRouter()
	r.Use(
		chimw.Logger,
		auth.Authenticate(keySource),
		auth.Protect(),
	)
	createRoutes(app, baseURL, r)
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

func (t *HTTPTransport) Start() error {
	log.Printf("HTTP server listening at: %s\n", t.server.Addr)
	return t.server.ListenAndServe()
}

func (t *HTTPTransport) Stop(ctx context.Context) error {
	log.Println("HTTP Server shutting down...")
	return t.server.Shutdown(ctx)
}

type WorkersHTTPFilters struct {
	pagination.Request
	ListWorkerFilters
}

func createRoutes(app *Application, baseURL string, r chi.Router) {
	r.Get("/workers", func(w http.ResponseWriter, r *http.Request) {
		params, err := httpfilter.Parse[WorkersHTTPFilters](r)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		page, err := app.ListWorkers(r.Context(), params.ListWorkerFilters, params.Request)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusOK, pagination.CreateResponse(r, baseURL, *page))
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
	r.With(resolveWorker(app)).Get("/workers/{id}", func(w http.ResponseWriter, r *http.Request) {
		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
			Data: getWorker(r.Context()),
		})
	})
	r.With(resolveWorker(app)).Get("/workers/{id}/usercode", func(w http.ResponseWriter, r *http.Request) {
		worker := getWorker(r.Context())
		userCode, err := worker.GetUserCode()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
			Data: userCode,
		})
	})
	r.With(resolveWorker(app)).Get("/workers/{id}/source", func(w http.ResponseWriter, r *http.Request) {
		worker := getWorker(r.Context())
		src := base64.StdEncoding.EncodeToString(worker.ZipSource)
		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
			Data: src,
		})
	})
	r.With(resolveWorker(app)).Patch("/workers/{id}", func(w http.ResponseWriter, r *http.Request) {
		var dto UpdateWorkerOpts
		if err := web.DecodeJSON(r, &dto); err != nil {
			web.HTTPError(w, err)
			return
		}
		worker := getWorker(r.Context())
		if err := app.UpdateWorker(r.Context(), worker, dto); err != nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
			Message: "Worker has been updated",
		})
	})
}

type Middleware = func(next http.Handler) http.Handler

type ctxKey int

const (
	ctxWorker ctxKey = iota
)

func resolveWorker(app *Application) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			idStr := chi.URLParam(r, "id")
			id, err := uuid.Parse(idStr)
			if err != nil {
				web.HTTPError(w, ErrInvalidUUID)
				return
			}
			worker, err := app.GetWorker(r.Context(), id)
			if err != nil {
				web.HTTPError(w, err)
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), ctxWorker, worker))
			next.ServeHTTP(w, r)
		})
	}
}

func getWorker(ctx context.Context) *UserWorker {
	value := ctx.Value(ctxWorker)
	if value == nil {
		return nil
	}
	worker, ok := value.(*UserWorker)
	if !ok {
		return nil
	}
	return worker
}
