package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
)

var (
	HTTP_ADDR = env.Could("HTTP_ADDR", ":3000")
	AMQP_URL  = env.Must("AMQP_URL")
	AMQP_XCHG = env.Must("AMQP_XCHG")

	HARDCODED_PIPELINE_STEPS = []string{
		"sensorbucket/ttn-worker@0.0.1",
		"sensorbucket/mfm-worker@0.0.1",
		"service.measurements",
	}

	ErrInvalidUUID = web.NewError(
		http.StatusBadRequest,
		"Invalid pipeline UUID provided",
		"ERR_PIPELINE_UUID_INVALID",
	)
)

func main() {
	if err := Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v", err)
	}
}

func Run() error {
	// Setup AMQP Producer
	xchg := mq.NewAMQPPublisher(AMQP_URL, AMQP_XCHG, func(c *amqp.Channel) error {
		return c.ExchangeDeclare(AMQP_XCHG, "topic", true, false, false, false, nil)
	})
	go xchg.Start()
	log.Printf("AMQP Publisher started...\n")

	// Setup HTTP
	router := chi.NewRouter()
	router.Post("/{uuid}", httpPostUplink(xchg))
	srv := &http.Server{
		Addr:         HTTP_ADDR,
		Handler:      router,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}

	log.Printf("HTTP Server listening on: %s\n", srv.Addr)
	return srv.ListenAndServe()
}

func httpPostUplink(xchg *mq.AMQPPublisher) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		pipelineID, err := uuid.Parse(chi.URLParam(r, "uuid"))
		if err != nil {
			web.HTTPError(rw, ErrInvalidUUID)
			return
		}

		payload, err := io.ReadAll(r.Body)
		if err != nil {
			web.HTTPError(rw, nil)
			return
		}

		// TODO: Fetch pipeline steps based on pipeline UUID from pipeline service
		// instead of having them hardcoded like here
		msg := pipeline.NewMessage(pipelineID.String(), HARDCODED_PIPELINE_STEPS)
		msg.SetPayload(payload)
		step, err := msg.NextStep()
		if err != nil {
			web.HTTPError(rw, err)
			return
		}

		msgData, _ := json.Marshal(&msg)
		xchg.Publish(step, amqp.Publishing{Body: msgData})

		web.HTTPResponse(rw, http.StatusAccepted, &web.APIResponse{
			Message: "Received uplink message",
		})
	}
}
