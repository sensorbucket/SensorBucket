package coretransport

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"sensorbucket.nl/sensorbucket/internal/httpfilter"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/core/measurements"
)

func (transport *CoreTransport) httpListFeaturesOfInterest() http.HandlerFunc {
	type params struct {
		measurements.FeatureOfInterestFilter
		pagination.Request
	}
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := httpfilter.Parse[params](r)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		page, err := transport.measurementService.ListFeaturesOfInterest(r.Context(), params.FeatureOfInterestFilter, params.Request)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusOK, pagination.CreateResponse(r, "", *page))
	}
}

func (transport *CoreTransport) httpGetFeatureOfInterest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "FeatureOfInterest ID must be a number",
			})
			return
		}
		feature, err := transport.measurementService.GetFeatureOfInterest(r.Context(), id)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
			Message: "",
			Data:    feature,
		})
	}
}
