package routes

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/api"
	"sensorbucket.nl/sensorbucket/services/dashboard/views"
)

func CreatePipelinePageHandler(client *api.APIClient) http.Handler {
	handler := &PipelinePageHandler{
		router: chi.NewRouter(),
		client: client,
	}

	// Setup routes
	handler.router.Get("/", handler.pipelineListPage())
	handler.router.Post("/{pipeline_id}/steps", handler.updatePipelineSteps())
	handler.router.
		With(handler.resolveWorkers).
		With(handler.resolvePipeline).
		Get("/edit/{pipeline_id}", pipelineDetailPage())
	return handler
}

type PipelinePageHandler struct {
	router chi.Router
	client *api.APIClient
}

func (h PipelinePageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *PipelinePageHandler) pipelineListPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pipelines, _, err := h.client.PipelinesApi.ListPipelines(r.Context()).Cursor("").Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		page := &views.PipelinePage{
			Pipelines:         pipelines.Data,
			PipelinesNextPage: "",
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
		pipeline, ok := r.Context().Value("pipeline").(api.Pipeline)
		if !ok {
			web.HTTPResponse(w, http.StatusBadRequest, "invalid pipeline model")
			return
		}
		workers, ok := r.Context().Value("workers").([]api.UserWorker)
		if !ok {
			web.HTTPResponse(w, http.StatusBadRequest, "invalid workers array")
			return
		}
		workersCursor := r.Context().Value("workers_cursor").(string)
		page := &views.PipelineEditPage{
			Pipeline:        pipeline,
			Workers:         workers,
			WorkersNextPage: workersCursor,
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
		pipelineId := chi.URLParam(r, "pipeline_id")
		if pipelineId == "" {
			web.HTTPResponse(w, http.StatusBadRequest, "pipeline_id cannot be empty")
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

		updatDto := api.UpdatePipelineRequest{
			Steps: newOrder,
		}
		_, resp, err := h.client.PipelinesApi.UpdatePipeline(r.Context(), pipelineId).UpdatePipelineRequest(updatDto).Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		// TODO: API returns status created instead of found for some reason
		if resp.StatusCode != http.StatusCreated {
			web.HTTPResponse(w, resp.StatusCode, "pipeline not found")
			return
		}

		if isHX(r) {
			views.WriteRenderPipelineStepsSortable(w, newOrder)
		}
	}
}

func (h *PipelinePageHandler) resolvePipeline(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pipeline, resp, err := h.client.PipelinesApi.GetPipeline(r.Context(), chi.URLParam(r, "pipeline_id")).Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			web.HTTPResponse(w, resp.StatusCode, "status not ok")
			return
		}
		if pipeline.Data == nil {
			web.HTTPResponse(w, http.StatusNotFound, "pipeline not found")
			return
		}
		r = r.WithContext(
			context.WithValue(
				r.Context(),
				"pipeline",
				*pipeline.Data,
			),
		)
		next.ServeHTTP(w, r)
	})
}

func (h *PipelinePageHandler) resolveWorkers(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		workers, resp, err := h.client.WorkersApi.ListWorkers(r.Context()).Cursor("").Execute()
		if err != nil {
			fmt.Println("workers not found!!", err)
			web.HTTPError(w, err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			web.HTTPResponse(w, resp.StatusCode, "status not ok")
			return
		}
		ctx := context.WithValue(
			r.Context(),
			"workers",
			workers.Data,
		)
		ctx = context.WithValue(
			ctx,
			"workers_cursor",
			"",
		)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
