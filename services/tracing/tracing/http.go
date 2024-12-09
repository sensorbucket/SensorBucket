package tracing

import (
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

type httpGetTracesParameters struct {
	TraceFilter
	pagination.Request
}

func (t *HTTPTransport) httpGetTraces() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := httpfilter.Parse[httpGetTracesParameters](r)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

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
