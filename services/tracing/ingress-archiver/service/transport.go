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

func StartIngressDTOConsumer(conn *mq.AMQPConnection, svc *Application, queue, xchg, topic string) {
	consume := conn.Consume(queue, func(c *amqp091.Channel) error {
		_, err := c.QueueDeclare(queue, true, false, false, false, nil)
		if err != nil {
			return err
		}

		// Create exchange and bind if both arguments are provided, this is optional
		if xchg != "" && topic != "" {
			if err := c.ExchangeDeclare(xchg, "topic", true, false, false, false, nil); err != nil {
				return err
			}
			if err := c.QueueBind(queue, topic, xchg, false, nil); err != nil {
				return err
			}
		}
		return nil
	})

	for delivery := range consume {
		tracingID, err := uuid.Parse(delivery.MessageId)
		if err != nil {
			fmt.Printf("Delivery TracingID is not a UUID (%v)\n", err.Error())
			tracingID = uuid.UUID{}
		}
		rawMessage := delivery.Body
		if err := svc.ArchiveIngressDTO(tracingID, rawMessage); err != nil {
			fmt.Printf("Error processing ingress DTO: %v\n", err)
			delivery.Nack(false, false)
			continue
		}
		delivery.Ack(false)
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
