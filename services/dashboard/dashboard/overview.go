package dashboard

//go:generate go get -u github.com/valyala/quicktemplate/qtc
//go:generate qtc -dir=views

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"sensorbucket.nl/sensorbucket/internal/web"
	coretransport "sensorbucket.nl/sensorbucket/services/core/transport"
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
