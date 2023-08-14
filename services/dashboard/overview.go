package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"

	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/core/devices"
	"sensorbucket.nl/sensorbucket/services/core/measurements"
	coretransport "sensorbucket.nl/sensorbucket/services/core/transport"
	"sensorbucket.nl/sensorbucket/services/dashboard/views"
)

func createOverviewPageHandler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.GetHead)
	r.Get("/", overview())
	r.Get("/sensors/{id}", overviewSensors())
	r.Get("/datastreams/{id}", overviewDatastream())
	r.Get("/datastreams/{id}/stream", overviewDatastreamStream())
	return r
}

func isHX(r *http.Request) bool {
	return r.Header.Get("hx-request") == "true"
}

func URLParamInt(r *http.Request, name string) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, name), 10, 64)
}

func overview() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, _ := http.Get("http://core:3000/devices")
		var resBody web.APIResponse[[]devices.Device]
		if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
			return
		}

		page := &views.OverviewPage{}
		page.Devices = resBody.Data
		views.WriteIndex(w, page)
	}
}

func overviewSensors() http.HandlerFunc {
	getSensor := func(id int64) (*devices.Sensor, error) {
		res, _ := http.Get(fmt.Sprintf("http://core:3000/sensors/%d", id))
		if res.StatusCode != 200 {
			body, _ := io.ReadAll(res.Body)
			return nil, fmt.Errorf("error getting sensor: %s", string(body))
		}
		var resBody web.APIResponse[devices.Sensor]
		if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
			return nil, err
		}
		return &resBody.Data, nil
	}
	getDatastreamsForSensor := func(id int64) ([]measurements.Datastream, error) {
		res, _ := http.Get(fmt.Sprintf("http://core:3000/datastreams?sensor=%d", id))
		if res.StatusCode != 200 {
			body, _ := io.ReadAll(res.Body)
			return nil, fmt.Errorf("error getting datastreams for sensor: %s", string(body))
		}
		var resBody web.APIResponse[[]measurements.Datastream]
		if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
			return nil, err
		}
		return resBody.Data, nil
	}
	return func(w http.ResponseWriter, r *http.Request) {
		sensorIDQ := chi.URLParam(r, "id")
		sensorID, err := strconv.ParseInt(sensorIDQ, 10, 64)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		sensor, err := getSensor(sensorID)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		datastreams, err := getDatastreamsForSensor(sensorID)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		views.WriteSensorRow(w, *sensor, datastreams, true)
	}
}

func overviewDatastream() http.HandlerFunc {
	getDatastream := func(r *http.Request) (*coretransport.GetDatastreamResponse, error) {
		datastreamUUID := chi.URLParam(r, "id")
		res, err := http.Get("http://core:3000/datastreams/" + datastreamUUID)
		if err != nil {
			return nil, err
		}
		if res.StatusCode != 200 {
			body, _ := io.ReadAll(res.Body)
			fmt.Printf("Error fetching datastream (%s): %s\n", datastreamUUID, string(body))
			return nil, errors.New("could not fetch datastream")
		}
		var resBody web.APIResponse[coretransport.GetDatastreamResponse]
		if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
			return nil, err
		}
		return &resBody.Data, nil
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ds, err := getDatastream(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		page := &views.DatastreamPage{
			Device:     *ds.Device,
			Sensor:     *ds.Sensor,
			Datastream: *ds.Datastream,
		}
		views.WriteIndex(w, page)
	}
}

func overviewDatastreamStream() http.HandlerFunc {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		go func() {
			ws.Close()
		}()
	}
}
