package service

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"sensorbucket.nl/sensorbucket/internal/httpfilter"
	"sensorbucket.nl/sensorbucket/internal/web"
)

type Transport struct {
	router  chi.Router
	service *Service
}

func NewTransport(svc *Service) *Transport {
	r := chi.NewRouter()
	t := &Transport{
		router:  r,
		service: svc,
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
		var dto CreatePipelineDTO
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
		var dto UpdatePipelineDTO
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

func (PipelineStatus) FromString(str string) (any, error) {
	return StrToStatus(str)
}

func (t *Transport) httpListPipelines() http.HandlerFunc {
	createFilter := httpfilter.MustCreate[PipelinesFilter]()
	return func(rw http.ResponseWriter, r *http.Request) {
		var f PipelinesFilter
		if err := createFilter(r.URL.Query(), &f); err != nil {
			log.Printf("Failed to create filter for ListPipelines: %v\n", err)
			web.HTTPError(rw, err)
			return
		}

		p, err := t.service.ListPipelines(r.Context(), f)
		if err != nil {
			log.Printf("Failed to ListPipelines: %v\n", err)
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, web.APIResponseAny{Message: "Listed pipelines", Data: p})
	}
}

func (t *Transport) httpGetPipeline() http.HandlerFunc {
	createFilter := httpfilter.MustCreate[PipelinesFilter]()
	return func(rw http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, err := uuid.Parse(id); err != nil {
			web.HTTPResponse(rw, http.StatusBadRequest, web.APIResponseAny{Message: "id must be of UUID format"})
			return
		}

		// We parse the pipeline filters to see if status=inactive is in there
		// if it's in there then we show the pipeline even if its disabled.
		var f PipelinesFilter
		if err := createFilter(r.URL.Query(), &f); err != nil {
			log.Printf("Failed to GetPipeline: %v\n", err)
			web.HTTPError(rw, err)
			return
		}
		showInactive := false
		for _, v := range f.Status {
			if v == PipelineInactive {
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