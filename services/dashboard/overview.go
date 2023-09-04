package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"

	"sensorbucket.nl/sensorbucket/internal/pagination"
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
	r.Get("/devices/stream-map", devicesStreamMap())
	r.Get("/devices/table", func(w http.ResponseWriter, r *http.Request) {
		var err error
		sgID := r.URL.Query().Get("sensor_group")
		var sg *devices.SensorGroup
		if sgID != "" {
			sg, err = getSG(sgID)
			if err != nil {
				web.HTTPError(w, err)
				return
			}
		}
		url := "http://core:3000/devices?" + r.URL.Query().Encode()
		res, err := http.Get(url)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		var resBody web.APIResponse[[]devices.Device]
		if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
			web.HTTPError(w, err)
			return
		}
		w.Header().Set("hx-push-url", "/overview?"+r.URL.Query().Encode())
		w.Header().Set("hx-trigger-after-settle", "newDeviceList")
		views.WriteRenderFilters(w, sg, true)
		views.WriteRenderDeviceTable(w, resBody.Data)
	})
	r.With(resolveDevice).Get("/devices/{device_id}", deviceDetailPage())
	r.With(resolveDevice).With(resolveSensor).Get("/devices/{device_id}/sensors/{sensor_code}", sensorDetailPage())

	r.Get("/sensor-groups", searchSensorGroups())

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

func searchSensorGroups() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := http.Get("http://core:3000/sensor-groups")
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		var resBody web.APIResponse[[]devices.SensorGroup]
		if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
			web.HTTPError(w, err)
			return
		}
		views.WriteSensorGroupSearch(w, resBody.Data)
	}
}

func getSG(id string) (*devices.SensorGroup, error) {
	res, err := http.Get("http://core:3000/sensor-groups/" + id)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, errors.New("could not get SensorGroup")
	}
	var resBody web.APIResponse[devices.SensorGroup]
	if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
		return nil, err
	}

	return &resBody.Data, nil
}

func deviceListPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := &views.DeviceListPage{}

		sensorGroupID := r.URL.Query().Get("sensor_group")
		if sensorGroupID != "" {
			sg, err := getSG(sensorGroupID)
			if err != nil {
				web.HTTPError(w, err)
				return
			}
			page.SensorGroup = sg
		}

		q := url.Values{}
		if page.SensorGroup != nil {
			q.Set("sensor_group", strconv.FormatInt(page.SensorGroup.ID, 10))
		}
		url := "http://core:3000/devices?" + q.Encode()
		res, err := http.Get(url)
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		var resBody web.APIResponse[[]devices.Device]
		if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
			web.HTTPError(w, err)
			return
		}
		page.Devices = resBody.Data

		if isHX(r) {
			page.WriteBody(w)
			return
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

func devicesStreamMap() http.HandlerFunc {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	getDevicePage := func(sensorGroupID, cursor string) ([]devices.Device, string, error) {
		q := url.Values{}
		q.Set("cursor", cursor)
		if sensorGroupID != "" {
			q.Set("sensor_group", sensorGroupID)
		}
		res, err := http.Get("http://core:3000/devices?" + q.Encode())
		if err != nil {
			return nil, "", err
		}
		var resBody pagination.APIResponse[devices.Device]
		if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
			return nil, "", err
		}
		return resBody.Data, resBody.Links.Next, nil
	}
	type Marker struct {
		DeviceID  int64   `json:"device_id"`
		Label     string  `json:"label"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		sensorGroupID := r.URL.Query().Get("sensor_group")
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		go func() {
			defer ws.Close()
			var nextCursor string
			for {
				// Start fetching pages of devices and stream them to the client
				devices, cursor, err := getDevicePage(sensorGroupID, nextCursor)
				if err != nil {
					log.Printf("Failed to fetch devices for client: %v\n", err)
					return
				}

				for _, dev := range devices {
					if dev.Latitude == nil || dev.Longitude == nil {
						continue
					}
					writer, err := ws.NextWriter(websocket.TextMessage)
					if err != nil {
						log.Printf("cannot open writer for ws: %v\n", err)
						return
					}
					defer writer.Close()
					frame := fmt.Sprintf(`{"device_id": %d, "device_code": "%s", "coordinates": [%f,%f]}`, dev.ID, dev.Code, *dev.Latitude, *dev.Longitude)
					writer.Write([]byte(frame))
				}
				nextCursor = cursor
				if nextCursor == "" {
					return
				}
			}
		}()
	}
}

func overviewDatastream() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var start time.Time
		var end time.Time

		startQ := r.URL.Query().Get("start")
		endQ := r.URL.Query().Get("end")
		if startQ != "" {
			start, err = time.Parse("2006-01-02", startQ)
			if err != nil {
				web.HTTPError(w, err)
			}
		}
		if endQ != "" {
			end, err = time.Parse("2006-01-02", endQ)
			if err != nil {
				web.HTTPError(w, err)
			}
		}

		if start.IsZero() {
			start = time.Now().Add(-7 * 24 * time.Hour)
		}
		if end.IsZero() {
			end = time.Now()
		}
		res, err := http.Get(fmt.Sprintf("http://core:3000/datastreams/%s", chi.URLParam(r, "id")))
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		if res.StatusCode != 200 {
			var apiError web.APIError
			if err := json.NewDecoder(res.Body).Decode(&apiError); err != nil {
				web.HTTPError(w, err)
				return
			}
			apiError.HTTPStatus = res.StatusCode
			web.HTTPError(w, &apiError)
			return
		}
		var resBody web.APIResponse[coretransport.GetDatastreamResponse]
		if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
			web.HTTPError(w, err)
			return
		}

		page := &views.DatastreamPage{
			Datastream: *resBody.Data.Datastream,
			Device:     *resBody.Data.Device,
			Sensor:     *resBody.Data.Sensor,
			Start:      start,
			End:        end,
		}
		if isHX(r) {
			page.WriteBody(w)
			return
		}

		views.WriteIndex(w, page)
	}
}

func overviewDatastreamStream() http.HandlerFunc {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	getMeasurementsPage := func(dsID string, start, end time.Time, cursor string) ([]measurements.Measurement, string, error) {
		q := url.Values{}
		q.Set("datastream", dsID)
		q.Set("start", start.Format(time.RFC3339))
		q.Set("end", end.Format(time.RFC3339))
		q.Set("cursor", cursor)
		res, err := http.Get("http://core:3000/measurements?" + q.Encode())
		if err != nil {
			return nil, "", err
		}
		if res.StatusCode != 200 {
			var apiError web.APIError
			if err := json.NewDecoder(res.Body).Decode(&apiError); err != nil {
				return nil, "", err
			}
			apiError.HTTPStatus = res.StatusCode
			return nil, "", &apiError
		}
		var resBody pagination.APIResponse[measurements.Measurement]
		if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
			return nil, "", err
		}

		nextCursor := ""
		if resBody.Links.Next != "" {
			next, err := url.Parse(resBody.Links.Next)
			if err != nil {
				return nil, "", errors.New("stream datastream, invalid next link in paginated response")
			}
			nextCursor = next.Query().Get("cursor")
		}

		return resBody.Data, nextCursor, nil
	}
	return func(w http.ResponseWriter, r *http.Request) {
		datastreamID := chi.URLParam(r, "id")
		start, err := time.Parse(time.RFC3339, r.URL.Query().Get("start"))
		if err != nil {
			web.HTTPError(w, web.NewError(http.StatusBadRequest, "Start parameter is not ISO8601/RFC3339", ""))
		}
		end, err := time.Parse(time.RFC3339, r.URL.Query().Get("end"))
		if err != nil {
			web.HTTPError(w, web.NewError(http.StatusBadRequest, "End parameter is not ISO8601/RFC3339", ""))
		}

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		go func() {
			var nextCursor string
			defer ws.Close()
			for {
				// Start fetching pages of measurements and stream them to the client
				measurements, cursor, err := getMeasurementsPage(datastreamID, start, end, nextCursor)
				if err != nil {
					log.Printf("Failed to fetch devices for client: %v\n", err)
					return
				}
				log.Printf("%d measurements, %s cursor, %v err", len(measurements), cursor, err)

				for _, point := range measurements {
					writer, err := ws.NextWriter(websocket.BinaryMessage)
					if err != nil {
						log.Printf("cannot open writer for ws: %v\n", err)
						return
					}
					defer writer.Close()
					// Write to client
					binary.Write(writer, binary.BigEndian, point.MeasurementTimestamp.UnixMilli())
					binary.Write(writer, binary.BigEndian, point.MeasurementValue)
				}
				nextCursor = cursor
				if nextCursor == "" {
					return
				}
			}
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
			web.HTTPError(w, errors.New("resolveSensor middleware is missing device in context, did you use resolveDevice?\n"))
			return
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
