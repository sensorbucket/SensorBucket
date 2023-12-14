package routes

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/api"
	"sensorbucket.nl/sensorbucket/services/dashboard/views"
	"sensorbucket.nl/sensorbucket/services/tracing/tracing"
)

type IngressPageHandler struct {
	router chi.Router
	client *api.APIClient
}

func CreateIngressPageHandler(client *api.APIClient) *IngressPageHandler {
	handler := &IngressPageHandler{
		router: chi.NewRouter(),
		client: client,
	}
	handler.SetupRoutes(handler.router)
	return handler
}

func (h IngressPageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *IngressPageHandler) SetupRoutes(r chi.Router) {
	r.Get("/", h.ingressListPage())
	r.Get("/list", h.ingressListPartial())
}

func (h *IngressPageHandler) createViewIngresses(ctx context.Context) ([]views.Ingress, error) {
	resIngresses, _, err := h.client.TracingApi.ListIngresses(ctx).Limit(30).Execute()
	if err != nil {
		return nil, fmt.Errorf("error listing ingresses: %w", err)
	}
	resPipelines, _, err := h.client.PipelinesApi.ListPipelines(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("error listing pipelines: %w", err)
	}

	plSteps := lo.FlatMap(resPipelines.Data, func(p api.Pipeline, _ int) []string { return p.Steps })
	plSteps = lo.Uniq(plSteps)
	resWorkers, _, err := h.client.WorkersApi.ListWorkers(ctx).Id(plSteps).Execute()
	if err != nil {
		return nil, fmt.Errorf("error listing workers: %w", err)
	}
	workerNames := lo.SliceToMap(resWorkers.Data, func(w api.UserWorker) (string, string) {
		return w.GetId(), w.GetName()
	})

	traceIDs := lo.Map(resIngresses.Data, func(ing api.ArchivedIngress, _ int) string { return ing.GetTracingId() })
	traceIDs = lo.Uniq(traceIDs)
	resLogs, _, err := h.client.TracingApi.ListTraces(ctx).TracingId(traceIDs).Execute()
	if err != nil {
		return nil, fmt.Errorf("error listing traces: %w", err)
	}
	traceMap := lo.SliceToMap(resLogs.Data, func(steplog api.Trace) (string, api.Trace) {
		return steplog.TracingId, steplog
	})

	deviceIDs := lo.FilterMap(resLogs.Data, func(traceLog api.Trace, _ int) (int64, bool) {
		return traceLog.DeviceId, traceLog.DeviceId > 0
	})
	resDevices, _, err := h.client.DevicesApi.ListDevices(ctx).Id(lo.Uniq(deviceIDs)).Execute()
	if err != nil {
		return nil, fmt.Errorf("error listing devices: %w", err)
	}
	deviceMap := lo.SliceToMap(resDevices.Data, func(device api.Device) (int64, api.Device) {
		return device.Id, device
	})

	ingresses := make([]views.Ingress, 0, len(resIngresses.Data))
	for _, ingress := range resIngresses.Data {
		if ingress.IngressDto == nil {
			continue
		}
		pl, found := lo.Find(resPipelines.Data, func(pl api.Pipeline) bool {
			return pl.Id == ingress.IngressDto.PipelineId
		})
		if !found {
			continue
		}
		traceLog, ok := traceMap[ingress.TracingId]
		if !ok {
			log.Printf("warning: could not find trace for archived ingres: %s\n", ingress.TracingId)
			continue
		}

		viewSteps := []views.IngressStep{
			{
				Label:   "Information not available",
				Tooltip: "The pipeline was modified after this message was received, information is not available",
				Status:  int(tracing.Unknown),
			},
		}
		if len(traceLog.Steps) != len(pl.Steps) {
			// The default viewSteps array above already has an "error" set by default.
			// If we can show all the steps, then the whole array is overwritten. See the next "else"
			log.Printf(
				"warning: pipeline has %d steps, but log only has %d. Pipeline has probably been modified after this ingress. Showing no steps...\n",
				len(pl.Steps),
				len(traceLog.Steps),
			)
		} else {
			// TODO: This currently requires that there are an equal number of StepDTO's and Pipeline Steps
			// In the future pipelines will have revisions and are not directly mutable, thus this should always be equal
			viewSteps = lo.Map(pl.Steps, func(stepKey string, ix int) views.IngressStep {
				step := traceLog.Steps[ix]
				stepName := stepKey
				if workerName, ok := workerNames[stepName]; ok {
					stepName = workerName
				}
				viewStep := views.IngressStep{
					Label:  stepName,
					Status: int(step.Status),
				}
				if step.Error != "" {
					viewStep.Tooltip = step.Error
				} else if step.Duration != 0 {
					viewStep.Tooltip = time.Duration(step.Duration * float64(time.Second)).String()
				} else if step.Status == 3 || viewStep.Status == 4 {
					viewStep.Tooltip = "<1s"
				}
				return viewStep
			})
		}

		ingress := views.Ingress{
			TracingID: ingress.TracingId,
			CreatedAt: ingress.IngressDto.CreatedAt,
			Steps:     viewSteps,
		}
		if traceLog.DeviceId != 0 {
			ingress.Device = deviceMap[traceLog.DeviceId]
		}
		ingresses = append(ingresses, ingress)
	}
	return ingresses, nil
}

func (h *IngressPageHandler) ingressListPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ingresses, err := h.createViewIngresses(r.Context())
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		page := &views.IngressPage{Ingresses: ingresses}
		if isHX(r) {
			page.WriteBody(w)
			return
		}
		views.WriteIndex(w, page)
	}
}

func (h *IngressPageHandler) ingressListPartial() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ingresses, err := h.createViewIngresses(r.Context())
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		views.WriteRenderIngressList(w, ingresses)
	}
}
