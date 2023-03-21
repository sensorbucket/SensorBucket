package transport

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"sensorbucket.nl/sensorbucket/internal/httpfilter"
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

func (t *HTTPTransport) httpGetMeasurements() http.HandlerFunc {
	createFilter := httpfilter.MustCreate[service.QueryFilters]()
	return func(w http.ResponseWriter, r *http.Request) {
		start, end, err := parseTimeRange(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var filters service.QueryFilters
		if err := createFilter(r.URL.Query(), &filters); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		pagination, err := parsePagination(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		measurements, nextPage, err := t.svc.QueryMeasurements(service.Query{
			Start:   start,
			End:     end,
			Filters: filters,
		}, pagination)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := paginatedResponse{
			Data:  measurements,
			Count: len(measurements),
		}
		if nextPage != nil {
			response.Next, err = t.buildNextURL(r, *nextPage)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		sendJSON(w, response)
	}
}

func (t *HTTPTransport) httpListDatastream() http.HandlerFunc {
	createFilter := httpfilter.MustCreate[service.DatastreamFilter]()
	return func(w http.ResponseWriter, r *http.Request) {
		var filter service.DatastreamFilter
		if err := createFilter(r.URL.Query(), &filter); err != nil {
			web.HTTPError(w, err)
			return
		}

		ds, err := t.svc.ListDatastreams(r.Context(), filter)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
			Data: ds,
		})
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

// paginatedResponse is a paginated response.
type paginatedResponse struct {
	Next  string      `json:"next"`
	Count int         `json:"count"`
	Data  interface{} `json:"data"`
}

func parsePagination(r *http.Request) (service.Pagination, error) {
	var err error
	pagination := service.Pagination{
		Limit: 100,
	}
	q := r.URL.Query()

	if q.Has("cursor") {
		// TODO: Decode cursor to pagination struct
		cursor, err := url.QueryUnescape(r.URL.Query().Get("cursor"))
		if err != nil {
			return pagination, fmt.Errorf("could not get cursor query parameter: %w", err)
		}
		pagination, err = decodePagination(cursor)
		if err != nil {
			return pagination, err
		}
		return pagination, nil
	}

	//
	if q.Has("limit") {
		limitQ := r.URL.Query().Get("limit")
		pagination.Limit, err = strconv.Atoi(limitQ)
		if err != nil {
			return service.Pagination{}, fmt.Errorf("limit must be a number: %w", err)
		}
	}

	return pagination, nil
}

func encodePagination(p service.Pagination) (string, error) {
	jsonData, err := json.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("could not encode pagination: %w", err)
	}
	b64Data := base64.StdEncoding.EncodeToString(jsonData)
	return b64Data, nil
}

func decodePagination(cursor string) (service.Pagination, error) {
	jsonData, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return service.Pagination{}, fmt.Errorf("could not decode pagination cursor: %w", err)
	}
	var p service.Pagination
	if err := json.Unmarshal(jsonData, &p); err != nil {
		return service.Pagination{}, fmt.Errorf("could not decode pagination cursor: %w", err)
	}

	return p, nil
}

func (t *HTTPTransport) buildNextURL(r *http.Request, nextPage service.Pagination) (string, error) {
	cursor, err := encodePagination(nextPage)
	if err != nil {
		return "", err
	}
	q := r.URL.Query()
	q.Set("cursor", cursor)
	return fmt.Sprintf("%s%s?%s", t.url, r.URL.Path, q.Encode()), nil
}

func sendJSON(w http.ResponseWriter, v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
