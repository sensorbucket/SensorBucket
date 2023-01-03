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
	HTTP_ADDR    = env.Could("HTTP_ADDR", ":3000")
	AMQP_URL     = env.Must("AMQP_URL")
	AMQP_XCHG    = env.Must("AMQP_XCHG")
	SVC_PIPELINE = env.Must("SVC_PIPELINE")

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

type Pipeline struct {
	Data struct {
		Steps []string `json:"steps"`
	} `json:"data"`
}

func getPipelineSteps(id string) ([]string, error) {
	res, err := http.Get(fmt.Sprintf("%s/pipelines/%s", SVC_PIPELINE, id))
	if err != nil {
		return nil, fmt.Errorf("could not get pipeline definition: %w", err)
	}
	defer res.Body.Close()

	// if error status, then pipeline service should have responded with APIError
	// We forward that error to the requester.
	if res.StatusCode < 200 || res.StatusCode > 299 {
		var err web.APIError
		if err := json.NewDecoder(res.Body).Decode(&err); err != nil {
			return nil, fmt.Errorf("could not read pipline service response: %w", err)
		}
		err.HTTPStatus = res.StatusCode
		log.Printf("Error status: %v\n", err)
		return nil, &err
	}

    var p Pipeline
	if err := json.NewDecoder(res.Body).Decode(&p); err != nil {
		return nil, fmt.Errorf("could not parse pipeline service response: %w", err)
	}

	return p.Data.Steps, nil
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
			web.HTTPError(rw, err)
			return
		}

		steps, err := getPipelineSteps(pipelineID.String())
		if err != nil {
			web.HTTPError(rw, err)
			return
		}

		msg := pipeline.NewMessage(pipelineID.String(), steps)
		msg.SetPayload(payload)
		step, err := msg.NextStep()
		if err != nil {
			web.HTTPError(rw, err)
			return
		}

		msgData, err := json.Marshal(&msg)
        if err != nil {
            web.HTTPError(rw, err)
            return
        }
		xchg.Publish(step, amqp.Publishing{Body: msgData})

		web.HTTPResponse(rw, http.StatusAccepted, &web.APIResponse{
			Message: "Received uplink message",
		})
	}
}
