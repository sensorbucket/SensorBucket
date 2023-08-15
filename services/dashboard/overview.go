package main

import (
	"context"
	"encoding/json"
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
	r.Get("/", deviceListPage())
	r.With(resolveDevice).Get("/devices/{device_id}", deviceDetailPage())
	r.With(resolveDevice).With(resolveSensor).Get("/devices/{device_id}/sensors/{sensor_code}", sensorDetailPage())

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

func deviceListPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := http.Get("http://core:3000/devices")
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		var resBody web.APIResponse[[]devices.Device]
		if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
			web.HTTPError(w, err)
			return
		}
		page := &views.DeviceListPage{
			Devices: resBody.Data,
		}
		views.WriteIndex(w, page)
	}
}

func deviceDetailPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		device := r.Context().Value("device").(*devices.Device)
		page := &views.DeviceDetailPage{
			Device: *device,
		}
		if isHX(r) {
			page.WriteBody(w)
			return
		}
		views.WriteIndex(w, page)
	}
}

func sensorDetailPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		device := r.Context().Value("device").(*devices.Device)
		sensor := r.Context().Value("sensor").(*devices.Sensor)

		res, err := http.Get(fmt.Sprintf("http://core:3000/datastreams?sensor=%d", sensor.ID))
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		var resBody web.APIResponse[[]measurements.Datastream]
		if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
			web.HTTPError(w, err)
			return
		}

		page := &views.SensorDetailPage{
			Device:      *device,
			Sensor:      *sensor,
			Datastreams: resBody.Data,
		}
		if isHX(r) {
			page.WriteBody(w)
			return
		}
		views.WriteIndex(w, page)
	}
}

func overviewDatastream() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := http.Get(fmt.Sprintf("http://core:3000/datastreams/%s", chi.URLParam(r, "id")))
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		var resBody web.APIResponse[coretransport.GetDatastreamResponse]
		if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
			web.HTTPError(w, err)
			return
		}

		views.WriteIndex(w, &views.DatastreamPage{
			Datastream: *resBody.Data.Datastream,
			Device:     *resBody.Data.Device,
			Sensor:     *resBody.Data.Sensor,
		})
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
