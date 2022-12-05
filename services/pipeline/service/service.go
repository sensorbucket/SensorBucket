package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"sensorbucket.nl/sensorbucket/internal/web"
)

var (
	ErrPipelineNotFound = errors.New("pipeline not found")
)

type Store interface {
	CreatePipeline(*Pipeline) error
	UpdatePipeline(string, UpdatePipelineDTO) error
	ListPipelines() ([]Pipeline, error)
	GetPipeline(string) (*Pipeline, error)
}

type Service struct {
	router chi.Router
	store  Store
}

func New(store Store) *Service {
	r := chi.NewRouter()
	s := &Service{r, store}

	r.Post("/pipelines", s.httpCreatePipeline())
	r.Get("/pipelines", s.httpListPipelines())
	r.Get("/pipelines/{id}", s.httpGetPipeline())
	r.Patch("/pipelines/{id}", s.httpUpdatePipeline())

	return s
}

func (s Service) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(rw, r)
}

type Pipeline struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Steps       []string `json:"steps"`
}

func (s *Service) httpCreatePipeline() http.HandlerFunc {
	type request struct {
		Description string   `json:"description,omitempty"`
		Steps       []string `json:"steps,omitempty"`
	}
	return func(rw http.ResponseWriter, r *http.Request) {
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Failed to decode request body: %v\n", err)
			web.HTTPResponse(rw, http.StatusBadRequest, web.APIResponse{Message: "Could not decode request body"})
			return
		}

		p := &Pipeline{uuid.Must(uuid.NewRandom()).String(), req.Description, req.Steps}
		if err := s.store.CreatePipeline(p); err != nil {
			log.Printf("Store failed to CreatePipeline: %v\n", err)
			web.HTTPResponse(rw, http.StatusInternalServerError, web.APIResponse{Message: "Internal error"})
			return
		}

		web.HTTPResponse(rw, http.StatusCreated, web.APIResponse{Message: "Created pipeline", Data: p})
	}
}

type UpdatePipelineDTO struct {
	Description *string  `json:"description,omitempty"`
	Steps       []string `json:"steps,omitempty"`
}

func (s *Service) httpUpdatePipeline() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var req UpdatePipelineDTO
		id := chi.URLParam(r, "id")
		if _, err := uuid.Parse(id); err != nil {
			web.HTTPResponse(rw, http.StatusBadRequest, web.APIResponse{Message: "id must be of UUID format"})
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Failed to decode request body: %v\n", err)
			web.HTTPResponse(rw, http.StatusBadRequest, web.APIResponse{Message: "Could not decode request body"})
			return
		}

		if err := s.store.UpdatePipeline(id, req); err != nil {
			log.Printf("Store failed to UpdatePipeline: %v\n", err)
			web.HTTPResponse(rw, http.StatusInternalServerError, web.APIResponse{Message: "Internal error"})
			return
		}

		web.HTTPResponse(rw, http.StatusCreated, web.APIResponse{Message: "Updated pipeline"})
	}
}

func (s *Service) httpListPipelines() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		p, err := s.store.ListPipelines()
		if err != nil {
			log.Printf("Store failed to GetPipeline: %v", err)
			web.HTTPResponse(rw, http.StatusInternalServerError, web.APIResponse{Message: "Internal error"})
			return
		}

		web.HTTPResponse(rw, http.StatusOK, web.APIResponse{Message: "Listed pipelines", Data: p})
	}
}

func (s *Service) httpGetPipeline() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, err := uuid.Parse(id); err != nil {
			web.HTTPResponse(rw, http.StatusBadRequest, web.APIResponse{Message: "id must be of UUID format"})
			return
		}

		p, err := s.store.GetPipeline(id)
		if errors.Is(err, ErrPipelineNotFound) {
			web.HTTPResponse(rw, http.StatusNotFound, web.APIResponse{Message: fmt.Sprintf("Pipeline with id '%s' was not found", id)})
			return
		}
		if err != nil {
			log.Printf("Store failed to GetPipeline: %v", err)
			web.HTTPResponse(rw, http.StatusInternalServerError, web.APIResponse{Message: "Internal error"})
			return
		}

		web.HTTPResponse(rw, http.StatusOK, web.APIResponse{Message: "Fetched pipeline", Data: p})
	}
}
