package main

import (
	"context"
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

type middlewareFunc = func(next http.Handler) http.Handler

func createOverviewPageHandler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.GetHead)
	r.Get("/", overview())
	r.With(resolveDevice).Get("/devices/{device_id}", overviewSelectDevice())
	r.With(resolveDevice, resolveSensor).
		Get("/devices/{device_id}/sensors/{sensor_code}", overviewSelectSensor())
	// r.Get("/devices/{device_id}/sensors/{sensor_id}/datastreams/{datastream_id}", overviewSensors())

	// r.Get("/datastreams/{id}", overviewDatastream())
	// r.Get("/datastreams/{id}/stream", overviewDatastreamStream())
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

		device, ok := r.Context().Value("device").(*devices.Device)
		if ok {
			page.SelectedDevice = device
			page.Sensors = device.Sensors
		}
		sensor, ok := r.Context().Value("sensor").(*devices.Sensor)
		if ok {
			page.SelectedSensor = sensor
			datastreams, err := getDatastreamsBySensor(sensor.ID)
			if err != nil {
				web.HTTPError(w, err)
				return
			}
			page.Datastreams = datastreams
		}
		views.WriteIndex(w, page)
	}
}

func overviewSelectDevice() http.HandlerFunc {
	fallback := overview()
	return func(w http.ResponseWriter, r *http.Request) {
		if !isHX(r) {
			fallback(w, r)
			return
		}
		device := r.Context().Value("device").(*devices.Device)
		views.WriteRenderSensorTable(w, device.Sensors, 0)
	}
}

func overviewSelectSensor() http.HandlerFunc {
	fallback := overview()
	return func(w http.ResponseWriter, r *http.Request) {
		if !isHX(r) {
			fallback(w, r)
			return
		}
		sensor := r.Context().Value("sensor").(*devices.Sensor)
		datastreams, err := getDatastreamsBySensor(sensor.ID)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		views.WriteRenderDatastreamTable(w, datastreams)
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

// =============
// Helpers and middleware
// =============

func resolveDevice(next http.Handler) http.Handler {
	getDevice := func(id int64) (*devices.Device, error) {
		res, _ := http.Get(fmt.Sprintf("http://core:3000/devices/%d", id))
		if res.StatusCode != 200 {
			body, _ := io.ReadAll(res.Body)
			return nil, fmt.Errorf("error getting device: %s", string(body))
		}
		var resBody web.APIResponse[devices.Device]
		if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
			return nil, err
		}
		return &resBody.Data, nil
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deviceID, err := URLParamInt(r, "device_id")
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		device, err := getDevice(deviceID)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		r = r.WithContext(
			context.WithValue(
				r.Context(),
				"device",
				device,
			),
		)
		next.ServeHTTP(w, r)
	})
}

func resolveSensor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sensorCode := chi.URLParam(r, "sensor_code")

		device, ok := r.Context().Value("device").(*devices.Device)
		if !ok {
			panic("NODEVICE")
		}

		sensor, err := device.GetSensorByCode(sensorCode)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		r = r.WithContext(
			context.WithValue(
				r.Context(),
				"sensor",
				sensor,
			),
		)
		next.ServeHTTP(w, r)
	})
}

func getDatastreamsBySensor(id int64) ([]measurements.Datastream, error) {
	res, err := http.Get(fmt.Sprintf("http://core:3000/datastreams?sensor=%d", id))
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("could not fetch datastreams from remote: %s", string(body))
	}
	var resBody web.APIResponse[[]measurements.Datastream]
	if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
		return nil, err
	}

	return resBody.Data, nil
}
