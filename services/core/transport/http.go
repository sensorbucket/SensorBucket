package coretransport

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/core/devices"
	"sensorbucket.nl/sensorbucket/services/core/measurements"
)

func Create(r chi.Router, measurementService *measurements.Service, deviceService *devices.Service) http.Handler {
	r.Get("/datastreams/{id}", getDatastreams(measurementService, deviceService))
	return r
}

type GetDatastreamResponse struct {
	Datastream                 *measurements.Datastream `json:"datastream"`
	Device                     *devices.Device          `json:"device"`
	Sensor                     *devices.Sensor          `json:"sensor"`
	LatestMeasurementValue     float64                  `json:"latest_measurement_value"`
	LatestMeasurementTimestamp time.Time                `json:"latest_measurement_timestamp"`
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
		res := GetDatastreamResponse{
			Datastream: ds,
			Device:     device,
			Sensor:     sensor,
		}

		m, err := measurementService.QueryMeasurements(measurements.Filter{
			Datastream: []string{ds.ID.String()},
		}, pagination.Request{Limit: 1})
		if len(m.Data) > 0 {
			res.LatestMeasurementValue = m.Data[0].MeasurementValue
			res.LatestMeasurementTimestamp = m.Data[0].MeasurementTimestamp
		}

		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
			Message: "Fetched detailed datastream",
			Data:    res,
		})
	}
}
