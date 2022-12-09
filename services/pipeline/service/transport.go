package service

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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
			web.HTTPResponse(rw, http.StatusBadRequest, web.APIResponse{Message: "Could not decode request body"})
			return
		}

		p, err := t.service.CreatePipeline(r.Context(), dto)
		if err != nil {
			log.Printf("Failed to CreatePipeline: %v\n", err)
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusCreated, web.APIResponse{Message: "Created pipeline", Data: p})
	}
}

func (t *Transport) httpUpdatePipeline() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var dto UpdatePipelineDTO
		id := chi.URLParam(r, "id")
		if _, err := uuid.Parse(id); err != nil {
			web.HTTPResponse(rw, http.StatusBadRequest, web.APIResponse{Message: "id must be of UUID format"})
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
			log.Printf("Failed to decode request body: %v\n", err)
			web.HTTPResponse(rw, http.StatusBadRequest, web.APIResponse{Message: "Could not decode request body"})
			return
		}

		if err := t.service.UpdatePipeline(r.Context(), id, dto); err != nil {
			log.Printf("Failed to UpdatePipeline: %v\n", err)
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusCreated, web.APIResponse{Message: "Updated pipeline"})
	}
}

func (t *Transport) httpListPipelines() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		f := NewPipelinesFilter()
		f.OnlyInactive = r.URL.Query().Has("inactive")

		p, err := t.service.ListPipelines(r.Context(), f)
		if err != nil {
			log.Printf("Failed to GetPipeline: %v", err)
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, web.APIResponse{Message: "Listed pipelines", Data: p})
	}
}

func (t *Transport) httpGetPipeline() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, err := uuid.Parse(id); err != nil {
			web.HTTPResponse(rw, http.StatusBadRequest, web.APIResponse{Message: "id must be of UUID format"})
			return
		}
		f := NewPipelinesFilter()
		if r.URL.Query().Has("inactive") {
			f.OnlyInactive = true
		}

		p, err := t.service.GetPipeline(r.Context(), id, f.OnlyInactive)
		if err != nil {
			log.Printf("Failed to GetPipeline: %v", err)
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, web.APIResponse{Message: "Fetched pipeline", Data: p})
	}
}

func (t *Transport) httpDeletePipeline() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, err := uuid.Parse(id); err != nil {
			web.HTTPResponse(rw, http.StatusBadRequest, web.APIResponse{Message: "id must be of UUID format"})
			return
		}

		if err := t.service.DisablePipeline(r.Context(), id); err != nil {
			log.Printf("Failed to GetPipeline: %v", err)
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, web.APIResponse{Message: "Pipeline set to inactive"})
	}
}
