package transport

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/go-chi/chi/v5"
	"sensorbucket.nl/sensorbucket/internal/httpfilter"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/measurements/service"
)

// HTTPTransport exposes API endpoints to query measurements.
type HTTPTransport struct {
	router chi.Router
	svc    *service.Service
	url    string
}

func NewHTTP(svc *service.Service, url string) *HTTPTransport {
	t := &HTTPTransport{
		router: chi.NewRouter(),
		svc:    svc,
		url:    url,
	}

	t.router.Get("/health", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("healthy"))
	})
	t.router.Get("/measurements", t.httpGetMeasurements())
	t.router.Get("/datastreams", t.httpListDatastream())

	return t
}

func (t *HTTPTransport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.router.ServeHTTP(w, r)
}

func sliceToSingleHook(t1, t2 reflect.Type, i any) (any, error) {
	if t1.Kind() == reflect.Slice && t2.Kind() != reflect.Slice {
		return reflect.ValueOf(i).Index(0).Interface(), nil
	}
	return i, nil
}

func stringToTimeHook(from, to reflect.Type, data any) (any, error) {
	if to == reflect.TypeOf(time.Time{}) && from == reflect.TypeOf("") {
		return time.Parse(time.RFC3339, data.(string))
	}
	return data, nil
}

func (t *HTTPTransport) httpGetMeasurements() http.HandlerFunc {
	type Params struct {
		service.Filter     `pagination:",squash"`
		pagination.Request `pagination:",squash"`
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
		service.DatastreamFilter
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

func parseTimeRange(r *http.Request) (time.Time, time.Time, error) {
	var zero time.Time
	q := r.URL.Query()

	startStr, err := url.PathUnescape(q.Get("start"))
	if err != nil {
		return zero, zero, fmt.Errorf("invalid start: %w", err)
	}
	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		return zero, zero, fmt.Errorf("invalid start time: %w", err)
	}
	if start.IsZero() {
		return zero, zero, fmt.Errorf("start time is required")
	}

	endStr, err := url.PathUnescape(q.Get("end"))
	if err != nil {
		return zero, zero, fmt.Errorf("invalid end: %w", err)
	}
	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		return zero, zero, fmt.Errorf("invalid start time: %w", err)
	}
	if end.IsZero() {
		return zero, zero, fmt.Errorf("end time is required")
	}

	if start.After(end) {
		return zero, zero, fmt.Errorf("start time must be before end time")
	}

	return start, end, nil
}
