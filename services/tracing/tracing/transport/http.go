package tracingtransport

import (
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

		if params.DurationGreaterThan != nil && *params.DurationSmallerThan != 0 &&
			*params.DurationGreaterThan <= *params.DurationSmallerThan {
			http.Error(rw, "duration_greater_than cannot be smaller than or equal to duration_smaller_than", http.StatusBadRequest)
			return
		}

		page, err := t.svc.QueryTraces(params.Filter, params.Request)
		if err != nil {
			log.Printf("[Error] %v\n", err)
			http.Error(rw, "internal server error", http.StatusInternalServerError)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, pagination.CreateResponse(r, t.url, *page))
	}
}
