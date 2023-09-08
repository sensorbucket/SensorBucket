package tracingtransport

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"sensorbucket.nl/sensorbucket/internal/httpfilter"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/tracing/tracing"
)

type HTTPTransport struct {
	router chi.Router
	svc    *tracing.Service
	url    string
}

func NewHTTP(svc *tracing.Service, url string) *HTTPTransport {
	t := &HTTPTransport{
		router: chi.NewRouter(),
		svc:    svc,
		url:    url,
	}
	t.SetupRoutes(t.router)
	return t
}

func (t *HTTPTransport) SetupRoutes(r chi.Router) {
	r.Get("/tracing", t.httpGetTraces())
}

func (t *HTTPTransport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.router.ServeHTTP(w, r)
}

func (t *HTTPTransport) httpGetTraces() http.HandlerFunc {
	type Params struct {
		tracing.Filter     `pagination:",squash"`
		pagination.Request `pagination:",squash"`
	}
	return func(rw http.ResponseWriter, r *http.Request) {
		params, err := httpfilter.Parse[Params](r)
		if err != nil {
			web.HTTPError(rw, web.NewError(http.StatusBadRequest, "invalid params", ""))
			return
		}

		if params.DurationGreaterThan != nil &&
			params.DurationLowerThan != nil &&
			*params.DurationGreaterThan >= *params.DurationLowerThan {
			web.HTTPError(rw, web.NewError(http.StatusBadRequest, "duration_greater_than cannot be greater than or equal to duration_smaller_than", ""))
			return
		}

		page, err := t.svc.QueryTraces(params.Filter, params.Request)
		if err != nil {
			log.Printf("[Error] %v\n", err)
			web.HTTPError(rw, fmt.Errorf("internal server error"))
			return
		}

		web.HTTPResponse(rw, http.StatusOK, pagination.CreateResponse(r, t.url, *page))
	}
}
