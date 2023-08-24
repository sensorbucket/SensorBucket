package routes

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/core/processing"
	"sensorbucket.nl/sensorbucket/services/dashboard/views"
	ingressarchiver "sensorbucket.nl/sensorbucket/services/tracing/ingress-archiver/service"
)

type IngressStore interface {
	ListIngresses() ([]ingressarchiver.ArchivedIngressDTO, error)
}

type TraceDTO struct {
	TracingId string    `json:"tracing_id"`
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

type IngressPageHandler struct {
	router    chi.Router
	ingresses IngressStore
	traces    TracesStore
	pipelines PipelineStore
}

func CreateIngressPageHandler(ingresses IngressStore, traces TracesStore, pipelines PipelineStore) *IngressPageHandler {
	handler := &IngressPageHandler{
		router:    chi.NewRouter(),
		ingresses: ingresses,
		traces:    traces,
		pipelines: pipelines,
	}
	handler.SetupRoutes(handler.router)
	return handler
}

func (h IngressPageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *IngressPageHandler) SetupRoutes(r chi.Router) {
	r.Get("/", h.ingressListPage())
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

	steplogs, err := h.traces.ListTraces(lo.Map(archivedIngresses, func(ing ingressarchiver.ArchivedIngressDTO, _ int) uuid.UUID { return ing.TracingID }))
	if err != nil {
		return nil, err
	}
	// TODO: Should mustparse be here?
	traceMap := lo.SliceToMap(steplogs, func(steplog TraceDTO) (uuid.UUID, TraceDTO) {
		return uuid.MustParse(steplog.TracingId), steplog
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
		traceSteps := traceMap[ingress.TracingID]
		ingress := views.Ingress{
			TracingID: ingress.TracingID.String(),
			CreatedAt: ingress.IngressDTO.CreatedAt,
			Steps: lo.Map(pl.Steps, func(stepLabel string, ix int) views.IngressStep {
				// TODO: This currently requires that there are an equal number of StepDTO's and Pipeline Steps
				// In the future pipelines will have revisions and are not directly mutable, thus this should always be equal
				step := traceSteps.Steps[ix]
				return views.IngressStep{
					Label:    stepLabel,
					Status:   int(step.Status),
					Duration: step.Duration.String(),
				}
			}),
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
		views.WriteIndex(w, &views.IngressPage{
			Ingresses: ingresses,
		})
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
