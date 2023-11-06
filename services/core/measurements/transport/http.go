package measurementtransport

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"sensorbucket.nl/sensorbucket/internal/httpfilter"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/core/measurements"
)

// HTTPTransport exposes API endpoints to query measurements.
type HTTPTransport struct {
	router chi.Router
	svc    *measurements.Service
	url    string
}

func NewHTTP(svc *measurements.Service, url string) *HTTPTransport {
	t := &HTTPTransport{
		router: chi.NewRouter(),
		svc:    svc,
		url:    url,
	}
	t.SetupRoutes(t.router)
	return t
}

func (t *HTTPTransport) SetupRoutes(r chi.Router) {
	r.Get("/measurements", t.httpGetMeasurements())
	r.Get("/stream-measurements", t.httpStreamMeasurements())
	r.Get("/datastreams", t.httpListDatastream())
	r.Get("/datastreams/{id}", t.httpGetDatastream())
}

func (t *HTTPTransport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.router.ServeHTTP(w, r)
}

func (t *HTTPTransport) httpGetMeasurements() http.HandlerFunc {
	type Params struct {
		measurements.Filter
		pagination.Request
	}
	return func(rw http.ResponseWriter, r *http.Request) {
		params, err := httpfilter.Parse[Params](r)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}

		if !params.Start.IsZero() && !params.End.IsZero() && params.Start.After(params.End) {
			web.HTTPError(rw, web.NewError(http.StatusBadRequest, "Start time cannot be after end time", "ERR_BAD_REQUEST"))
			return
		}

		page, err := t.svc.QueryMeasurements(params.Filter, params.Request)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		web.HTTPResponse(rw, http.StatusOK, pagination.CreateResponse(r, t.url, *page))
	}
}

func (t *HTTPTransport) httpStreamMeasurements() http.HandlerFunc {
	type Params struct {
		measurements.Filter
	}
	return func(rw http.ResponseWriter, r *http.Request) {
		params, err := httpfilter.Parse[Params](r)
		if err != nil {
			web.HTTPError(rw, web.NewError(http.StatusBadRequest, "Invalid params", "ERR_INVALID_PARAMS"))
			return
		}

		if !params.Start.IsZero() && !params.End.IsZero() && params.Start.After(params.End) {
			web.HTTPError(rw, web.NewError(http.StatusBadRequest, "Start time cannot be after end time", "ERR_BAD_REQUEST"))
			return
		}

		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}
		ws, err := upgrader.Upgrade(rw, r, nil)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}

		go func() {
			fmt.Println("connected, sending measurements")
			stream := t.svc.StreamMeasurements(params.Filter)

			defer func(ws *websocket.Conn) {
				err := ws.Close()
				if err != nil {
					log.Println("[Warning] couldn't properly close websocket connection")
				}
			}(ws)
			for {
				measurement, ok := <-stream
				if !ok {
					// Measurement stream has been closed
					return
				}
				writer, err := ws.NextWriter(websocket.TextMessage)
				if err != nil {
					log.Printf("cannot open writer for ws: %v\n", err)
					return
				}
				fmt.Println("got measurement, sending over ws")
				b, err := json.Marshal(&measurement)
				if err != nil {
					log.Printf("erred while marshalling measurement: %v\n", err)
					return
				}
				_, err = writer.Write(b)
				if err != nil {
					log.Printf("erred while sending frame to ws: %v\n", err)
					return
				}
			}
		}()
	}
}

func (t *HTTPTransport) httpListDatastream() http.HandlerFunc {
	type params struct {
		measurements.DatastreamFilter
		pagination.Request
	}
	return func(rw http.ResponseWriter, r *http.Request) {
		params, err := httpfilter.Parse[params](r)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}

		page, err := t.svc.ListDatastreams(r.Context(), params.DatastreamFilter, params.Request)
		if err != nil {
			web.HTTPError(rw, err)
			return
		}
		web.HTTPResponse(rw, http.StatusOK, pagination.CreateResponse(r, t.url, *page))
	}
}

func (t *HTTPTransport) httpGetDatastream() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idQ := chi.URLParam(r, "id")
		id, err := uuid.Parse(idQ)
		if err != nil {
			web.HTTPError(w, web.NewError(http.StatusBadRequest, "Invalid datastream ID", ""))
			return
		}

		ds, err := t.svc.GetDatastream(r.Context(), id)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{Data: ds})
	}
}
