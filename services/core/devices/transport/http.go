package devicetransport

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"sensorbucket.nl/sensorbucket/internal/httpfilter"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/core/devices"
)

var ErrHTTPDeviceIDInvalid = web.NewError(
	http.StatusBadRequest,
	"Device ID must be an integer",
	"DEVICE_ID_INVALID",
)

type middleware = func(next http.Handler) http.Handler

// HTTPTransport ...
type HTTPTransport struct {
	svc     *devices.Service
	router  chi.Router
	baseURL string
}

func NewHTTPTransport(svc *devices.Service, baseURL string) *HTTPTransport {
	transport := &HTTPTransport{
		svc:     svc,
		router:  chi.NewRouter(),
		baseURL: baseURL,
	}

	transport.router.Use(chimw.Logger)
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
	r.Get("/sensors/{id}", t.httpGetSensor())
	// TODO: Should we be able to fetch sensor by global unique ID?
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
}

type HTTPDeviceFilters struct {
	devices.DeviceFilter
	pagination.Request
	SensorGroup int64 `schema:"sensor_group"`
}

func (t *HTTPTransport) httpListDevices() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		filter, err := httpfilter.Parse[HTTPDeviceFilters](r)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}

		if filter.SensorGroup != 0 {
			sg, err := t.svc.GetSensorGroup(r.Context(), filter.SensorGroup)
			if err != nil {
				web.HTTPError(rw, err)
				return
			}
			filter.Sensor = append(filter.Sensor, sg.Sensors...)
			// If sensorgroup sensors is empty it would not filter on anything and return everything
			// we want atleast one filter such that nothing is returned
			// as nothing can have id 0 this filter will match nothing.
			filter.Sensor = append(filter.Sensor, 0)
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
		if err != nil {
			web.HTTPError(rw, err)
			return
		}
		page, err := t.svc.ListSensors(r.Context(), p)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}
		web.HTTPResponse(rw, http.StatusOK, pagination.CreateResponse(r, t.baseURL, *page))
	}
}

func (t *HTTPTransport) httpGetSensor() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sensorIDQ := chi.URLParam(r, "id")
		sensorID, err := strconv.ParseInt(sensorIDQ, 10, 64)
		if err != nil {
			web.HTTPError(w, web.NewError(http.StatusBadRequest, "invalid sensor id", ""))
			return
		}

		sensor, err := t.svc.GetSensor(r.Context(), sensorID)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
			Message: "Fetched sensor",
			Data:    sensor,
		})
	}
}

//
// Sensor Groups
//

func (t *HTTPTransport) httpCreateSensorGroup() http.HandlerFunc {
	type request struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if err := web.DecodeJSON(r, &req); err != nil {
			web.HTTPError(w, err)
			return
		}

		group, err := t.svc.CreateSensorGroup(r.Context(), req.Name, req.Description)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		web.HTTPResponse(w, http.StatusCreated, web.APIResponseAny{
			Message: fmt.Sprintf("Created sensor group '%s'", group.Name),
			Data:    group,
		})
	}
}

func (t *HTTPTransport) httpListSensorGroups() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, err := httpfilter.Parse[pagination.Request](r)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		page, err := t.svc.ListSensorGroups(r.Context(), p)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusOK, pagination.CreateResponse(r, t.baseURL, *page))
	}
}

func (t *HTTPTransport) httpGetSensorGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		qID := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(qID, 10, 64)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		sg, err := t.svc.GetSensorGroup(r.Context(), id)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
			Data: sg,
		})
	}
}

func (t *HTTPTransport) httpAddSensorToSensorGroup() http.HandlerFunc {
	type request struct {
		SensorID int64 `json:"sensor_id"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		sensorGroupID, err := urlParamInt64(r, "id")
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		var req request
		if err := web.DecodeJSON(r, &req); err != nil {
			web.HTTPError(w, err)
			return
		}
		err = t.svc.AddSensorToSensorGroup(r.Context(), sensorGroupID, req.SensorID)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusCreated, web.APIResponseAny{
			Message: "Added sensor to group",
		})
	}
}

func (t *HTTPTransport) httpDeleteSensorFromSensorGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sensorGroupID, err := urlParamInt64(r, "id")
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		sensorID, err := urlParamInt64(r, "sid")
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		err = t.svc.DeleteSensorFromSensorGroup(r.Context(), sensorGroupID, sensorID)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusCreated, web.APIResponseAny{
			Message: "Deleted sensor from group",
		})
	}
}

func (t *HTTPTransport) httpDeleteSensorGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sensorGroupID, err := urlParamInt64(r, "id")
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		group, err := t.svc.GetSensorGroup(r.Context(), sensorGroupID)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		err = t.svc.DeleteSensorGroup(r.Context(), group)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{Message: "Deleted sensor group"})
	}
}

func (t *HTTPTransport) httpUpdateSensorGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sensorGroupID, err := urlParamInt64(r, "id")
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		var dto devices.UpdateSensorGroupOpts
		if err := web.DecodeJSON(r, &dto); err != nil {
			web.HTTPError(w, err)
			return
		}

		group, err := t.svc.GetSensorGroup(r.Context(), sensorGroupID)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		err = t.svc.UpdateSensorGroup(r.Context(), group, dto)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
			Message: "Updated sensor group",
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

func urlParamInt64(r *http.Request, name string) (int64, error) {
	q := strings.Trim(chi.URLParam(r, name), " \r\n")
	if q == "" {
		return 0, web.NewError(http.StatusBadRequest, fmt.Sprintf("could not parse url parameter: missing %s url parameter", name), "")
	}
	i, err := strconv.ParseInt(q, 10, 64)
	if err != nil {
		return 0, web.NewError(http.StatusBadRequest, fmt.Sprintf("parameter %s is not an integer: %s", name, err), "")
	}
	return i, nil
}
