package routes

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/samber/lo"
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
		With(handler.resolveWorkersInPipeline).
		Get("/edit/{pipeline_id}", pipelineDetailPage())
	handler.router.Get("/create", pipelineDetailPage())
	// handler.router.Post("/create", createPipeline())
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

func pipelineCreatePage() http.HandlerFunc {

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
		workersCursor, ok := r.Context().Value("workers_cursor").(string)
		if !ok {
			web.HTTPResponse(w, http.StatusBadRequest, "no cursor found")
			return
		}
		workersInPipeline, ok := r.Context().Value("pipeline_workers").([]api.UserWorker)
		if !ok {
			web.HTTPResponse(w, http.StatusBadRequest, "no workers in pipeline found")
			return
		}
		page := &views.PipelineEditPage{
			Pipeline:          pipeline,
			Workers:           workers,
			WorkersInPipeline: workersInPipeline,
			WorkersNextPage:   workersCursor,
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

		// TODO: maybe not do another call..
		allWorkers, err := h.getAllWorkers(r, newOrder)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		if isHX(r) {
			views.WriteRenderPipelineStepsSortable(SnackbarSaveSuccessful(w), allWorkers)
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

func (h *PipelinePageHandler) resolveWorkersInPipeline(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pipeline, ok := r.Context().Value("pipeline").(api.Pipeline)
		if !ok {
			web.HTTPResponse(w, http.StatusBadRequest, "invalid pipeline model")
			return
		}

		workersInPipeline, err := h.getAllWorkers(r, pipeline.Steps)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		r = r.WithContext(
			context.WithValue(
				r.Context(),
				"pipeline_workers",
				workersInPipeline,
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

// TODO:
// The system is currently undergoing a change to support userworkers. in the near future it is planned to move the existing workers to this feature as well
// after this change, each step will be a UUID. However, with just the images now, the steps consist of the image names.
// while this feature is being developed we need to accomodate retrieving user workers based on an image.
// This method and it's references can simply be deleted once the workers have been rewritten to userworkers.

// Images:
// - sbox
// - http-import
// - ttn
// - mfm
// - meas
// - mfm-2

func (h *PipelinePageHandler) getAllWorkers(r *http.Request, steps []string) ([]api.UserWorker, error) {
	// First check if there are any image workers left in this pipeline
	workersInPipeline := userWorkersFromImageWorkers(steps)
	if len(workersInPipeline) != len(steps) {
		// Not all workers were image workers, append any remaining workers by getting them from the workers api
		userWorkers, resp, err := h.client.WorkersApi.ListWorkers(r.Context()).Id(steps).Execute()
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, err
		}
		if userWorkers.Data == nil {
			return nil, err
		}
		workersInPipeline = append(workersInPipeline, userWorkers.Data...)
	}

	if len(workersInPipeline) != len(steps) {
		return nil, fmt.Errorf("some pipeline workers not found")
	}

	// Now that we have the full result we need to make sure the order of the workers is the same as the order of the pipeline steps
	res := make([]api.UserWorker, len(workersInPipeline))
	for i, step := range steps {
		worker, ok := lo.Find(workersInPipeline, func(item api.UserWorker) bool {
			return item.Id == step
		})
		if !ok {
			return nil, fmt.Errorf("some pipeline workers were not found")
		}

		res[i] = worker
	}
	return res, nil
}

// This function returns any steps that match an image worker to as an API userworker
func userWorkersFromImageWorkers(steps []string) []api.UserWorker {
	userWorkers := []api.UserWorker{}
	for _, s := range steps {
		if worker, ok := lo.Find(imageWorkers, func(item api.UserWorker) bool {
			return item.Name == s
		}); ok {
			userWorkers = append(userWorkers, worker)
		}
	}
	return userWorkers
}

var imageWorkers = []api.UserWorker{
	imageWorker("sbox", "Worker for the sensorbox"),
	imageWorker("http-import", "http importer"),
	imageWorker("ttn", "ttn description"),
	imageWorker("meas", "stores the measurement in the database (immutable)"),
	imageWorker("mfm", "multiflexmeter"),
	imageWorker("mfm-2", "particulate matter"),
}

func imageWorker(name string, description string) api.UserWorker {
	return api.UserWorker{
		Id:          name,
		Name:        name,
		Major:       1,
		Revision:    0,
		Description: description,
	}
}
