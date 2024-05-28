package coretransport

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/internal/httpfilter"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/core/processing"
)

func (t *CoreTransport) httpCreatePipeline() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var dto processing.CreatePipelineDTO
		if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
			log.Printf("Failed to decode request body: %v\n", err)
			web.HTTPResponse(rw, http.StatusBadRequest, web.APIResponseAny{Message: "Could not decode request body"})
			return
		}

		p, err := t.processingService.CreatePipeline(r.Context(), dto)
		if err != nil {
			log.Printf("Failed to CreatePipeline: %v\n", err)
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusCreated, web.APIResponseAny{Message: "Created pipeline", Data: p})
	}
}

func (t *CoreTransport) httpUpdatePipeline() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var dto processing.UpdatePipelineDTO
		id := chi.URLParam(r, "id")
		if _, err := uuid.Parse(id); err != nil {
			web.HTTPResponse(rw, http.StatusBadRequest, web.APIResponseAny{Message: "id must be of UUID format"})
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
			log.Printf("Failed to decode request body: %v\n", err)
			web.HTTPResponse(rw, http.StatusBadRequest, web.APIResponseAny{Message: "Could not decode request body"})
			return
		}

		if lo.ContainsBy(dto.Steps, func(item string) bool {
			return item == ""
		}) {
			web.HTTPResponse(rw, http.StatusBadRequest, web.APIResponseAny{Message: "Worker with empty ID not allowed"})
			return
		}

		dup := lo.FindDuplicates(dto.Steps)
		if len(dup) != 0 {
			web.HTTPResponse(rw, http.StatusBadRequest, web.APIResponseAny{Message: "Duplicate workers not allowed"})
			return
		}

		if err := t.processingService.UpdatePipeline(r.Context(), id, dto); err != nil {
			log.Printf("Failed to UpdatePipeline: %v\n", err)
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusCreated, web.APIResponseAny{Message: "Updated pipeline"})
	}
}

func (t *CoreTransport) httpListPipelines() http.HandlerFunc {
	type filter struct {
		processing.PipelinesFilter
		pagination.Request
	}
	return func(rw http.ResponseWriter, r *http.Request) {
		filter, err := httpfilter.Parse[filter](r)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}

		page, err := t.processingService.ListPipelines(r.Context(), filter.PipelinesFilter, filter.Request)
		if err != nil {
			log.Printf("Failed to ListPipelines: %v\n", err)
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, pagination.CreateResponse(r, t.baseURL, page))
	}
}

func (t *CoreTransport) httpGetPipeline() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, err := uuid.Parse(id); err != nil {
			web.HTTPResponse(rw, http.StatusBadRequest, web.APIResponseAny{Message: "id must be of UUID format"})
			return
		}

		// We parse the pipeline filters to see if status=inactive is in there
		// if it's in there then we show the pipeline even if its disabled.
		filter, err := httpfilter.Parse[processing.PipelinesFilter](r)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}
		showInactive := false
		for _, v := range filter.Status {
			if v == processing.PipelineInactive {
				showInactive = true
				break
			}
		}

		p, err := t.processingService.GetPipeline(r.Context(), id, showInactive)
		if err != nil {
			log.Printf("Failed to GetPipeline: %v\n", err)
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, web.APIResponseAny{Message: "Fetched pipeline", Data: p})
	}
}

func (t *CoreTransport) httpDeletePipeline() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, err := uuid.Parse(id); err != nil {
			web.HTTPResponse(rw, http.StatusBadRequest, web.APIResponseAny{Message: "id must be of UUID format"})
			return
		}

		if err := t.processingService.DisablePipeline(r.Context(), id); err != nil {
			log.Printf("Failed to GetPipeline: %v\n", err)
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, web.APIResponseAny{Message: "Pipeline set to inactive"})
	}
}
