package devicetransport

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"sensorbucket.nl/sensorbucket/internal/httpfilter"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/core/devices"
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
	svc     devices.Service
	router  chi.Router
	baseURL string
}

func NewHTTPTransport(svc devices.Service, baseURL string) *HTTPTransport {
	transport := &HTTPTransport{
		svc:     svc,
		router:  chi.NewRouter(),
		baseURL: baseURL,
	}

	// Register endpoints
	transport.SetupRoutes(transport.router)

	return transport
}

func (t *HTTPTransport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.router.ServeHTTP(w, r)
}

// SetupRoutes creates router for the user SetupRoutes
func (t *HTTPTransport) SetupRoutes(r chi.Router) {
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
			r.Delete("/{sensor_code}", t.httpDeleteSensor())
		})
	})
	r.Get("/sensors", t.httpListSensors())
	// TODO: Should we be able to fetch sensor by global unique ID?
}

type HTTPDeviceFilters struct {
	devices.DeviceFilter
	pagination.Request
}

func (t *HTTPTransport) httpListDevices() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		filter, err := httpfilter.Parse[HTTPDeviceFilters](r)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}

		page, err := t.svc.ListDevices(r.Context(), filter.DeviceFilter, filter.Request)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, pagination.CreateResponse(r, t.baseURL, *page))
	}
}

func (t *HTTPTransport) httpGetDevice() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		device := r.Context().Value(ctxDeviceKey).(*devices.Device)
		web.HTTPResponse(rw, http.StatusOK, &web.APIResponseAny{
			Message: "Fetched device",
			Data:    device,
		})
	}
}

func (t *HTTPTransport) httpCreateDevice() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var req devices.NewDeviceOpts
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
		dev := r.Context().Value(ctxDeviceKey).(*devices.Device)

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
		dev := r.Context().Value(ctxDeviceKey).(*devices.Device)

		var dto devices.UpdateDeviceOpts
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

func (t *HTTPTransport) httpListDeviceSensors() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		device := r.Context().Value(ctxDeviceKey).(*devices.Device)

		web.HTTPResponse(rw, http.StatusOK, &web.APIResponseAny{
			Message: "Listed sensors",
			Data:    device.Sensors,
		})
	}
}

func (t *HTTPTransport) httpAddSensor() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		dev := r.Context().Value(ctxDeviceKey).(*devices.Device)

		var dto devices.NewSensorDTO
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
		dev := r.Context().Value(ctxDeviceKey).(*devices.Device)

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

func (t *HTTPTransport) httpListSensors() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		p, err := httpfilter.Parse[pagination.Request](r)
		page, err := t.svc.ListSensors(r.Context(), p)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}
		web.HTTPResponse(rw, http.StatusOK, pagination.CreateResponse(r, t.baseURL, *page))
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
