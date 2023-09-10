package coretransport

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/core/devices"
	"sensorbucket.nl/sensorbucket/services/core/measurements"
)

func Create(r chi.Router, measurementService *measurements.Service, deviceService *devices.Service) http.Handler {
	r.Get("/datastreams/{id}", getDatastreams(measurementService, deviceService))
	return r
}

type GetDatastreamResponse struct {
	*measurements.Datastream
	Device *devices.Device
	Sensor *devices.Sensor
}

func getDatastreams(measurementService *measurements.Service, deviceService *devices.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idQ := chi.URLParam(r, "id")
		id, err := uuid.Parse(idQ)
		if err != nil {
			web.HTTPError(w, web.NewError(http.StatusBadRequest, "Invalid datastream ID", ""))
			return
		}

		ds, err := measurementService.GetDatastream(r.Context(), id)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		sensor, err := deviceService.GetSensor(r.Context(), ds.SensorID)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		device, err := deviceService.GetDevice(r.Context(), sensor.DeviceID)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
			Message: "",
			Data: GetDatastreamResponse{
				Datastream: ds,
				Device:     device,
				Sensor:     sensor,
			},
		})
	}
}
