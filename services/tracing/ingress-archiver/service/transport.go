package ingressarchiver

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"

	"sensorbucket.nl/sensorbucket/internal/httpfilter"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/mq"
)

func MQIngressProcessor(svc *Application) mq.ProcessorFuncBuilder {
	return func() mq.ProcessorFunc {
		return func(delivery amqp091.Delivery) error {
			tracingID, err := uuid.Parse(delivery.MessageId)
			if err != nil {
				fmt.Printf("Delivery TracingID is not a UUID (%v)\n", err.Error())
				tracingID = uuid.UUID{}
			}
			if err := svc.ArchiveIngressDTO(tracingID, delivery.Body); err != nil {
				return fmt.Errorf("processing ingress DTO: %w", err)
			}
			return nil
		}
	}
}

type HTTPIngressesFilter struct {
	ArchiveFilters
	pagination.Request
}

func CreateHTTPTransport(r chi.Router, app *Application) {
	r.Get("/ingresses", func(w http.ResponseWriter, r *http.Request) {
		params, err := httpfilter.Parse[HTTPIngressesFilter](r)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		page, err := app.ListIngresses(r.Context(), params.ArchiveFilters, params.Request)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		for ix := range page.Data {
			page.Data[ix].RawMessage = nil
			if page.Data[ix].IngressDTO != nil {
				page.Data[ix].IngressDTO.Payload = nil
			}
		}

		web.HTTPResponse(w, http.StatusOK, pagination.CreateResponse(r, "", *page))
	})
}
