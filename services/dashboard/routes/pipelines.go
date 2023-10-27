package routes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/internal/env"
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

	handler.router.
		With(handler.validatePipelineSteps).
		Patch("/validate", handler.pipelineStepsView())

	handler.router.
		With(handler.validatePipelineSteps).
		With(handler.updatePipeline).
		Patch("/edit/{pipeline_id}", handler.pipelineStepsView())

	handler.router.
		With(handler.resolveWorkers).
		With(handler.resolvePipeline).
		With(handler.resolveWorkersInPipeline).
		Get("/edit/{pipeline_id}", pipelineDetailPage())

	handler.router.
		With(handler.resolveWorkers).
		Get("/create", pipelineCreatePage())
	handler.router.With(handler.validatePipelineSteps).Post("/create", handler.createPipeline())

	handler.router.Get("/workers/table", handler.getWorkersTable())
	handler.router.Get("/table", handler.getPipelinesTable())
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
		pipelines, _, err := h.client.PipelinesApi.ListPipelines(r.Context()).Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		page := &views.PipelinePage{
			Pipelines: pipelines.Data,
		}
		if pipelines.Links.GetNext() != "" {
			page.PipelinesNextPage = "/pipelines/table?cursor=" + getCursor(pipelines.Links.GetNext())
		}
		if isHX(r) {
			page.WriteBody(w)
			return
		}
		views.WriteIndex(w, page)
	}
}

func (h *PipelinePageHandler) createPipeline() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			web.HTTPError(w, web.NewError(http.StatusBadRequest, "Bad request", ""))
			return
		}

		allWorkers, ok := r.Context().Value("pipeline_workers").([]api.UserWorker)
		if !ok {
			WithSnackbarError(w, "Couldn't find workers", http.StatusBadRequest)
			return
		}

		var dto api.CreatePipelineRequest
		dto.SetDescription(r.FormValue("pipeline-descr"))
		dto.SetSteps(stepsFromWorkersList(allWorkers))

		_, resp, err := h.client.PipelinesApi.CreatePipeline(r.Context()).CreatePipelineRequest(dto).Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		if resp.StatusCode != http.StatusCreated {
			SnackbarSomethingWentWrong(w)
			return
		}

		w.Header().Set("HX-Redirect", "/pipelines")
		w.WriteHeader(http.StatusOK)
	}
}

func pipelineCreatePage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		page := &views.PipelineEditPage{
			Pipeline:          nil,
			Workers:           workers,
			WorkersInPipeline: nil,
			WorkersNextPage:   workersCursor,
		}
		if isHX(r) {
			page.WriteBody(w)
			return
		}
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
			Pipeline:          &pipeline,
			Workers:           workers,
			WorkersInPipeline: &workersInPipeline,
			WorkersNextPage:   workersCursor,
		}
		if isHX(r) {
			page.WriteBody(w)
			return
		}
		views.WriteIndex(w, page)
	}
}

func (h *PipelinePageHandler) updatePipeline(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			web.HTTPResponse(w, http.StatusBadRequest, "invalid form")
			return
		}
		pipelineId := chi.URLParam(r, "pipeline_id")
		if pipelineId == "" {
			WithSnackbarError(w, "Pipeline must be given", http.StatusBadRequest)
			return
		}
		pipelineDescr, ok := r.Form["pipeline-descr"]
		if !ok {
			WithSnackbarError(w, "Pipeline description cannot be empty", http.StatusBadRequest)
			return
		}

		if len(pipelineDescr) != 1 {
			WithSnackbarError(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if pipelineDescr[0] == "" {
			WithSnackbarError(w, "Pipeline description cannot be emptry", http.StatusBadRequest)
			return
		}

		allWorkers, ok := r.Context().Value("pipeline_workers").([]api.UserWorker)
		if !ok {
			WithSnackbarError(w, "Couldn't find workers", http.StatusBadRequest)
			return
		}

		var updateDto api.UpdatePipelineRequest
		updateDto.SetDescription(pipelineDescr[0])
		updateDto.SetSteps(stepsFromWorkersList(allWorkers))

		_, resp, err := h.client.PipelinesApi.UpdatePipeline(r.Context(), pipelineId).UpdatePipelineRequest(updateDto).Execute()
		if err != nil {
			SnackbarSomethingWentWrong(w)
			return
		}

		// TODO: API returns status created instead of found for some reason
		if resp.StatusCode != http.StatusCreated {
			if resp.StatusCode == http.StatusInternalServerError {
				SnackbarSomethingWentWrong(w)
			} else {
				var apierror *web.APIError
				if errors.As(err, &apierror) {
					WithSnackbarError(w, apierror.Message, apierror.HTTPStatus)
					return
				}
			}
			return
		}

		SnackbarSaveSuccessful(w)
		next.ServeHTTP(w, r)
	})
}

func (h *PipelinePageHandler) pipelineStepsView() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allWorkers, ok := r.Context().Value("pipeline_workers").([]api.UserWorker)
		if !ok {
			web.HTTPResponse(w, http.StatusBadRequest, "Couldn't find workers")
			return
		}

		if isHX(r) {
			views.WriteRenderPipelineStepsSortable(w, allWorkers)
		}
	}
}

func (h *PipelinePageHandler) validatePipelineSteps(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			web.HTTPResponse(w, http.StatusBadRequest, "invalid form")
			return
		}
		steps, ok := r.Form["steps"]
		if !ok {
			WithSnackbarError(w, "No steps provided", http.StatusBadRequest)
			return
		}

		if len(steps) != 1 {
			SnackbarSomethingWentWrong(w)
			return
		}

		stepMap := map[string][]string{}
		err = json.Unmarshal([]byte(steps[0]), &stepMap)
		if err != nil {
			WithSnackbarError(w, "Invalid steps", http.StatusBadRequest)
			return
		}

		if len(stepMap) == 0 {
			WithSnackbarError(w, "No steps provided", http.StatusBadRequest)
			return
		}

		newOrder := make([]string, len(stepMap))
		for key, val := range stepMap {
			if len(val) != 1 {
				WithSnackbarError(w, "Duplicate workers are not allowed", http.StatusBadRequest)
				return
			}

			ix, err := strconv.Atoi(val[0])
			if err != nil {
				WithSnackbarError(w, "Invalid input", http.StatusBadRequest)
				return
			}
			if ix >= len(newOrder) || ix < 0 {
				WithSnackbarError(w, "Invalid input", http.StatusBadRequest)
				return
			}
			newOrder[ix] = key
		}

		if newOrder[len(newOrder)-1] != IMAGE_MEAS {
			WithSnackbarError(w, "Last step must be measurement storage", http.StatusBadRequest)
			return
		}

		allWorkers, err := h.getAllWorkers(r, newOrder)
		if err != nil {
			SnackbarSomethingWentWrong(w)
			return
		}

		r = r.WithContext(
			context.WithValue(
				r.Context(),
				"pipeline_workers",
				allWorkers,
			),
		)

		next.ServeHTTP(w, r)
	})
}

func stepsFromWorkersList(workers []api.UserWorker) []string {
	res := make([]string, len(workers))
	for i, w := range workers {
		res[i] = w.Id
	}
	return res
}

func (h *PipelinePageHandler) getWorkersTable() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req := h.client.WorkersApi.ListWorkers(r.Context())
		if r.URL.Query().Has("cursor") {
			req = req.Cursor(r.URL.Query().Get("cursor"))
		}
		res, _, err := req.Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		nextCursor := ""
		if res.Links.GetNext() != "" {
			nextCursor = "/pipelines/workers/table?cursor=" + getCursor(res.Links.GetNext())
		}
		views.WriteRenderPipelineEditWorkerTableRows(w, res.Data, nextCursor)
	}
}

func (h *PipelinePageHandler) getPipelinesTable() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req := h.client.PipelinesApi.ListPipelines(r.Context())
		if r.URL.Query().Has("cursor") {
			req = req.Cursor(r.URL.Query().Get("cursor"))
		}
		res, _, err := req.Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		nextCursor := ""
		if res.Links.GetNext() != "" {
			nextCursor = "/pipelines/table?cursor=" + getCursor(res.Links.GetNext())
		}
		views.WriteRenderPipelineTableRows(w, res.Data, nextCursor)
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
			append(imageWorkers, workers.Data...),
		)

		nextCursor := ""
		if workers.Links.GetNext() != "" {
			nextCursor = "/pipelines/workers/table?cursor=" + getCursor(workers.Links.GetNext())
		}

		ctx = context.WithValue(
			ctx,
			"workers_cursor",
			nextCursor,
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

var (
	IMAGE_SBOX = env.Must("IMAGE_SBOX")
	IMAGE_TTN  = env.Must("IMAGE_TTN")
	IMAGE_MEAS = env.Must("IMAGE_MEAS")
	IMAGE_MFM  = env.Must("IMAGE_MFM")
	IMAGE_MFM2 = env.Must("IMAGE_MFM2")
)

var imageWorkers = []api.UserWorker{
	imageWorker(IMAGE_SBOX, "Worker for the sensorbox"),
	imageWorker(IMAGE_TTN, "ttn description"),
	imageWorker(IMAGE_MEAS, "stores the measurement in the database (immutable)"),
	imageWorker(IMAGE_MFM, "multiflexmeter"),
	imageWorker(IMAGE_MFM2, "particulate matter"),
}

func imageWorker(name string, description string) api.UserWorker {
	return api.UserWorker{
		Id:          name,
		Name:        name,
		Revision:    0,
		Description: description,
	}
}
