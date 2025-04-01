package coretransport

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"sensorbucket.nl/sensorbucket/internal/httpfilter"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/core/devices"
	"sensorbucket.nl/sensorbucket/services/core/measurements"
)

type GetDatastreamResponse struct {
	Datastream                 *measurements.Datastream `json:"datastream"`
	Device                     *devices.Device          `json:"device"`
	Sensor                     *devices.Sensor          `json:"sensor"`
	LatestMeasurementValue     float64                  `json:"latest_measurement_value"`
	LatestMeasurementTimestamp time.Time                `json:"latest_measurement_timestamp"`
}

func (transport *CoreTransport) httpGetDatastream() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idQ := chi.URLParam(r, "id")
		id, err := uuid.Parse(idQ)
		if err != nil {
			web.HTTPError(w, web.NewError(http.StatusBadRequest, "Invalid datastream ID", ""))
			return
		}

		ds, err := transport.measurementService.GetDatastream(r.Context(), id)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		sensor, err := transport.deviceService.GetSensor(r.Context(), ds.SensorID)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		device, err := transport.deviceService.GetDevice(r.Context(), sensor.DeviceID)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		res := GetDatastreamResponse{
			Datastream: ds,
			Device:     device,
			Sensor:     sensor,
		}

		m, err := transport.measurementService.QueryMeasurements(r.Context(), measurements.Filter{
			Datastream: []string{ds.ID.String()},
		}, pagination.Request{Limit: 1})
		if err != nil {
			web.HTTPError(w, err)
			return
		}
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

func (transport *CoreTransport) httpListDatastream() http.HandlerFunc {
	type params struct {
		measurements.DatastreamFilter
		pagination.Request
	}
	return func(rw http.ResponseWriter, r *http.Request) {
		params, err := httpfilter.Parse[params](r)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}

		page, err := transport.measurementService.ListDatastreams(r.Context(), params.DatastreamFilter, params.Request)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}
		web.HTTPResponse(rw, http.StatusOK, pagination.CreateResponse(r, transport.baseURL, *page))
	}
}
