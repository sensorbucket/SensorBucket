package service

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	pipelineService "sensorbucket.nl/sensorbucket/services/pipeline/service"
)

var (
	ErrInvalidUUID = web.NewError(
		http.StatusBadRequest,
		"Invalid pipeline UUID provided",
		"ERR_PIPELINE_UUID_INVALID",
	)
)

type MessageQueue interface {
	Publish(*pipeline.Message) error
}
type PipelineService interface {
	Get(string) (*pipelineService.Pipeline, error)
}
type HTTPImporter struct {
	router   chi.Router
	pipeline PipelineService
	queue    MessageQueue
}

func New(queue MessageQueue, pipeline PipelineService) *HTTPImporter {
	svc := &HTTPImporter{
		router:   chi.NewRouter(),
		pipeline: pipeline,
		queue:    queue,
	}
	svc.router.Post("/{uuid}", svc.httpPostUplink())
	return svc
}

func (h HTTPImporter) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(rw, r)
}

func (h *HTTPImporter) httpPostUplink() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		pipelineID, err := uuid.Parse(chi.URLParam(r, "uuid"))
		if err != nil {
			web.HTTPError(rw, ErrInvalidUUID)
			return
		}

		payload, err := io.ReadAll(r.Body)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}

		pipelineModel, err := h.pipeline.Get(pipelineID.String())
		if err != nil {
			web.HTTPError(rw, err)
			return
		}
		steps := pipelineModel.Steps

		msg := pipeline.NewMessage(pipelineID.String(), steps)
		msg.Payload = payload

		if err := h.queue.Publish(msg); err != nil {
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusAccepted, &web.APIResponseAny{
			Message: "Received uplink message",
			Data:    msg.ID,
		})
	}
}
