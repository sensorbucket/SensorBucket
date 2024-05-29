package service

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/auth"
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

func New(publisher IngressDTOPublisher, keySource auth.JWKSClient) *HTTPImporter {
	svc := &HTTPImporter{
		router:    chi.NewRouter(),
		publisher: publisher,
	}
	svc.router.Use(
		chimw.Logger,
		auth.Authenticate(keySource),
		auth.Protect(),
	)
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

		if err := auth.MustHavePermissions(r.Context(), auth.Permissions{auth.WRITE_MEASUREMENTS}); err != nil {
			web.HTTPError(rw, err)
			return
		}
		tenantID, err := auth.GetTenant(r.Context())
		if err != nil {
			web.HTTPError(rw, err)
			return
		}

		payload, err := io.ReadAll(r.Body)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}

		dto := processing.CreateIngressDTO(pipelineID, tenantID, payload)
		h.publisher <- dto

		web.HTTPResponse(rw, http.StatusAccepted, &web.APIResponseAny{
			Message: "Received uplink message",
			Data:    dto.TracingID.String(),
		})
	}
}
