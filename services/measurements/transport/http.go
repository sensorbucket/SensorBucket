package transport

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"sensorbucket.nl/services/measurements/service"
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

	t.router.Get("/{start}/{end}", t.httpGetMeasurements())

	return t
}

func (t *HTTPTransport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.router.ServeHTTP(w, r)
}

func (t *HTTPTransport) httpGetMeasurements() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start, end, err := parseTimeRange(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		filters, err := parseFilters(r)
		if err != nil {
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
			response.Next = t.buildNextURL(r, *nextPage)
		}
		sendJSON(w, response)
	}
}

func parseTimeRange(r *http.Request) (time.Time, time.Time, error) {
	var zero time.Time

	startStr, err := url.PathUnescape(chi.URLParam(r, "start"))
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

	endStr, err := url.PathUnescape(chi.URLParam(r, "end"))
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

func parseFilters(r *http.Request) (service.QueryFilters, error) {
	var filters service.QueryFilters
	q := r.URL.Query()

	if len(q["thing_urn"]) > 0 {
		filters.ThingURNs = q["thing_urn"]
	}

	if len(q["location_id"]) > 0 {
		filters.LocationIDs = make([]int, 0, len(q["location_id"]))
		for _, valQ := range q["location_id"] {
			valStr, err := url.QueryUnescape(valQ)
			if err != nil {
				return filters, fmt.Errorf("invalid location_id: %w", err)
			}
			id, err := strconv.Atoi(valStr)
			if err != nil {
				return filters, fmt.Errorf("invalid location_id: %w", err)
			}
			filters.LocationIDs = append(filters.LocationIDs, id)
		}
	}

	if len(q["measurement_type"]) > 0 {
		filters.MeasurementTypes = q["measurement_type"]
	}

	return filters, nil
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
		pagination.Cursor, err = url.QueryUnescape(r.URL.Query().Get("cursor"))
		if err != nil {
			return service.Pagination{}, fmt.Errorf("invalid cursor: %w", err)
		}
	}

	if q.Has("limit") {
		limitQ := r.URL.Query().Get("limit")
		pagination.Limit, err = strconv.Atoi(limitQ)
		if err != nil {
			return service.Pagination{}, fmt.Errorf("limit must be a number: %w", err)
		}
	}

	return pagination, nil
}

func (t *HTTPTransport) buildNextURL(r *http.Request, nextPage service.Pagination) string {
	q := r.URL.Query()
	q.Set("cursor", nextPage.Cursor)
	q.Set("limit", strconv.Itoa(nextPage.Limit))
	return fmt.Sprintf("%s%s?%s", t.url, r.URL.Path, q.Encode())
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
