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

	"sensorbucket.nl/sensorbucket/internal/flash_messages"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/api"
	"sensorbucket.nl/sensorbucket/services/dashboard/views"
)

type OverviewRoute struct {
	router     chi.Router
	coreClient *api.APIClient
}

func CreateOverviewPageHandler(core *api.APIClient) *OverviewRoute {
	t := &OverviewRoute{
		coreClient: core,
		router:     chi.NewRouter(),
	}
	t.SetupRoutes(t.router)
	return t
}

func (t OverviewRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.router.ServeHTTP(w, r)
}

func (t *OverviewRoute) SetupRoutes(r chi.Router) {
	r.Use(
		middleware.GetHead,
		flash_messages.ExtractFlashMessage,
	)
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
			res, _, err := t.coreClient.DevicesApi.GetSensorGroup(r.Context(), sgID).Execute()
			if err != nil {
				web.HTTPError(w, err)
				return
			}
			sg = res.Data
		}
		req := t.coreClient.DevicesApi.ListDevices(r.Context())
		if sg != nil {
			req = req.SensorGroup([]int64{sg.GetId()})
		}
		res, _, err := req.Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		w.Header().Set("hx-push-url", views.U("/overview?%s", r.URL.Query().Encode()))
		w.Header().Set("hx-trigger-after-settle", "newDeviceList")
		views.WriteRenderFilters(w, sg, true)
		views.WriteRenderDeviceTable(w, res.Data, getCursor(res.Links.GetNext()))
	}
}

func (t *OverviewRoute) deleteSensorGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, _, err := t.coreClient.DevicesApi.ListDevices(r.Context()).Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		w.Header().Set("hx-push-url", views.U("/overview?%s", r.URL.Query().Encode()))
		w.Header().Set("hx-trigger-after-settle", "newDeviceList")
		views.WriteRenderFilters(w, nil, true)
		views.WriteRenderDeviceTable(w, res.Data, getCursor(res.Links.GetNext()))
	}
}

func (t *OverviewRoute) getDevicesTable() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := t.coreClient.DevicesApi.ListDevices(r.Context())
		if r.URL.Query().Has("sensor_group") {
			sgID, err := strconv.ParseInt(r.URL.Query().Get("sensor_group"), 10, 64)
			if err != nil {
				web.HTTPError(w, err)
				return
			}
			req = req.SensorGroup([]int64{sgID})
		}
		if r.URL.Query().Has("cursor") {
			req = req.Cursor(r.URL.Query().Get("cursor"))
		}
		res, _, err := req.Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		nextCursor := ""
		if res.Links.GetNext() != "" {
			nextCursor = views.U("/overview/devices/table?cursor=%s", getCursor(res.Links.GetNext()))
		}
		views.WriteRenderDeviceTableRows(w, res.Data, nextCursor)
	}
}

func (t *OverviewRoute) searchSensorGroups() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, _, err := t.coreClient.DevicesApi.ListSensorGroups(r.Context()).Execute()
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
		page := &views.DeviceListPage{
			BasePage: createBasePage(r),
		}

		sensorGroupIDStr := r.URL.Query().Get("sensor_group")
		if sensorGroupIDStr != "" {
			sensorGroupID, err := strconv.ParseInt(sensorGroupIDStr, 10, 64)
			if err != nil {
				web.HTTPError(w, web.NewError(http.StatusBadRequest, "Sensor Group ID is not an integer", "ERR_BAD_REQUEST"))
				return
			}
			res, _, err := t.coreClient.DevicesApi.GetSensorGroup(r.Context(), sensorGroupID).Execute()
			if err != nil {
				web.HTTPError(w, fmt.Errorf("error getting sensor group: %w", err))
				return
			}
			page.SensorGroup = res.Data
		}
		req := t.coreClient.DevicesApi.ListDevices(r.Context())
		if page.SensorGroup != nil {
			req = req.SensorGroup([]int64{page.SensorGroup.GetId()})
		}
		res, _, err := req.Execute()
		if err != nil {
			web.HTTPError(w, fmt.Errorf("error listing devices: %w", err))
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
			BasePage: createBasePage(r),
			Device:   *device,
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

		res, _, err := t.coreClient.MeasurementsApi.ListDatastreams(r.Context()).Sensor([]int64{sensor.Id}).Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		page := &views.SensorDetailPage{
			BasePage:    createBasePage(r),
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

	return func(w http.ResponseWriter, r *http.Request) {
		var sensorGroups []int64
		if r.URL.Query().Has("sensor_group") {
			for _, idString := range r.URL.Query()["sensor_group"] {
				id, err := strconv.ParseInt(idString, 10, 64)
				if err != nil {
					web.HTTPError(w, err)
					return
				}
				sensorGroups = append(sensorGroups, id)
			}
		}

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		// Build websocket context and copy auth token
		wsCTX, cancel := context.WithCancel(context.Background())
		wsCTX = context.WithValue(wsCTX, api.ContextAccessToken, r.Context().Value(api.ContextAccessToken))
		go func(ctx context.Context) {
			defer cancel()
			defer ws.Close()
			var nextCursor string
			for {
				// Start fetching pages of devices and stream them to the client
				res, _, err := t.coreClient.DevicesApi.ListDevices(ctx).Cursor(nextCursor).SensorGroup(sensorGroups).Execute()
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
					if _, err := writer.Write([]byte(frame)); err != nil {
						log.Printf("Failed to write to websocket: %v\n", err)
						return
					}
				}
				nextCursor = getCursor(res.Links.GetNext())
				if nextCursor == "" {
					return
				}
			}
		}(wsCTX)
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

		res, _, err := t.coreClient.MeasurementsApi.GetDatastream(r.Context(), chi.URLParam(r, "id")).Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		page := &views.DatastreamPage{
			BasePage:   createBasePage(r),
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

		// Build websocket context and copy auth token
		wsCTX := context.WithValue(context.Background(), api.ContextAccessToken, r.Context().Value(api.ContextAccessToken))
		go func(ctx context.Context) {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			var nextCursor string
			defer ws.Close()
			for {
				// Stop if the context is canceled
				select {
				case <-ctx.Done():
					return
				default:
				}
				// Start fetching pages of measurements and stream them to the client
				res, _, err := t.coreClient.MeasurementsApi.QueryMeasurements(ctx).
					Cursor(nextCursor).
					Datastream(datastreamID).
					Start(start).
					Limit(1000). // Currently maximum allowed limit
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
				for _, point := range res.Data {
					// Write to client
					if err := binary.Write(writer, binary.BigEndian, point.MeasurementTimestamp.UnixMilli()); err != nil {
						log.Printf("Error writing measurement timestamp to WebSocket: %v\n", err)
					}
					if err := binary.Write(writer, binary.BigEndian, point.MeasurementValue); err != nil {
						log.Printf("Error writing measurement value to WebSocket: %v\n", err)
					}
				}
				if err := writer.Close(); err != nil {
					cancel()
				}

				nextCursor = getCursor(res.Links.GetNext())
				if nextCursor == "" {
					return
				}
			}
		}(wsCTX)
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
		res, _, err := t.coreClient.DevicesApi.GetDevice(r.Context(), deviceID).Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		r = r.WithContext(
			context.WithValue(
				r.Context(),
				ctxDevice,
				res.Data,
			),
		)
		next.ServeHTTP(w, r)
	})
}

func getDevice(ctx context.Context) (*api.Device, bool) {
	dev, ok := ctx.Value(ctxDevice).(*api.Device)
	return dev, ok
}

func getSensor(ctx context.Context) (*api.Sensor, bool) {
	dev, ok := ctx.Value(ctxSensor).(*api.Sensor)
	return dev, ok
}

func (t *OverviewRoute) resolveSensor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sensorCode := chi.URLParam(r, "sensor_code")

		device, ok := getDevice(r.Context())
		if !ok {
			web.HTTPError(w, errors.New("resolveSensor middleware is missing device in context, you probably didn't use resolveDevice middleware"))
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
				ctxSensor,
				&sensor,
			),
		)
		next.ServeHTTP(w, r)
	})
}
