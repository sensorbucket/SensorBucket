package coretransport

//go:generate moq -pkg coretransport_test -out mock_test.go . MeasurementService

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/services/core/devices"
	"sensorbucket.nl/sensorbucket/services/core/measurements"
	"sensorbucket.nl/sensorbucket/services/core/processing"
)

type MeasurementService interface {
	QueryMeasurements(context.Context, measurements.Filter, pagination.Request) (*pagination.Page[measurements.Measurement], error)
	GetDatastream(context.Context, uuid.UUID) (*measurements.Datastream, error)
	ListDatastreams(context.Context, measurements.DatastreamFilter, pagination.Request) (*pagination.Page[measurements.Datastream], error)
}

type CoreTransport struct {
	baseURL            string
	router             chi.Router
	keySource          auth.JWKSClient
	deviceService      *devices.Service
	measurementService MeasurementService
	processingService  *processing.Service
}

func New(
	baseURL string,
	keySource auth.JWKSClient,
	deviceService *devices.Service,
	measurementService MeasurementService,
	processingService *processing.Service,
) *CoreTransport {
	t := &CoreTransport{
		baseURL:            baseURL,
		keySource:          keySource,
		deviceService:      deviceService,
		measurementService: measurementService,
		processingService:  processingService,
	}
	t.routes()
	return t
}

func (t CoreTransport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.router.ServeHTTP(w, r)
}

func (t *CoreTransport) routes() {
	r := chi.NewRouter()
	t.router = r

	r.Use(
		chimw.Logger,
		auth.Authenticate(t.keySource),
		auth.Protect(),
	)

	r.Get("/devices", t.httpListDevices())
	r.Post("/devices", t.httpCreateDevice())
	r.Route("/devices/{device_id}", func(r chi.Router) {
		r.Use(t.useDeviceResolver())
		r.Get("/", t.httpGetDevice())
		r.Patch("/", t.httpUpdateDevice())
		r.Delete("/", t.httpDeleteDevice())

		r.Route("/sensors", func(r chi.Router) {
			r.Get("/", t.httpListDeviceSensors())
			r.Post("/", t.httpAddSensor())
			r.Route("/{sensor_code}", func(r chi.Router) {
				r.Use(t.useSensorResolver())
				r.Get("/", t.httpGetSensor())
				r.Delete("/", t.httpDeleteSensor())
				r.Patch("/", t.httpUpdateSensor())
			})
		})
	})

	r.Get("/sensors", t.httpListSensors())
	r.Route("/sensors/{sensor_id}", func(r chi.Router) {
		r.Get("/", t.httpGetSensor())
	})
	r.Route("/sensor-groups", func(r chi.Router) {
		r.Post("/", t.httpCreateSensorGroup())
		r.Get("/", t.httpListSensorGroups())
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", t.httpGetSensorGroup())
			r.Delete("/", t.httpDeleteSensorGroup())
			r.Patch("/", t.httpUpdateSensorGroup())
			r.Post("/sensors", t.httpAddSensorToSensorGroup())
			r.Delete("/sensors/{sid}", t.httpDeleteSensorFromSensorGroup())
		})
	})

	r.Route("/datastreams", func(r chi.Router) {
		r.Get("/", t.httpListDatastream())
		r.Get("/{id}", t.httpGetDatastream())
	})

	r.Route("/pipelines", func(r chi.Router) {
		r.Post("/", t.httpCreatePipeline())
		r.Get("/", t.httpListPipelines())
		r.Get("/{id}", t.httpGetPipeline())
		r.Patch("/{id}", t.httpUpdatePipeline())
		r.Delete("/{id}", t.httpDeletePipeline())
	})

	r.Get("/measurements", t.httpGetMeasurements())
}
