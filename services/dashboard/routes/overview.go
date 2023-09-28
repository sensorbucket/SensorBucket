package routes

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/api"
	"sensorbucket.nl/sensorbucket/services/dashboard/views"
)

type middlewareFunc = func(next http.Handler) http.Handler

type OverviewRoute struct {
	router chi.Router
	client *api.APIClient
}

func CreateOverviewPageHandler(client *api.APIClient) *OverviewRoute {
	t := &OverviewRoute{
		client: client,
		router: chi.NewRouter(),
	}
	t.SetupRoutes(t.router)
	return t
}

func (t OverviewRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.router.ServeHTTP(w, r)
}

func (t *OverviewRoute) SetupRoutes(r chi.Router) {
	r.Use(middleware.GetHead)
	r.Get("/", t.deviceListPage())
	r.Get("/devices/stream-map", t.devicesStreamMap())
	r.Get("/devices/table", t.getDevicesTable())
	r.With(t.resolveDevice).Get("/devices/{device_id}", t.deviceDetailPage())
	r.With(t.resolveDevice, t.resolveSensor).Get("/devices/{device_id}/sensors/{sensor_code}", t.sensorDetailPage())

	r.Get("/sensor-groups", t.searchSensorGroups())
	r.Post("/sensor-groups", t.createSensorGroup())
	r.Delete("/sensor-groups", t.deleteSensorGroup())

	r.Get("/datastreams/{id}", t.overviewDatastream())
	r.Get("/datastreams/{id}/stream", t.overviewDatastreamStream())
}

func isHX(r *http.Request) bool {
	return r.Header.Get("hx-request") == "true"
}

func URLParamInt(r *http.Request, name string) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, name), 10, 64)
}

// TODO: Rename to FilterOnSensorGroup
func (t *OverviewRoute) createSensorGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		sgIDStr := r.URL.Query().Get("sensor_group")
		var sg *api.SensorGroup
		if sgIDStr != "" {
			sgID, err := strconv.ParseInt(sgIDStr, 10, 64)
			if err != nil {
				web.HTTPError(w, err)
				return
			}
			res, _, err := t.client.DevicesApi.GetSensorGroup(r.Context(), sgID).Execute()
			if err != nil {
				web.HTTPError(w, err)
				return
			}
			sg = res.Data
		}
		req := t.client.DevicesApi.ListDevices(r.Context())
		if sg != nil {
			req = req.SensorGroup([]int64{sg.GetId()})
		}
		res, _, err := req.Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		w.Header().Set("hx-push-url", "/overview?"+r.URL.Query().Encode())
		w.Header().Set("hx-trigger-after-settle", "newDeviceList")
		views.WriteRenderFilters(w, sg, true)
		views.WriteRenderDeviceTable(w, res.Data, getCursor(res.Links.GetNext()))
	}
}

func (t *OverviewRoute) deleteSensorGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, _, err := t.client.DevicesApi.ListDevices(r.Context()).Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		w.Header().Set("hx-push-url", "/overview?"+r.URL.Query().Encode())
		w.Header().Set("hx-trigger-after-settle", "newDeviceList")
		views.WriteRenderFilters(w, nil, true)
		views.WriteRenderDeviceTable(w, res.Data, getCursor(res.Links.GetNext()))
	}
}

func (t *OverviewRoute) getDevicesTable() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := t.client.DevicesApi.ListDevices(r.Context())
		if r.URL.Query().Has("sensor_group") {
			sgID, err := strconv.ParseInt(r.URL.Query().Get("sensor_group"), 10, 64)
			if err != nil {
				web.HTTPError(w, err)
				return
			}
			req = req.SensorGroup([]int64{sgID})
		}
		res, _, err := req.Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		nextCursor := ""
		if res.Links.GetNext() != "" {
			nextCursor = "/overview/devices/table?cursor=" + getCursor(res.Links.GetNext())
		}
		views.WriteRenderDeviceTable(w, res.Data, nextCursor)
	}
}

func (t *OverviewRoute) searchSensorGroups() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, _, err := t.client.DevicesApi.ListSensorGroups(r.Context()).Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		views.WriteSensorGroupSearch(w, res.Data)
	}
}

func (t *OverviewRoute) deviceListPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		page := &views.DeviceListPage{}
		sensorGroupIDStr := r.URL.Query().Get("sensor_group")
		if sensorGroupIDStr != "" {
			sensorGroupID, err := strconv.ParseInt(sensorGroupIDStr, 10, 64)
			res, _, err := t.client.DevicesApi.GetSensorGroup(r.Context(), sensorGroupID).Execute()
			if err != nil {
				web.HTTPError(w, err)
				return
			}
			page.SensorGroup = res.Data
		}
		req := t.client.DevicesApi.ListDevices(r.Context())
		if page.SensorGroup != nil {
			req = req.SensorGroup([]int64{page.SensorGroup.GetId()})
		}
		res, _, err := req.Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		page.Devices = res.Data

		if res.Links.GetNext() != "" {
			u, err := url.Parse(res.Links.GetNext())
			if err == nil {
				page.DevicesNextPage = "/overview/devices/table?cursor=" + u.Query().Get("cursor")
			}
		}

		if isHX(r) {
			page.WriteBody(w)
			return
		}
		views.WriteIndex(w, page)
	}
}

func (t *OverviewRoute) deviceDetailPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		device, _ := getDevice(r.Context())
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

func (t *OverviewRoute) sensorDetailPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		device, _ := getDevice(r.Context())
		sensor, _ := getSensor(r.Context())

		res, _, err := t.client.MeasurementsApi.ListDatastreams(r.Context()).Sensor([]int64{sensor.Id}).Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		page := &views.SensorDetailPage{
			Device:      *device,
			Sensor:      *sensor,
			Datastreams: res.Data,
		}
		if isHX(r) {
			page.WriteBody(w)
			return
		}
		views.WriteIndex(w, page)
	}
}

func (t *OverviewRoute) devicesStreamMap() http.HandlerFunc {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	type Marker struct {
		DeviceID  int64   `json:"device_id"`
		Label     string  `json:"label"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var sgID int64 = 0
		if r.URL.Query().Has("sensor_group") {
			id, err := strconv.ParseInt(r.URL.Query().Get("sensor_group"), 10, 64)
			if err != nil {
				web.HTTPError(w, err)
				return
			}
			sgID = id
		}

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		go func() {
			ctx, cancel := context.WithCancel(context.TODO())
			defer cancel()
			defer ws.Close()
			var nextCursor string
			for {
				// Start fetching pages of devices and stream them to the client
				res, _, err := t.client.DevicesApi.ListDevices(ctx).Cursor(nextCursor).SensorGroup([]int64{sgID}).Execute()
				if err != nil {
					log.Printf("Failed to fetch devices for client: %v\n", err)
					return
				}

				for _, dev := range res.Data {
					if dev.Latitude == nil || dev.Longitude == nil {
						continue
					}
					writer, err := ws.NextWriter(websocket.TextMessage)
					if err != nil {
						log.Printf("cannot open writer for ws: %v\n", err)
						return
					}
					defer writer.Close()
					frame := fmt.Sprintf(`{"device_id": %d, "device_code": "%s", "coordinates": [%f,%f]}`, dev.Id, dev.Code, dev.GetLatitude(), dev.GetLongitude())
					writer.Write([]byte(frame))
				}
				nextCursor = getCursor(res.Links.GetNext())
				if nextCursor == "" {
					return
				}
			}
		}()
	}
}

func (t *OverviewRoute) overviewDatastream() http.HandlerFunc {
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
				return
			}
		}
		if endQ != "" {
			end, err = time.Parse("2006-01-02", endQ)
			if err != nil {
				web.HTTPError(w, err)
				return
			}
		}

		if start.IsZero() {
			start = time.Now().Add(-7 * 24 * time.Hour)
		}
		if end.IsZero() {
			end = time.Now()
		}

		res, _, err := t.client.MeasurementsApi.GetDatastream(r.Context(), chi.URLParam(r, "id")).Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		page := &views.DatastreamPage{
			Datastream: *res.Data.Datastream,
			Device:     *res.Data.Device,
			Sensor:     *res.Data.Sensor,
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

func (t *OverviewRoute) overviewDatastreamStream() http.HandlerFunc {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
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
			ctx, cancel := context.WithCancel(context.TODO())
			defer cancel()
			var nextCursor string
			defer ws.Close()
			for {
				// Start fetching pages of measurements and stream them to the client
				res, _, err := t.client.MeasurementsApi.QueryMeasurements(ctx).
					Cursor(nextCursor).
					Datastream(datastreamID).
					Start(start).
					End(end).Execute()
				if err != nil {
					log.Printf("Failed to fetch devices for client: %v\n", err)
					return
				}

				writer, err := ws.NextWriter(websocket.BinaryMessage)
				if err != nil {
					log.Printf("cannot open writer for ws: %v\n", err)
					return
				}
				defer writer.Close()
				for _, point := range res.Data {
					// Write to client
					binary.Write(writer, binary.BigEndian, point.MeasurementTimestamp.UnixMilli())
					binary.Write(writer, binary.BigEndian, point.MeasurementValue)
				}
				nextCursor = getCursor(res.Links.GetNext())
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

func getCursor(next string) string {
	if next == "" {
		return ""
	}
	u, err := url.Parse(next)
	if err != nil {
		return ""
	}
	return u.Query().Get("cursor")
}

func (t *OverviewRoute) resolveDevice(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deviceID, err := URLParamInt(r, "device_id")
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		res, _, err := t.client.DevicesApi.GetDevice(r.Context(), deviceID).Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		r = r.WithContext(
			context.WithValue(
				r.Context(),
				"device",
				res.Data,
			),
		)
		next.ServeHTTP(w, r)
	})
}

func getDevice(ctx context.Context) (*api.Device, bool) {
	dev, ok := ctx.Value("device").(*api.Device)
	return dev, ok
}

func getSensor(ctx context.Context) (*api.Sensor, bool) {
	dev, ok := ctx.Value("sensor").(*api.Sensor)
	return dev, ok
}

func (t *OverviewRoute) resolveSensor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sensorCode := chi.URLParam(r, "sensor_code")

		device, ok := getDevice(r.Context())
		if !ok {
			web.HTTPError(w, errors.New("resolveSensor middleware is missing device in context, did you use resolveDevice?\n"))
			return
		}

		sensor, ok := lo.Find(device.Sensors, func(item api.Sensor) bool { return item.GetCode() == sensorCode })
		if !ok {
			web.HTTPError(w, web.NewError(http.StatusNotFound, "sensor not found for device", ""))
			return
		}
		r = r.WithContext(
			context.WithValue(
				r.Context(),
				"sensor",
				&sensor,
			),
		)
		next.ServeHTTP(w, r)
	})
}
