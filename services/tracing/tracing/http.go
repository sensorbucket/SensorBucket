package tracing

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"sensorbucket.nl/sensorbucket/internal/httpfilter"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
)

type HTTPTransport struct {
	service *Service
	router  chi.Router
}

func CreateTransport(svc *Service) *HTTPTransport {
	t := &HTTPTransport{
		service: svc,
		router:  chi.NewRouter(),
	}
	t.setupRoutes()
	return t
}

func (t *HTTPTransport) httpGetTraces() http.HandlerFunc {
	type Params struct {
		TraceFilter
		pagination.Request
	}
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := httpfilter.Parse[Params](r)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		fmt.Printf("params: %v\n", params)

		page, err := t.service.Query(r.Context(), params.TraceFilter, params.Request)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		web.HTTPResponse(w, http.StatusOK, pagination.CreateResponse(r, "/api", *page))
	}
}

func (t *HTTPTransport) setupRoutes() {
	t.router.Get("/traces", t.httpGetTraces())
}

func (t HTTPTransport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.router.ServeHTTP(w, r)
}