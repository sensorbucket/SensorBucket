package coretransport

import (
	"net/http"

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

type HTTPDeviceFilters struct {
	devices.DeviceFilter
	pagination.Request
	SensorGroup int64 `schema:"sensor_group"`
}

func (t *CoreTransport) httpListDevices() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		filter, err := httpfilter.Parse[HTTPDeviceFilters](r)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}

		if filter.SensorGroup != 0 {
			sg, err := t.deviceService.GetSensorGroup(r.Context(), filter.SensorGroup)
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

		page, err := t.deviceService.ListDevices(r.Context(), filter.DeviceFilter, filter.Request)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, pagination.CreateResponse(r, t.baseURL, *page))
	}
}

func (t *CoreTransport) httpGetDevice() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		device := r.Context().Value(ctxDeviceKey).(*devices.Device)
		web.HTTPResponse(rw, http.StatusOK, &web.APIResponseAny{
			Message: "Fetched device",
			Data:    device,
		})
	}
}

func (t *CoreTransport) httpCreateDevice() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var req devices.NewDeviceOpts
		if err := web.DecodeJSON(r, &req); err != nil {
			web.HTTPError(rw, err)
			return
		}

		dev, err := t.deviceService.CreateDevice(r.Context(), req)
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

func (t *CoreTransport) httpDeleteDevice() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		dev := r.Context().Value(ctxDeviceKey).(*devices.Device)

		if err := t.deviceService.DeleteDevice(r.Context(), dev); err != nil {
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, &web.APIResponseAny{
			Message: "Deleted device",
		})
	}
}

func (t *CoreTransport) httpUpdateDevice() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		dev := r.Context().Value(ctxDeviceKey).(*devices.Device)

		var dto devices.UpdateDeviceOpts
		if err := web.DecodeJSON(r, &dto); err != nil {
			web.HTTPError(rw, err)
			return
		}

		if err := t.deviceService.UpdateDevice(r.Context(), dev, dto); err != nil {
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, &web.APIResponseAny{
			Message: "Updated device",
		})
	}
}
