package routes

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ory/nosurf"
	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/api"
	"sensorbucket.nl/sensorbucket/pkg/layout"
	"sensorbucket.nl/sensorbucket/services/dashboard/views"
)

var STORAGE_STEP = env.Could("STORAGE_STEP", "storage")

type PipelinePageHandler struct {
	router        chi.Router
	workersClient *api.APIClient
	coreClient    *api.APIClient
}

func CreatePipelinePageHandler(workers, core *api.APIClient) http.Handler {
	handler := &PipelinePageHandler{
		router:        chi.NewRouter(),
		workersClient: workers,
		coreClient:    core,
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

func (h PipelinePageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *PipelinePageHandler) pipelineListPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pipelines, _, err := h.coreClient.PipelinesApi.ListPipelines(r.Context()).Execute()
		if err != nil {
			web.HTTPError(w, fmt.Errorf("could not list pipelines: %w", err))
			return
		}

		page := &views.PipelinePage{
			BasePage:  createBasePage(r),
			Pipelines: pipelines.Data,
		}
		if pipelines.Links.GetNext() != "" {
			page.PipelinesNextPage = views.U("/pipelines/table?cursor=%s", getCursor(pipelines.Links.GetNext()))
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

		allWorkers := getPipelineWorkers(r.Context())
		if allWorkers == nil {
			layout.WithSnackbarError(w, "Couldn't find workers")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var dto api.CreatePipelineRequest
		dto.SetDescription(r.FormValue("pipeline-descr"))
		dto.SetSteps(stepsFromWorkersList(allWorkers))

		_, resp, err := h.coreClient.PipelinesApi.CreatePipeline(r.Context()).CreatePipelineRequest(dto).Execute()
		if err != nil {
			web.HTTPError(w, fmt.Errorf("could not create pipeline: %w", err))
			return
		}

		if resp.StatusCode != http.StatusCreated {
			responseBody, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("in createPipeline, err reading response body: %s\n", err)
			} else {
				log.Printf("in createPipeline, err: %s\n", string(responseBody))
			}
			layout.SnackbarSomethingWentWrong(w)
			return
		}

		w.Header().Set("HX-Redirect", views.U("/pipelines"))
		w.WriteHeader(http.StatusOK)
	}
}

func pipelineCreatePage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workers := getWorkers(r.Context())
		if workers == nil {
			web.HTTPResponse(w, http.StatusBadRequest, "invalid workers array")
			return
		}
		workersCursor := getWorkersCursor(r.Context())
		if workersCursor == nil {
			web.HTTPResponse(w, http.StatusBadRequest, "no cursor found")
			return
		}
		page := &views.PipelineEditPage{
			BasePage:          createBasePage(r),
			Pipeline:          nil,
			Workers:           workers,
			WorkersInPipeline: nil,
			WorkersNextPage:   *workersCursor,
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
		pipeline := getPipeline(r.Context())
		if pipeline == nil {
			web.HTTPResponse(w, http.StatusBadRequest, "invalid pipeline model")
			return
		}
		workers := getWorkers(r.Context())
		if workers == nil {
			web.HTTPResponse(w, http.StatusBadRequest, "invalid workers array")
			return
		}
		workersCursor := getWorkersCursor(r.Context())
		if workersCursor == nil {
			web.HTTPResponse(w, http.StatusBadRequest, "no cursor found")
			return
		}
		workersInPipeline := getPipelineWorkers(r.Context())
		if workersInPipeline == nil {
			web.HTTPResponse(w, http.StatusBadRequest, "no workers in pipeline found")
			return
		}

		page := &views.PipelineEditPage{
			BasePage:          createBasePage(r),
			Pipeline:          pipeline,
			Workers:           workers,
			WorkersInPipeline: &workersInPipeline,
			WorkersNextPage:   *workersCursor,
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
			layout.WithSnackbarError(w, "Pipeline must be given")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		pipelineDescr, ok := r.Form["pipeline-descr"]
		if !ok {
			layout.WithSnackbarError(w, "Pipeline description cannot be empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if len(pipelineDescr) != 1 {
			layout.WithSnackbarError(w, "Invalid request")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if pipelineDescr[0] == "" {
			layout.WithSnackbarError(w, "Pipeline description cannot be emptry")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		allWorkers := getPipelineWorkers(r.Context())
		if allWorkers == nil {
			layout.WithSnackbarError(w, "Couldn't find workers")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var updateDto api.UpdatePipelineRequest
		updateDto.SetDescription(pipelineDescr[0])
		updateDto.SetSteps(stepsFromWorkersList(allWorkers))

		_, resp, err := h.coreClient.PipelinesApi.UpdatePipeline(r.Context(), pipelineId).UpdatePipelineRequest(updateDto).Execute()
		if err != nil {
			responseBody, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("in createPipeline, err reading response body: %s\n", err)
			} else {
				log.Printf("in createPipeline, err: %s\n", string(responseBody))
			}
			layout.SnackbarSomethingWentWrong(w)
			return
		}

		// TODO: API returns status created instead of found for some reason
		if resp.StatusCode != http.StatusCreated {
			if resp.StatusCode == http.StatusInternalServerError {
				responseBody, err := io.ReadAll(resp.Body)
				if err != nil {
					log.Printf("in createPipeline, err reading response body: %s\n", err)
				} else {
					log.Printf("in createPipeline, err: %s\n", string(responseBody))
				}
				layout.SnackbarSomethingWentWrong(w)
			} else {
				var apierror *web.APIError
				if errors.As(err, &apierror) {
					layout.WithSnackbarError(w, apierror.Message)
					w.WriteHeader(apierror.HTTPStatus)
					return
				}
			}
			return
		}

		layout.SnackbarSaveSuccessful(w)
		next.ServeHTTP(w, r)
	})
}

func (h *PipelinePageHandler) pipelineStepsView() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allWorkers := getPipelineWorkers(r.Context())
		if allWorkers == nil {
			web.HTTPResponse(w, http.StatusBadRequest, "Couldn't find workers")
			return
		}

		if isHX(r) {
			views.WriteRenderPipelineStepsSortable(w, nosurf.Token(r), allWorkers)
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
			layout.WithSnackbarError(w, "No steps provided")
			return
		}

		if len(steps) == 0 {
			layout.WithSnackbarError(w, "expected single step")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		steps = fixStorageStep(steps)

		allWorkers, err := h.getWorkersForSteps(r, steps)
		if err != nil {
			log.Printf("failed to get workers for step: %s\n", err.Error())
			layout.WithSnackbarError(w, "Failed to get workers for step")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		r = r.WithContext(
			context.WithValue(
				r.Context(),
				ctxPipelineWorkers,
				allWorkers,
			),
		)

		next.ServeHTTP(w, r)
	})
}

func fixStorageStep(steps []string) []string {
	steps = lo.Filter(steps, func(item string, index int) bool {
		return item != STORAGE_STEP
	})
	steps = append(steps, STORAGE_STEP)
	return steps
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
		req := h.workersClient.WorkersApi.ListWorkers(r.Context())
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
		req := h.coreClient.PipelinesApi.ListPipelines(r.Context())
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
			nextCursor = views.U("/pipelines/table?cursor=%s", getCursor(res.Links.GetNext()))
		}
		views.WriteRenderPipelineTableRows(w, res.Data, nextCursor)
	}
}

func (h *PipelinePageHandler) resolvePipeline(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pipeline, resp, err := h.coreClient.PipelinesApi.GetPipeline(r.Context(), chi.URLParam(r, "pipeline_id")).Execute()
		if err != nil {
			web.HTTPError(w, fmt.Errorf("could not get pipeline by id: %w", err))
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
				ctxPipeline,
				pipeline.Data,
			),
		)
		next.ServeHTTP(w, r)
	})
}

func (h *PipelinePageHandler) resolveWorkersInPipeline(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pipeline := getPipeline(r.Context())
		if pipeline == nil {
			web.HTTPResponse(w, http.StatusBadRequest, "invalid pipeline model")
			return
		}

		workersInPipeline, err := h.getWorkersForSteps(r, pipeline.Steps)
		if err != nil {
			web.HTTPError(w, fmt.Errorf("could not get workers in steps: %w", err))
			return
		}

		r = r.WithContext(
			context.WithValue(
				r.Context(),
				ctxPipelineWorkers,
				workersInPipeline,
			),
		)
		next.ServeHTTP(w, r)
	})
}

func (h *PipelinePageHandler) resolveWorkers(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		workers, resp, err := h.workersClient.WorkersApi.ListWorkers(r.Context()).Cursor("").Execute()
		if err != nil {
			web.HTTPError(w, fmt.Errorf("could not list workers: %w", err))
			return
		}
		if resp.StatusCode != http.StatusOK {
			web.HTTPResponse(w, resp.StatusCode, "status not ok")
			return
		}
		// Include old workers or some arbitrary stuff
		ctx := context.WithValue(
			r.Context(),
			ctxWorkers,
			workers.Data,
		)

		nextCursor := ""
		if workers.Links.GetNext() != "" {
			nextCursor = views.U("/pipelines/workers/table?cursor=%s", getCursor(workers.Links.GetNext()))
		}

		ctx = context.WithValue(
			ctx,
			ctxWorkersCursor,
			&nextCursor,
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

func (h *PipelinePageHandler) getWorkersForSteps(r *http.Request, steps []string) ([]api.UserWorker, error) {
	if len(steps) == 0 {
		return []api.UserWorker{}, nil
	}
	// Try and fetch all workers from user-workers service.
	// For any worker not found create a "placeholder" worker
	res, _, err := h.workersClient.WorkersApi.ListWorkers(r.Context()).Id(steps).Execute()
	if err != nil {
		return nil, err
	}
	workersIDs := lo.Map(res.Data, func(w api.UserWorker, _ int) string { return w.GetId() })
	missingWorkers, _ := lo.Difference(steps, workersIDs)
	workers := res.GetData()
	workers = append(workers, createPlaceholderWorkers(missingWorkers)...)

	fmt.Printf("workers: %v\n", workers)
	fmt.Printf("steps: %v\n", steps)
	if len(workers) != len(steps) {
		return nil, fmt.Errorf("some pipeline workers not found")
	}

	workersByID := lo.SliceToMap(workers, func(w api.UserWorker) (string, api.UserWorker) {
		return w.GetId(), w
	})

	// Now that we have the full result we need to make sure the order of the workers is the same as the order of the pipeline steps
	orderedWorkers := lo.Map(steps, func(step string, _ int) api.UserWorker {
		return workersByID[step]
	})

	return orderedWorkers, nil
}

// This function returns any steps that match an image worker to as an API userworker
func createPlaceholderWorkers(steps []string) []api.UserWorker {
	return lo.Map(steps, func(step string, _ int) api.UserWorker {
		return placeholderWorker(step, step)
	})
}

func placeholderWorker(name string, description string) api.UserWorker {
	return api.UserWorker{
		Id:          name,
		Name:        name,
		Revision:    0,
		Description: description,
	}
}
