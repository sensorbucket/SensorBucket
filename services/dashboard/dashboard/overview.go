package dashboard

//go:generate go get -u github.com/valyala/quicktemplate/qtc
//go:generate qtc -dir=views

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/core/devices"
	"sensorbucket.nl/sensorbucket/services/core/measurements"
	"sensorbucket.nl/sensorbucket/services/dashboard/dashboard/views"
)

func createOverviewPageHandler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.GetHead)
	r.Get("/datastreams/{id}", overviewDatastream())
	return r
}

func isHX(r *http.Request) bool {
	return r.Header.Get("hx-request") == "true"
}

func URLParamInt(r *http.Request, name string) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, name), 10, 64)
}

func overviewDatastream() http.HandlerFunc {
	getDatastream := func(r *http.Request) (*measurements.Datastream, error) {
		datastreamUUID := chi.URLParam(r, "id")
		res, err := http.Get("http://core/datastreams/" + datastreamUUID)
		if err != nil {
			return nil, err
		}
		var resBody web.APIResponse[measurements.Datastream]
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
		fmt.Printf("ds: %v\n", ds)
		page := &views.DatastreamPage{
			Device: devices.Device{
				ID:                  5,
				Code:                "Test device",
				Description:         "For testing purposes",
				Organisation:        "PZLD",
				State:               devices.DeviceEnabled,
				Sensors:             []devices.Sensor{},
				Properties:          json.RawMessage("{}"),
				Latitude:            lo.ToPtr(53.5),
				Longitude:           lo.ToPtr(3.5),
				Altitude:            lo.ToPtr(2.0),
				LocationDescription: "Ocean",
				CreatedAt:           time.Now(),
			},
		}
		views.WriteIndex(w, page)
	}
}
