package tracingtransport

import (
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
	r.Get("/health", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("healthy"))
	})
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
			web.HTTPError(rw, err)
			return
		}

		if *params.Filter.TraceIds == nil || len(*params.Filter.TraceIds) == 0 {
			http.Error(rw, "at least 1 trace_id must be included in the request", http.StatusBadRequest)
			return
		}

		page, err := t.svc.QueryTraces(params.Filter, params.Request)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, pagination.CreateResponse(r, t.url, *page))
	}
}
