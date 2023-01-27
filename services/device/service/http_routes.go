package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
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
}

//
// Routes
//

func parseQueryFilter(r *http.Request) (DeviceFilter, error) {
	var filter DeviceFilter
	q := r.URL.Query()

	// Configuration filter
	configurationFilter := q.Get("configuration")
	if configurationFilter != "" {
		if err := json.Unmarshal([]byte(configurationFilter), &filter.Configuration); err != nil {
			return filter, err
		}
	}

	return filter, nil
}

func parseBoundingBoxRequest(r *http.Request) (BoundingBox, error) {
	var err error
	var bb BoundingBox
	q := r.URL.Query()

	northQ := q.Get("north")
	westQ := q.Get("west")
	southQ := q.Get("south")
	eastQ := q.Get("east")
	if northQ != "" && westQ != "" && eastQ != "" && southQ != "" {
		bb.North, err = strconv.ParseFloat(northQ, 64)
		if err != nil {
			return bb, err
		}
		bb.West, err = strconv.ParseFloat(westQ, 64)
		if err != nil {
			return bb, err
		}
		bb.South, err = strconv.ParseFloat(southQ, 64)
		if err != nil {
			return bb, err
		}
		bb.East, err = strconv.ParseFloat(eastQ, 64)
		if err != nil {
			return bb, err
		}

		// TODO: Validate parameters
	}
	return bb, nil
}

func parseWithinRangeRequest(r *http.Request) (LocationRange, error) {
	var err error
	var lr LocationRange
	q := r.URL.Query()

	latitudeQ := q.Get("latitude")
	longitudeQ := q.Get("longitude")
	distanceQ := q.Get("distance")
	if latitudeQ != "" && longitudeQ != "" && distanceQ != "" {
		lr.Latitude, err = strconv.ParseFloat(latitudeQ, 64)
		if err != nil {
			return lr, err
		}
		lr.Longitude, err = strconv.ParseFloat(longitudeQ, 64)
		if err != nil {
			return lr, err
		}
		lr.Distance, err = strconv.ParseFloat(distanceQ, 64)
		if err != nil {
			return lr, err
		}

		// TODO: Validate parameters
	}

	return lr, nil
}

func isWithinRangeRequest(r *http.Request) bool {
	q := r.URL.Query()
	return q.Has("latitude") && q.Has("longitude") && q.Has("distance")
}

func isWithinBoundingBoxRequest(r *http.Request) bool {
	q := r.URL.Query()
	return q.Has("north") && q.Has("west") && q.Has("east") && q.Has("south")
}

func (t *HTTPTransport) httpListDevices() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		filter, err := parseQueryFilter(r)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}
		var devices []Device

		// figure out what kind of query this is
		if isWithinRangeRequest(r) {
			lr, err := parseWithinRangeRequest(r)
			if err != nil {
				web.HTTPError(rw, err)
				return
			}
			devices, err = t.svc.ListInRange(r.Context(), lr, filter)
		} else if isWithinBoundingBoxRequest(r) {
			bb, err := parseBoundingBoxRequest(r)
			if err != nil {
				web.HTTPError(rw, err)
				return
			}
			devices, err = t.svc.ListInBoundingBox(r.Context(), bb, filter)
		} else {
			devices, err = t.svc.ListDevices(r.Context(), filter)
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

		var dto NewSensorOpts
		if err := web.DecodeJSON(r, &dto); err != nil {
			web.HTTPError(rw, err)
			return
		}

		fmt.Printf("Got: %+v\n", dto)
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
			id, err := strconv.Atoi(idString)
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
