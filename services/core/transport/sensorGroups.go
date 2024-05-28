package coretransport

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"sensorbucket.nl/sensorbucket/internal/httpfilter"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/core/devices"
)

//
// Sensor Groups
//

func (t *CoreTransport) httpCreateSensorGroup() http.HandlerFunc {
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

		group, err := t.deviceService.CreateSensorGroup(r.Context(), req.Name, req.Description)
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

func (t *CoreTransport) httpListSensorGroups() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, err := httpfilter.Parse[pagination.Request](r)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		page, err := t.deviceService.ListSensorGroups(r.Context(), p)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusOK, pagination.CreateResponse(r, t.baseURL, *page))
	}
}

func (t *CoreTransport) httpGetSensorGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		qID := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(qID, 10, 64)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		sg, err := t.deviceService.GetSensorGroup(r.Context(), id)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
			Data: sg,
		})
	}
}

func (t *CoreTransport) httpAddSensorToSensorGroup() http.HandlerFunc {
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
		err = t.deviceService.AddSensorToSensorGroup(r.Context(), sensorGroupID, req.SensorID)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusCreated, web.APIResponseAny{
			Message: "Added sensor to group",
		})
	}
}

func (t *CoreTransport) httpDeleteSensorFromSensorGroup() http.HandlerFunc {
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
		err = t.deviceService.DeleteSensorFromSensorGroup(r.Context(), sensorGroupID, sensorID)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusCreated, web.APIResponseAny{
			Message: "Deleted sensor from group",
		})
	}
}

func (t *CoreTransport) httpDeleteSensorGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sensorGroupID, err := urlParamInt64(r, "id")
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		group, err := t.deviceService.GetSensorGroup(r.Context(), sensorGroupID)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		err = t.deviceService.DeleteSensorGroup(r.Context(), group)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{Message: "Deleted sensor group"})
	}
}

func (t *CoreTransport) httpUpdateSensorGroup() http.HandlerFunc {
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

		group, err := t.deviceService.GetSensorGroup(r.Context(), sensorGroupID)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		err = t.deviceService.UpdateSensorGroup(r.Context(), group, dto)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
			Message: "Updated sensor group",
		})
	}
}
