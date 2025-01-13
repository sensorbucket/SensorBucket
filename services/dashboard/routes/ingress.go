package routes

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/api"
	"sensorbucket.nl/sensorbucket/services/dashboard/views"
)

type TracesPageHandler struct {
	router        chi.Router
	coreClient    *api.APIClient
	tracesClient  *api.APIClient
	workersClient *api.APIClient
}

func CreateTracesPageHandler(core, traces, workers *api.APIClient) *TracesPageHandler {
	handler := &TracesPageHandler{
		router:        chi.NewRouter(),
		coreClient:    core,
		tracesClient:  traces,
		workersClient: workers,
	}
	handler.SetupRoutes(handler.router)
	return handler
}

func (h TracesPageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *TracesPageHandler) SetupRoutes(r chi.Router) {
	r.Get("/list", h.listPartial())
}

func (h *TracesPageHandler) listPartial() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := h.tracesClient.TracingApi.ListTraces(r.Context())

		if pipelineIDs, ok := r.URL.Query()["pipeline"]; ok {
			req = req.Pipeline(pipelineIDs)
		}
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			limit, err := strconv.Atoi(limitStr)
			if err != nil {
				web.HTTPError(w, err)
				return
			}
			req = req.Limit(int32(limit))
		}

		traces, _, err := req.Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		viewModels, err := h.createViewData(r.Context(), traces.GetData())
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		views.WriteRenderTracesList(w, viewModels)
	}
}

func formatSince(t time.Time) string {
	d := time.Since(t)
	if d.Hours() > 24 {
		return "More than a day ago"
	}
	if int(d.Hours()) > 1 {
		return fmt.Sprintf("About %d hours ago", int(d.Hours()))
	}
	if int(d.Hours()) > 0 {
		return fmt.Sprintf("About %d hours ago", int(d.Hours()))
	}
	if int(d.Minutes()) > 1 {
		return fmt.Sprintf("About %d minutes ago", int(d.Minutes()))
	}
	if int(d.Minutes()) > 0 {
		return fmt.Sprintf("About %d minute ago", int(d.Minutes()))
	}
	return fmt.Sprintf("About %d seconds ago", int(d.Seconds()))
}

func (h *TracesPageHandler) createViewData(ctx context.Context, traces []api.Trace) ([]views.Trace, error) {
	deviceIDs := lo.Map(traces, func(trace api.Trace, _ int) int64 { return trace.GetDeviceId() })
	pipelineIDs := lo.Map(traces, func(trace api.Trace, _ int) string { return trace.GetPipelineId() })
	workerIDs := lo.FlatMap(traces, func(trace api.Trace, _ int) []string { return trace.GetWorkers() })

	devices, pipelines, workers, err := h.enrichData(ctx, deviceIDs, pipelineIDs, workerIDs)
	if err != nil {
		return nil, err
	}

	viewModels := make([]views.Trace, len(traces))
	for i, trace := range traces {

		viewModels[i] = views.Trace{
			ID:         trace.GetId(),
			StartTime:  trace.GetStartTime(),
			TimeAgo:    formatSince(trace.GetStartTime()),
			PipelineID: pipelineIDs[i],
			DeviceID:   deviceIDs[i],
			Steps:      make([]views.Step, 0, len(trace.GetWorkers())),
		}

		pipeline, ok := pipelines[pipelineIDs[i]]
		if ok {
			viewModels[i].PipelineName = pipeline.GetDescription()
		}
		device, ok := devices[deviceIDs[i]]
		if ok {
			viewModels[i].DeviceCode = device.GetCode()
		}

		// First add all workers from the trace to the viewModel
		workerIndex := 0
		for j, workerID := range trace.GetWorkers() {
			step := views.Step{
				Name:   workerID,
				Status: views.StatusCompleted,
			}
			worker, ok := workers[workerID]
			if ok {
				step.Name = worker.GetName()
			}
			viewModels[i].Steps = append(viewModels[i].Steps, step)

			// update last worker with duration
			if j > 0 {
				viewModels[i].Steps[j-1].Label = trace.WorkerTimes[j].Sub(trace.WorkerTimes[j-1]).String()
			}

			workerIndex++
		}

		// Set the last trace to pending, unless its storage
		if workerIndex > 0 && trace.HasError() {
			viewModels[i].Steps[workerIndex-1].Status = views.StatusError
			viewModels[i].Steps[workerIndex-1].Label = trace.GetError()
		} else if workerIndex > 0 && trace.Workers[workerIndex-1] != "storage" {
			viewModels[i].Steps[workerIndex-1].Status = views.StatusPending
		}

		// Then add all the workers from the pipeline to the model, starting from the last added worker
		for ; workerIndex < len(pipeline.Steps); workerIndex++ {
			workerID := pipeline.Steps[workerIndex]
			step := views.Step{
				Name: workerID,
			}
			worker, ok := workers[workerID]
			if ok {
				step.Name = worker.GetName()
			}
			viewModels[i].Steps = append(viewModels[i].Steps, step)
		}
	}
	return viewModels, nil
}

func (h *TracesPageHandler) enrichData(ctx context.Context, deviceIDs []int64, pipelineIDs, workerIDs []string) (
	devices map[int64]api.Device, pipelines map[string]api.Pipeline, workers map[string]api.UserWorker, err error,
) {
	errs := make([]error, 3)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		res, _, err := h.coreClient.DevicesApi.ListDevices(ctx).Id(deviceIDs).Execute()
		if err != nil {
			errs[0] = err
			return
		}
		devices = lo.KeyBy(res.GetData(), func(d api.Device) int64 { return d.GetId() })
	}()
	go func() {
		defer wg.Done()
		res, _, err := h.coreClient.PipelinesApi.ListPipelines(ctx).Id(pipelineIDs).Execute()
		if err != nil {
			errs[1] = err
			return
		}
		pipelines = lo.KeyBy(res.GetData(), func(p api.Pipeline) string { return p.GetId() })
		workerIDs = append(workerIDs, lo.FlatMap(res.GetData(), func(p api.Pipeline, _ int) []string { return p.GetSteps() })...)

		wRes, _, err := h.workersClient.WorkersApi.ListWorkers(ctx).Id(workerIDs).Execute()
		if err != nil {
			errs[2] = err
			return
		}
		workers = lo.KeyBy(wRes.GetData(), func(w api.UserWorker) string { return w.GetId() })
	}()

	wg.Wait()
	err = errors.Join(errs...)
	return
}
