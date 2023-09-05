package measurementtransport

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"sensorbucket.nl/sensorbucket/internal/httpfilter"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
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
	r.Get("/measurements", t.httpGetMeasurements())
	r.Get("/datastreams", t.httpListDatastream())
	r.Get("/datastreams/{id}", t.httpGetDatastream())
}

func (t *HTTPTransport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.router.ServeHTTP(w, r)
}

func (t *HTTPTransport) httpGetMeasurements() http.HandlerFunc {
	type Params struct {
		measurements.Filter
		pagination.Request
	}
	return func(rw http.ResponseWriter, r *http.Request) {
		params, err := httpfilter.Parse[Params](r)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}

		if !params.Start.IsZero() && !params.End.IsZero() && params.Start.After(params.End) {
			web.HTTPError(rw, web.NewError(http.StatusBadRequest, "Start time cannot be after end time", "ERR_BAD_REQUEST"))
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

func (t *HTTPTransport) httpGetDatastream() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idQ := chi.URLParam(r, "id")
		id, err := uuid.Parse(idQ)
		if err != nil {
			web.HTTPError(w, web.NewError(http.StatusBadRequest, "Invalid datastream ID", ""))
			return
		}

		ds, err := t.svc.GetDatastream(r.Context(), id)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{Data: ds})
	}
}
