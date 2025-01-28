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
	baseURL string
}

func CreateTransport(svc *Service, baseURL string) *HTTPTransport {
	t := &HTTPTransport{
		service: svc,
		router:  chi.NewRouter(),
		baseURL: baseURL,
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

		page, err := t.service.Query(r.Context(), params.TraceFilter, params.Request)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		web.HTTPResponse(w, http.StatusOK, pagination.CreateResponse(r, t.baseURL, *page))
	}
}

func (t *HTTPTransport) setupRoutes() {
	t.router.Get("/traces", t.httpGetTraces())
}

func (t HTTPTransport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.router.ServeHTTP(w, r)
}
