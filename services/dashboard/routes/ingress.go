package routes

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/core/devices"
	"sensorbucket.nl/sensorbucket/services/core/processing"
	"sensorbucket.nl/sensorbucket/services/dashboard/views"
	ingressarchiver "sensorbucket.nl/sensorbucket/services/tracing/ingress-archiver/service"
)

type IngressStore interface {
	ListIngresses() ([]ingressarchiver.ArchivedIngressDTO, error)
}

type TraceDTO struct {
	TracingId string    `json:"tracing_id"`
	DeviceID  int64     `json:"device_id"`
	Status    int       `json:"status"`
	Steps     []StepDTO `json:"steps"`
}

type StepDTO struct {
	Status   int           `json:"status"`
	Duration time.Duration `json:"duration"`
	Error    string        `json:"error"`
}

type TracesStore interface {
	ListTraces(tracingIDS []uuid.UUID) ([]TraceDTO, error)
}

type PipelineStore interface {
	ListPipelines(ids []uuid.UUID) ([]processing.Pipeline, error)
}

type DeviceStore interface {
	ListDevices(ids []int64) ([]devices.Device, error)
}

type IngressPageHandler struct {
	router    chi.Router
	ingresses IngressStore
	traces    TracesStore
	pipelines PipelineStore
	devices   DeviceStore
}

func CreateIngressPageHandler(ingresses IngressStore, traces TracesStore, pipelines PipelineStore, devices DeviceStore) *IngressPageHandler {
	handler := &IngressPageHandler{
		router:    chi.NewRouter(),
		ingresses: ingresses,
		traces:    traces,
		pipelines: pipelines,
		devices:   devices,
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

func (h *IngressPageHandler) createViewIngresses() ([]views.Ingress, error) {
	archivedIngresses, err := h.ingresses.ListIngresses()
	if err != nil {
		return nil, err
	}
	pipelineIDs := lo.FilterMap(archivedIngresses, func(ingr ingressarchiver.ArchivedIngressDTO, _ int) (uuid.UUID, bool) {
		if ingr.IngressDTO == nil {
			return uuid.UUID{}, false
		}
		return ingr.IngressDTO.PipelineID, true
	})
	pipelineIDs = lo.Uniq(pipelineIDs)
	pipelines, err := h.pipelines.ListPipelines(pipelineIDs)
	if err != nil {
		return nil, err
	}

	traceLogs, err := h.traces.ListTraces(lo.Map(archivedIngresses, func(ing ingressarchiver.ArchivedIngressDTO, _ int) uuid.UUID { return ing.TracingID }))
	if err != nil {
		return nil, err
	}
	traceMap := lo.SliceToMap(traceLogs, func(steplog TraceDTO) (string, TraceDTO) {
		return steplog.TracingId, steplog
	})

	deviceIDs := lo.FilterMap(traceLogs, func(traceLog TraceDTO, _ int) (int64, bool) {
		return traceLog.DeviceID, traceLog.DeviceID > 0
	})
	deviceList, err := h.devices.ListDevices(deviceIDs)
	if err != nil {
		return nil, err
	}
	deviceMap := lo.SliceToMap(deviceList, func(device devices.Device) (int64, devices.Device) {
		return device.ID, device
	})

	ingresses := make([]views.Ingress, 0, len(archivedIngresses))
	for _, ingress := range archivedIngresses {
		if ingress.IngressDTO == nil {
			continue
		}
		pl, found := lo.Find(pipelines, func(pl processing.Pipeline) bool {
			return pl.ID == ingress.IngressDTO.PipelineID.String()
		})
		if !found {
			continue
		}
		traceLog, ok := traceMap[ingress.TracingID.String()]
		if !ok {
			log.Printf("warning: could not find trace for archived ingres: %s\n", ingress.TracingID.String())
			continue
		}
		ingress := views.Ingress{
			TracingID: ingress.TracingID.String(),
			CreatedAt: ingress.IngressDTO.CreatedAt,
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
					viewStep.Tooltip = step.Duration.String()
				} else if step.Status == 3 || viewStep.Status == 4 {
					viewStep.Tooltip = "<1s"
				}
				return viewStep
			}),
		}
		if traceLog.DeviceID != 0 {
			ingress.Device = deviceMap[traceLog.DeviceID]
		}
		ingresses = append(ingresses, ingress)
	}
	return ingresses, nil
}

func (h *IngressPageHandler) ingressListPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ingresses, err := h.createViewIngresses()
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
		ingresses, err := h.createViewIngresses()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		views.WriteRenderIngressList(w, ingresses)
	}
}
