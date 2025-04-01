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
	SensorGroup int64 `url:"sensor_group"`
}

func (transport *CoreTransport) httpListDevices() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		filter, err := httpfilter.Parse[HTTPDeviceFilters](r)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}

		if filter.SensorGroup != 0 {
			sg, err := transport.deviceService.GetSensorGroup(r.Context(), filter.SensorGroup)
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

		page, err := transport.deviceService.ListDevices(r.Context(), filter.DeviceFilter, filter.Request)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, pagination.CreateResponse(r, transport.baseURL, *page))
	}
}

func (transport *CoreTransport) httpGetDevice() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		device := r.Context().Value(ctxDeviceKey).(*devices.Device)
		web.HTTPResponse(rw, http.StatusOK, &web.APIResponseAny{
			Message: "Fetched device",
			Data:    device,
		})
	}
}

func (transport *CoreTransport) httpCreateDevice() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var req devices.NewDeviceOpts
		if err := web.DecodeJSON(r, &req); err != nil {
			web.HTTPError(rw, err)
			return
		}

		dev, err := transport.deviceService.CreateDevice(r.Context(), req)
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

func (transport *CoreTransport) httpDeleteDevice() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		dev := r.Context().Value(ctxDeviceKey).(*devices.Device)

		if err := transport.deviceService.DeleteDevice(r.Context(), dev); err != nil {
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, &web.APIResponseAny{
			Message: "Deleted device",
		})
	}
}

func (transport *CoreTransport) httpUpdateDevice() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		dev := r.Context().Value(ctxDeviceKey).(*devices.Device)

		var dto devices.UpdateDeviceOpts
		if err := web.DecodeJSON(r, &dto); err != nil {
			web.HTTPError(rw, err)
			return
		}

		if err := transport.deviceService.UpdateDevice(r.Context(), dev, dto); err != nil {
			web.HTTPError(rw, err)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, &web.APIResponseAny{
			Message: "Updated device",
		})
	}
}
