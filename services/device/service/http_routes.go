package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"sensorbucket.nl/sensorbucket/internal/httpfilter"
	"sensorbucket.nl/sensorbucket/internal/web"
)

var (
	ErrHTTPDeviceIDInvalid = web.NewError(
		http.StatusBadRequest,
		"Device ID must be an integer",
		"DEVICE_ID_INVALID",
	)
)

type middleware = func(next http.Handler) http.Handler

// HTTPTransport ...
type HTTPTransport struct {
	svc    Service
	router chi.Router
}

func NewHTTPTransport(svc Service) *HTTPTransport {
	transport := &HTTPTransport{
		svc:    svc,
		router: chi.NewRouter(),
	}

	// Register endpoints
	transport.setupRoutes()

	return transport
}

func (t *HTTPTransport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.router.ServeHTTP(w, r)
}

// setupRoutes creates router for the user setupRoutes
func (t *HTTPTransport) setupRoutes() {
	r := t.router

	r.Get("/devices", t.httpListDevices())
	r.Post("/devices", t.httpCreateDevice())
	r.Route("/devices/{device_id}", func(r chi.Router) {
		r.Use(t.useDeviceResolver())
		r.Get("/", t.httpGetDevice())
		r.Patch("/", t.httpUpdateDevice())
		r.Delete("/", t.httpDeleteDevice())

		r.Route("/sensors", func(r chi.Router) {
			r.Get("/", t.httpListSensors())
			r.Post("/", t.httpAddSensor())
			r.Delete("/{sensor_code}", t.httpDeleteSensor())
		})
	})
	// TODO: Should we be able to fetch sensor by global unique ID?
}

// Routes
type ListFilters struct {
	North      []float64
	West       []float64
	South      []float64
	East       []float64
	Latitude   []float64
	Longitude  []float64
	Distance   []float64
	Properties httpfilter.Bytes
}

func (f *ListFilters) boundingBox() (BoundingBox, bool) {
	var bb BoundingBox
	if len(f.North) == 0 || len(f.West) == 0 || len(f.South) == 0 || len(f.East) == 0 {
		return bb, false
	}
	bb.North = f.North[0]
	bb.West = f.West[0]
	bb.East = f.East[0]
	bb.South = f.South[0]
	return bb, true
}

func (f *ListFilters) locationRange() (LocationRange, bool) {
	var lr LocationRange
	if len(f.Latitude) == 0 || len(f.Longitude) == 0 || len(f.Distance) == 0 {
		return lr, false
	}
	lr.Latitude = f.Latitude[0]
	lr.Longitude = f.Longitude[0]
	lr.Distance = f.Distance[0]
	return lr, true
}

func (f *ListFilters) filters() DeviceFilter {
	return DeviceFilter{
		json.RawMessage(f.Properties),
	}
}

func (t *HTTPTransport) httpListDevices() http.HandlerFunc {
	parseFilter := httpfilter.MustCreate[ListFilters]()
	return func(rw http.ResponseWriter, r *http.Request) {
		var filter ListFilters
		if err := parseFilter(r.URL.Query(), &filter); err != nil {
			web.HTTPError(rw, err)
			return
		}
		fmt.Printf("%+v\n", filter)
		var devices []Device

		// figure out what kind of query this is
		var err error
		if lr, ok := filter.locationRange(); ok {
			devices, err = t.svc.ListInRange(r.Context(), lr, filter.filters())
		} else if bb, ok := filter.boundingBox(); ok {
			devices, err = t.svc.ListInBoundingBox(r.Context(), bb, filter.filters())
		} else {
			devices, err = t.svc.ListDevices(r.Context(), filter.filters())
		}
		if err != nil {
			web.HTTPError(rw, err)
			return
		}
		web.HTTPResponse(rw, http.StatusOK, &web.APIResponseAny{
			Message: "listed devices",
			Data:    devices,
		})
	}
}

func (t *HTTPTransport) httpGetDevice() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		device := r.Context().Value(ctxDeviceKey).(*Device)
		web.HTTPResponse(rw, http.StatusOK, &web.APIResponseAny{
			Message: "Fetched device",
			Data:    device,
		})
	}
}

func (t *HTTPTransport) httpCreateDevice() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var req NewDeviceOpts
		if err := web.DecodeJSON(r, &req); err != nil {
			web.HTTPError(rw, err)
			return
		}

		dev, err := t.svc.CreateDevice(r.Context(), req)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusCreated, &web.APIResponseAny{
			Message: "Created new device",
			Data:    dev,
		})
	}
}

func (t *HTTPTransport) httpDeleteDevice() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		dev := r.Context().Value(ctxDeviceKey).(*Device)

		if err := t.svc.DeleteDevice(r.Context(), dev); err != nil {
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, &web.APIResponseAny{
			Message: "Deleted device",
		})
	}
}

func (t *HTTPTransport) httpUpdateDevice() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		dev := r.Context().Value(ctxDeviceKey).(*Device)

		var dto UpdateDeviceOpts
		if err := web.DecodeJSON(r, &dto); err != nil {
			web.HTTPError(rw, err)
			return
		}

		if err := t.svc.UpdateDevice(r.Context(), dev, dto); err != nil {
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, &web.APIResponseAny{
			Message: "Updated device",
		})
	}
}

func (t *HTTPTransport) httpListSensors() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		device := r.Context().Value(ctxDeviceKey).(*Device)

		web.HTTPResponse(rw, http.StatusOK, &web.APIResponseAny{
			Message: "Listed sensors",
			Data:    device.Sensors,
		})
	}
}

func (t *HTTPTransport) httpAddSensor() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		dev := r.Context().Value(ctxDeviceKey).(*Device)

		var dto NewSensorDTO
		if err := web.DecodeJSON(r, &dto); err != nil {
			web.HTTPError(rw, err)
			return
		}

		if err := t.svc.AddSensor(r.Context(), dev, dto); err != nil {
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusCreated, &web.APIResponseAny{
			Message: "Created new sensor for device",
		})
	}
}

func (t *HTTPTransport) httpDeleteSensor() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		dev := r.Context().Value(ctxDeviceKey).(*Device)

		sensor, err := dev.GetSensorByCode(chi.URLParam(r, "sensor_code"))
		if err != nil {
			web.HTTPError(rw, err)
			return
		}

		if err := t.svc.DeleteSensor(r.Context(), dev, sensor); err != nil {
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, &web.APIResponseAny{
			Message: "Deleted sensor from device",
		})
	}
}

//
// Helpers
//

var ctxDeviceKey = struct{}{}

func (t *HTTPTransport) useDeviceResolver() middleware {
	return func(next http.Handler) http.Handler {
		mw := func(rw http.ResponseWriter, r *http.Request) {
			idString := chi.URLParam(r, "device_id")
			id, err := strconv.ParseInt(idString, 10, 64)
			if err != nil {
				web.HTTPError(rw, ErrHTTPDeviceIDInvalid)
				return
			}

			dev, err := t.svc.GetDevice(r.Context(), id)
			if err != nil {
				web.HTTPError(rw, err)
				return
			}

			r = r.WithContext(context.WithValue(
				r.Context(),
				ctxDeviceKey,
				dev,
			))

			next.ServeHTTP(rw, r)
		}
		return http.HandlerFunc(mw)
	}
}
