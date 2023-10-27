package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/core/processing"
	"sensorbucket.nl/sensorbucket/services/dashboard/views"
)

func CreatePipelinePageHandler(ps PipelineStore) http.Handler {
	handler := &PipelinePageHandler{
		router:           chi.NewRouter(),
		pipelineResolver: ps,
	}

	// Setup routes
	handler.router.Get("/", handler.pipelineListPage())
	handler.router.Post("/{pipeline_id}/steps", handler.updatePipelineSteps())
	handler.router.With(handler.resolvePipeline).Get("/edit/{pipeline_id}", pipelineDetailPage())
	return handler
}

type PipelinePageHandler struct {
	router           chi.Router
	pipelineResolver PipelineStore
}

func (h PipelinePageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *PipelinePageHandler) pipelineListPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pipelines, _, err := getPipelines("")
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		page := &views.PipelinePage{
			Pipelines: pipelines,
		}
		if isHX(r) {
			page.WriteBody(w)
			return
		}
		// TODO: what does below do? first page in index rendering?
		views.WriteIndex(w, page)
	}
}

func pipelineDetailPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: redirect to 404 when pipeline is empty
		pipeline := r.Context().Value("pipeline").(processing.Pipeline)
		page := &views.PipelineEditPage{
			Pipeline: pipeline,
		}
		if isHX(r) {
			page.WriteBody(w)
			return
		}
		views.WriteIndex(w, page)
	}
}

func (h *PipelinePageHandler) updatePipelineSteps() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			web.HTTPResponse(w, http.StatusBadRequest, "invalid form")
			return
		}

		newOrder := make([]string, len(r.Form))
		for key, val := range r.Form {
			if len(val) != 1 {
				web.HTTPResponse(w, http.StatusBadRequest, "only 1 value is allowed per key")
				return
			}

			ix, err := strconv.Atoi(val[0])
			if err != nil {
				web.HTTPResponse(w, http.StatusBadRequest, "each value must be a valid index number")
				return
			}
			if ix >= len(newOrder) {
				web.HTTPResponse(w, http.StatusBadRequest, "index cannot be higher than the length of the list")
				return
			}
			newOrder[ix] = key
		}

		h.pipelineResolver

		if isHX(r) {
			views.WriteRenderPipelineStepsSortable(w, newOrder)
		}
	}
}

func (h *PipelinePageHandler) resolvePipeline(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pipelineId, err := URLParamUUID(r, "pipeline_id")
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		pipelines, err := h.pipelineResolver.ListPipelines([]uuid.UUID{pipelineId})
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		if len(pipelines) == 0 {
			web.HTTPResponse(w, http.StatusNotFound, "pipeline not found")
			return
		}
		if len(pipelines) > 1 {
			web.HTTPError(w, fmt.Errorf("API returned more than one result"))
			return
		}
		r = r.WithContext(
			context.WithValue(
				r.Context(),
				"pipeline",
				pipelines[0],
			),
		)
		next.ServeHTTP(w, r)
	})
}

func getPipelines(cursor string) ([]processing.Pipeline, string, error) {
	q := url.Values{}
	q.Set("cursor", cursor)
	// TODO: add filtering based name, description with search string. check if this is a requirement first.
	res, err := http.Get("http://core:3000/pipelines")
	if err != nil {
		return nil, "", err
	}
	var resBody pagination.APIResponse[processing.Pipeline]
	if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
		return nil, "", err
	}
	nextCursor := getCursor(resBody)
	return resBody.Data, nextCursor, nil
}

func URLParamUUID(r *http.Request, name string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, name))
}
