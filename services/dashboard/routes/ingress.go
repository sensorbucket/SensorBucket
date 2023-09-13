package routes

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/api"
	"sensorbucket.nl/sensorbucket/services/dashboard/views"
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
	resIngresses, _, err := h.client.TracingApi.ListIngresses(ctx).Execute()
	if err != nil {
		return nil, err
	}
	pipelineIDs := lo.FilterMap(resIngresses.Data, func(ingr api.ArchivedIngress, _ int) (string, bool) {
		if ingr.IngressDto == nil {
			return "", false
		}
		return ingr.IngressDto.GetPipelineId(), true
	})
	pipelineIDs = lo.Uniq(pipelineIDs)
	resPipelines, _, err := h.client.PipelinesApi.ListPipelines(ctx).Execute()
	if err != nil {
		return nil, err
	}

	resLogs, _, err := h.client.TracingApi.ListTraces(ctx).TracingId(lo.Map(resIngresses.Data, func(ing api.ArchivedIngress, _ int) string { return ing.GetTracingId() })).Execute()
	if err != nil {
		return nil, err
	}
	traceMap := lo.SliceToMap(resLogs.Data, func(steplog api.Trace) (string, api.Trace) {
		return steplog.TracingId, steplog
	})

	deviceIDs := lo.FilterMap(resLogs.Data, func(traceLog api.Trace, _ int) (int64, bool) {
		return traceLog.DeviceId, traceLog.DeviceId > 0
	})
	resDevices, _, err := h.client.DevicesApi.ListDevices(ctx).Id(lo.Uniq(deviceIDs)).Execute()
	if err != nil {
		return nil, err
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
		ingress := views.Ingress{
			TracingID: ingress.TracingId,
			CreatedAt: ingress.IngressDto.CreatedAt,
			Steps: lo.Map(pl.Steps, func(stepLabel string, ix int) views.IngressStep {
				// TODO: This currently requires that there are an equal number of StepDTO's and Pipeline Steps
				// In the future pipelines will have revisions and are not directly mutable, thus this should always be equal
				step := traceLog.Steps[ix]
				viewStep := views.IngressStep{
					Label:  stepLabel,
					Status: int(step.Status),
				}
				if step.Error != "" {
					viewStep.Tooltip = step.Error
				} else if step.Duration != 0 {
					viewStep.Tooltip = time.Duration(step.Duration * int32(time.Second)).String()
				} else if step.Status == 3 || viewStep.Status == 4 {
					viewStep.Tooltip = "<1s"
				}
				return viewStep
			}),
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
