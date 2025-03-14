package coretransport

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"sensorbucket.nl/sensorbucket/internal/httpfilter"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/core/featuresofinterest"
)

func (transport *CoreTransport) httpListFeaturesOfInterest() http.HandlerFunc {
	type Params struct {
		featuresofinterest.FeatureOfInterestFilter
		pagination.Request
	}
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := httpfilter.Parse[Params](r)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		page, err := transport.featureOfInterestService.ListFeaturesOfInterest(r.Context(), params.FeatureOfInterestFilter, params.Request)
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
		feature, err := transport.featureOfInterestService.GetFeatureOfInterest(r.Context(), id)
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

func (transport *CoreTransport) httpCreateFeatureOfInterest() http.HandlerFunc {
	type DTO struct {
		Name         string          `json:"name"`
		Description  *string         `json:"description"`
		EncodingType *string         `json:"encoding_type"`
		Feature      json.RawMessage `json:"feature"`
		Properties   json.RawMessage `json:"properties"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var dto DTO
		if err := web.DecodeJSON(r, &dto); err != nil {
			web.HTTPError(w, err)
			return
		}

		feature, err := transport.featureOfInterestService.CreateFeatureOfInterest(r.Context(), featuresofinterest.CreateFeatureOfInterestOpts{
			Name:         dto.Name,
			Description:  dto.Description,
			EncodingType: dto.EncodingType,
			Feature:      dto.Feature,
			Properties:   dto.Properties,
		})
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
			Message: "Created new FeatureOfInterest",
			Data:    feature,
		})
	}
}

func (transport *CoreTransport) httpUpdateFeatureOfInterest() http.HandlerFunc {
	type DTO struct {
		Name         *string                      `json:"name"`
		Description  *string                      `json:"description"`
		EncodingType *string                      `json:"encoding_type"`
		Feature      *featuresofinterest.Geometry `json:"feature"`
		Properties   json.RawMessage              `json:"properties"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "FeatureOfInterest ID must be a number",
			})
			return
		}
		var dto DTO
		if err := web.DecodeJSON(r, &dto); err != nil {
			web.HTTPError(w, err)
			return
		}

		if err := transport.featureOfInterestService.UpdateFeatureOfInterest(r.Context(), id, featuresofinterest.UpdateFeatureOfInterestOpts{
			Name:         dto.Name,
			Description:  dto.Description,
			EncodingType: dto.EncodingType,
			Feature:      dto.Feature,
			Properties:   dto.Properties,
		}); err != nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
			Message: "Updated new FeatureOfInterest",
		})
	}
}

func (transport *CoreTransport) httpDeleteFeaturOfInterest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			web.HTTPResponse(w, http.StatusBadRequest, web.APIResponseAny{
				Message: "FeatureOfInterest ID must be a number",
			})
			return
		}
		if err := transport.featureOfInterestService.DeleteFeatureOfInterest(r.Context(), id); err != nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
			Message: "Deleted feature of interest",
		})
	}
}
