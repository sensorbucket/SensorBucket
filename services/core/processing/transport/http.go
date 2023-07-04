package processingtransport

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"sensorbucket.nl/sensorbucket/internal/httpfilter"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/core/processing"
)

type Transport struct {
	router  chi.Router
	service *processing.Service
	baseURL string
}

func NewTransport(svc *processing.Service, baseURL string) *Transport {
	r := chi.NewRouter()
	t := &Transport{
		router:  r,
		service: svc,
		baseURL: baseURL,
	}

	r.Post("/pipelines", t.httpCreatePipeline())
	r.Get("/pipelines", t.httpListPipelines())
	r.Get("/pipelines/{id}", t.httpGetPipeline())
	r.Patch("/pipelines/{id}", t.httpUpdatePipeline())
	r.Delete("/pipelines/{id}", t.httpDeletePipeline())

	return t
}

func (t Transport) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	t.router.ServeHTTP(rw, r)
}

func (t *Transport) httpCreatePipeline() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var dto processing.CreatePipelineDTO
		if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
			log.Printf("Failed to decode request body: %v\n", err)
			web.HTTPResponse(rw, http.StatusBadRequest, web.APIResponseAny{Message: "Could not decode request body"})
			return
		}

		p, err := t.service.CreatePipeline(r.Context(), dto)
		if err != nil {
			log.Printf("Failed to CreatePipeline: %v\n", err)
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusCreated, web.APIResponseAny{Message: "Created pipeline", Data: p})
	}
}

func (t *Transport) httpUpdatePipeline() http.HandlerFunc {
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

		if err := t.service.UpdatePipeline(r.Context(), id, dto); err != nil {
			log.Printf("Failed to UpdatePipeline: %v\n", err)
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusCreated, web.APIResponseAny{Message: "Updated pipeline"})
	}
}

func (t *Transport) httpListPipelines() http.HandlerFunc {
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

		page, err := t.service.ListPipelines(r.Context(), filter.PipelinesFilter, filter.Request)
		if err != nil {
			log.Printf("Failed to ListPipelines: %v\n", err)
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, pagination.CreateResponse(r, t.baseURL, page))
	}
}

func (t *Transport) httpGetPipeline() http.HandlerFunc {
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

		p, err := t.service.GetPipeline(r.Context(), id, showInactive)
		if err != nil {
			log.Printf("Failed to GetPipeline: %v\n", err)
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, web.APIResponseAny{Message: "Fetched pipeline", Data: p})
	}
}

func (t *Transport) httpDeletePipeline() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, err := uuid.Parse(id); err != nil {
			web.HTTPResponse(rw, http.StatusBadRequest, web.APIResponseAny{Message: "id must be of UUID format"})
			return
		}

		if err := t.service.DisablePipeline(r.Context(), id); err != nil {
			log.Printf("Failed to GetPipeline: %v\n", err)
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, web.APIResponseAny{Message: "Pipeline set to inactive"})
	}
}
