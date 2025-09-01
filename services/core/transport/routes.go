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
	"sensorbucket.nl/sensorbucket/services/core/featuresofinterest"
	"sensorbucket.nl/sensorbucket/services/core/measurements"
	"sensorbucket.nl/sensorbucket/services/core/processing"
	"sensorbucket.nl/sensorbucket/services/core/projects"
)

// var logger = slog.Default().With("component", "services/core/transport")

type MeasurementService interface {
	QueryMeasurements(
		context.Context,
		measurements.Filter,
		pagination.Request,
	) (*pagination.Page[measurements.Measurement], error)
	GetDatastream(context.Context, uuid.UUID) (*measurements.Datastream, error)
	ListDatastreams(
		context.Context,
		measurements.DatastreamFilter,
		pagination.Request,
	) (*pagination.Page[measurements.Datastream], error)
}

type CoreTransport struct {
	baseURL                  string
	router                   chi.Router
	keySource                auth.JWKSClient
	deviceService            *devices.Service
	measurementService       MeasurementService
	processingService        *processing.Service
	projectsService          *projects.Application
	featureOfInterestService *featuresofinterest.Service
}

func New(
	baseURL string,
	keySource auth.JWKSClient,
	deviceService *devices.Service,
	measurementService MeasurementService,
	processingService *processing.Service,
	projectsService *projects.Application,
	featureOfInterestService *featuresofinterest.Service,
) *CoreTransport {
	t := &CoreTransport{
		baseURL:                  baseURL,
		keySource:                keySource,
		deviceService:            deviceService,
		measurementService:       measurementService,
		processingService:        processingService,
		projectsService:          projectsService,
		featureOfInterestService: featureOfInterestService,
	}
	t.routes()
	return t
}

func (transport CoreTransport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	transport.router.ServeHTTP(w, r)
}

func (transport *CoreTransport) routes() {
	r := chi.NewRouter()
	transport.router = r

	r.Use(
		chimw.Logger,
		auth.Authenticate(transport.keySource),
		auth.Protect(),
	)

	r.Get("/devices", transport.httpListDevices())
	r.Post("/devices", transport.httpCreateDevice())
	r.Route("/devices/{device_id}", func(r chi.Router) {
		r.Use(transport.useDeviceResolver())
		r.Get("/", transport.httpGetDevice())
		r.Patch("/", transport.httpUpdateDevice())
		r.Delete("/", transport.httpDeleteDevice())

		r.Route("/sensors", func(r chi.Router) {
			r.Get("/", transport.httpListDeviceSensors())
			r.Post("/", transport.httpAddSensor())
			r.Route("/{sensor_code}", func(r chi.Router) {
				r.Use(transport.useSensorResolver())
				r.Get("/", transport.httpGetSensor())
				r.Delete("/", transport.httpDeleteSensor())
				r.Patch("/", transport.httpUpdateSensor())
			})
		})
	})

	r.Get("/sensors", transport.httpListSensors())
	r.Route("/sensors/{sensor_id}", func(r chi.Router) {
		r.Get("/", transport.httpGetSensor())
	})
	r.Route("/sensor-groups", func(r chi.Router) {
		r.Post("/", transport.httpCreateSensorGroup())
		r.Get("/", transport.httpListSensorGroups())
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", transport.httpGetSensorGroup())
			r.Delete("/", transport.httpDeleteSensorGroup())
			r.Patch("/", transport.httpUpdateSensorGroup())
			r.Post("/sensors", transport.httpAddSensorToSensorGroup())
			r.Delete("/sensors/{sid}", transport.httpDeleteSensorFromSensorGroup())
		})
	})

	r.Route("/datastreams", func(r chi.Router) {
		r.Get("/", transport.httpListDatastream())
		r.Get("/{id}", transport.httpGetDatastream())
	})

	r.Route("/pipelines", func(r chi.Router) {
		r.Post("/", transport.httpCreatePipeline())
		r.Get("/", transport.httpListPipelines())
		r.Get("/{id}", transport.httpGetPipeline())
		r.Patch("/{id}", transport.httpUpdatePipeline())
		r.Delete("/{id}", transport.httpDeletePipeline())
	})

	r.Route("/projects", func(r chi.Router) {
		r.Get("/", transport.httpListProjects())
	})

	r.Route("/features-of-interest", func(r chi.Router) {
		r.Get("/", transport.httpListFeaturesOfInterest())
		r.Post("/", transport.httpCreateFeatureOfInterest())
		r.Get("/{id}", transport.httpGetFeatureOfInterest())
		r.Delete("/{id}", transport.httpDeleteFeaturOfInterest())
		r.Patch("/{id}", transport.httpUpdateFeatureOfInterest())
	})

	r.Get("/measurements", transport.httpGetMeasurements())
}
