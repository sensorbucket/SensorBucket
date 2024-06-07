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

func (t *CoreTransport) httpListDeviceSensors() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		device := r.Context().Value(ctxDeviceKey).(*devices.Device)

		web.HTTPResponse(rw, http.StatusOK, &web.APIResponseAny{
			Message: "Listed sensors",
			Data:    device.Sensors,
		})
	}
}

func (t *CoreTransport) httpAddSensor() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		dev := r.Context().Value(ctxDeviceKey).(*devices.Device)

		var dto devices.NewSensorDTO
		if err := web.DecodeJSON(r, &dto); err != nil {
			web.HTTPError(rw, err)
			return
		}

		if err := t.deviceService.AddSensor(r.Context(), dev, dto); err != nil {
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusCreated, &web.APIResponseAny{
			Message: "Created new sensor for device",
		})
	}
}

func (t *CoreTransport) httpDeleteSensor() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		dev := r.Context().Value(ctxDeviceKey).(*devices.Device)

		sensor, err := dev.GetSensorByCode(chi.URLParam(r, "sensor_code"))
		if err != nil {
			web.HTTPError(rw, err)
			return
		}

		if err := t.deviceService.DeleteSensor(r.Context(), dev, sensor); err != nil {
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, &web.APIResponseAny{
			Message: "Deleted sensor from device",
		})
	}
}

func (t *CoreTransport) httpListSensors() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		p, err := httpfilter.Parse[pagination.Request](r)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}
		page, err := t.deviceService.ListSensors(r.Context(), p)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}
		web.HTTPResponse(rw, http.StatusOK, pagination.CreateResponse(r, t.baseURL, *page))
	}
}

func (t *CoreTransport) httpGetSensor() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sensorIDQ := chi.URLParam(r, "id")
		sensorID, err := strconv.ParseInt(sensorIDQ, 10, 64)
		if err != nil {
			web.HTTPError(w, web.NewError(http.StatusBadRequest, "invalid sensor id", ""))
			return
		}

		sensor, err := t.deviceService.GetSensor(r.Context(), sensorID)
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
