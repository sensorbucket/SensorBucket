package service

import (
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/core/processing"
)

var ErrInvalidUUID = web.NewError(
	http.StatusBadRequest,
	"Invalid pipeline UUID provided",
	"ERR_PIPELINE_UUID_INVALID",
)

type (
	IngressDTOPublisher chan<- processing.IngressDTO
	HTTPImporter        struct {
		router    chi.Router
		publisher IngressDTOPublisher
	}
)

func New(publisher IngressDTOPublisher) *HTTPImporter {
	svc := &HTTPImporter{
		router:    chi.NewRouter(),
		publisher: publisher,
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

		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			auth = auth[7:]
		} else {
			http.Error(rw, "Missing bearer token in Authorization header", http.StatusUnauthorized)
			return
		}

		dto := processing.CreateIngressDTO(pipelineID, auth, payload)
		h.publisher <- dto

		web.HTTPResponse(rw, http.StatusAccepted, &web.APIResponseAny{
			Message: "Received uplink message",
			Data:    dto.TracingID.String(),
		})
	}
}
