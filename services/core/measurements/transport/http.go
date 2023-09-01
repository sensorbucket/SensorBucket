package measurementtransport

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"sensorbucket.nl/sensorbucket/internal/httpfilter"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/web"
	"sensorbucket.nl/sensorbucket/services/core/measurements"
)

// HTTPTransport exposes API endpoints to query measurements.
type HTTPTransport struct {
	router chi.Router
	svc    *measurements.Service
	url    string
}

func NewHTTP(svc *measurements.Service, url string) *HTTPTransport {
	t := &HTTPTransport{
		router: chi.NewRouter(),
		svc:    svc,
		url:    url,
	}
	t.SetupRoutes(t.router)
	return t
}

func (t *HTTPTransport) SetupRoutes(r chi.Router) {
	r.Get("/health", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("healthy"))
	})
	r.Get("/measurements", t.httpGetMeasurements())
	r.Get("/datastreams", t.httpListDatastream())
}

func (t *HTTPTransport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.router.ServeHTTP(w, r)
}

func (t *HTTPTransport) httpGetMeasurements() http.HandlerFunc {
	type Params struct {
		measurements.Filter `pagination:",squash"`
		pagination.Request  `pagination:",squash"`
	}
	return func(rw http.ResponseWriter, r *http.Request) {
		params, err := httpfilter.Parse[Params](r)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}

		page, err := t.svc.QueryMeasurements(params.Filter, params.Request)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, pagination.CreateResponse(r, t.url, *page))
	}
}

func (t *HTTPTransport) httpListDatastream() http.HandlerFunc {
	type params struct {
		measurements.DatastreamFilter
		pagination.Request
	}
	return func(rw http.ResponseWriter, r *http.Request) {
		params, err := httpfilter.Parse[params](r)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}

		page, err := t.svc.ListDatastreams(r.Context(), params.DatastreamFilter, params.Request)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}
		web.HTTPResponse(rw, http.StatusOK, pagination.CreateResponse(r, t.url, *page))
	}
}
