package coretransport

import (
	"net/http"

	"sensorbucket.nl/sensorbucket/internal/httpfilter"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/core/measurements"
)

func (transport *CoreTransport) httpGetMeasurements() http.HandlerFunc {
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

		page, err := transport.measurementService.QueryMeasurements(r.Context(), params.Filter, params.Request)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, pagination.CreateResponse(r, transport.baseURL, *page))
	}
}
