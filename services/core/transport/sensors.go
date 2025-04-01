package coretransport

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"sensorbucket.nl/sensorbucket/internal/httpfilter"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/core/devices"
)

var ErrHTTPSensorIDInvalid = web.NewError(
	http.StatusBadRequest,
	"Sensor ID must be an integer",
	"SENSOR_ID_INVALID",
)

func (transport *CoreTransport) httpListDeviceSensors() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		device := r.Context().Value(ctxDeviceKey).(*devices.Device)

		web.HTTPResponse(rw, http.StatusOK, &web.APIResponseAny{
			Message: "Listed sensors",
			Data:    device.Sensors,
		})
	}
}

func (transport *CoreTransport) httpAddSensor() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		dev := r.Context().Value(ctxDeviceKey).(*devices.Device)

		var dto devices.NewSensorDTO
		if err := web.DecodeJSON(r, &dto); err != nil {
			web.HTTPError(rw, err)
			return
		}

		if err := transport.deviceService.AddSensor(r.Context(), dev, dto); err != nil {
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusCreated, &web.APIResponseAny{
			Message: "Created new sensor for device",
		})
	}
}

func (transport *CoreTransport) httpDeleteSensor() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		device := r.Context().Value(ctxDeviceKey).(*devices.Device)
		sensor := r.Context().Value(ctxSensorKey).(*devices.Sensor)

		if err := transport.deviceService.DeleteSensor(r.Context(), device, sensor); err != nil {
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, &web.APIResponseAny{
			Message: "Deleted sensor",
		})
	}
}

func (transport *CoreTransport) httpListSensors() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		p, err := httpfilter.Parse[pagination.Request](r)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}
		page, err := transport.deviceService.ListSensors(r.Context(), p)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}
		web.HTTPResponse(rw, http.StatusOK, pagination.CreateResponse(r, transport.baseURL, *page))
	}
}

func (transport *CoreTransport) httpGetSensor() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sensorCTX := r.Context().Value(ctxSensorKey)
		if sensorCTX != nil {
			web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
				Message: "Fetched sensor",
				Data:    sensorCTX.(*devices.Sensor),
			})
			return
		}
		sensorIDString := chi.URLParam(r, "sensor_id")
		if sensorIDString == "" {
			web.HTTPError(w, ErrHTTPSensorIDInvalid)
			return
		}
		sensorID, err := strconv.ParseInt(sensorIDString, 10, 64)
		if err != nil {
			web.HTTPError(w, ErrHTTPDeviceIDInvalid)
			return
		}
		sensor, err := transport.deviceService.GetSensor(r.Context(), sensorID)
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

func (transport *CoreTransport) httpUpdateSensor() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dev := r.Context().Value(ctxDeviceKey).(*devices.Device)
		sensor := r.Context().Value(ctxSensorKey).(*devices.Sensor)

		var dto devices.UpdateSensorOpts
		if err := web.DecodeJSON(r, &dto); err != nil {
			web.HTTPError(w, err)
			return
		}

		if err := transport.deviceService.UpdateSensor(r.Context(), dev, sensor, dto); err != nil {
			web.HTTPError(w, err)
			return
		}

		web.HTTPResponse(w, http.StatusOK, &web.APIResponseAny{
			Message: "Updated sensor",
		})
	}
}
