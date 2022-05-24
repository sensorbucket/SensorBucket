package transport

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"sensorbucket.nl/internal/locations/models"
)

var (
	ErrLocationNotFound      = errors.New("location not found")
	ErrThingLocationNotFound = errors.New("location for thing not found")
	ErrDuplicateLocationName = errors.New("location name is already is use")
)

type Store interface {
	GetAllLocations() ([]models.Location, error)
	GetLocationByName(name string) (*models.Location, error)
	GetLocationById(id int) (*models.Location, error)
	GetLocationOfThingByUrn(thingURN string) (*models.ThingLocation, error)
	CreateLocation(location models.Location) error
	DeleteLocationById(locationId int) error
	DeleteThingLocationByUrn(urn string) error
	UpdateLocationOfThing(thingURN string, locationID int) error
	CreateLocationOfThing(thingURN string, locationID int) error
	DeleteThingLocationsByLocationId(locationId int) error
}

type apiResponse struct {
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type HTTPTransport struct {
	router http.Handler
	store  Store
}

func New(store Store) HTTPTransport {
	r := chi.NewRouter()
	t := HTTPTransport{
		router: r,
		store:  store,
	}

	r.Post("/locations", t.CreateLocation())
	r.Delete("/location/{locationID}", t.DeleteLocation())
	r.Get("/locations", t.GetLocations())
	r.Get("/locations/things/{thingURN}", t.GetLocationForThing())
	r.Delete("/locations/things/{thingURN}", t.DeleteThingLocation())
	r.Put("/locations/things", t.SetThingLocation())
	return t
}

func (t HTTPTransport) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	t.router.ServeHTTP(rw, r)
}

func (t *HTTPTransport) GetLocations() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		locations, err := t.store.GetAllLocations()
		if err != nil {
			logrus.WithError(err).Info("erred while interacting with db")
			http.Error(w, "error occured while retrieving locations", http.StatusInternalServerError)
			return
		}

		sendJSON(w, http.StatusOK, apiResponse{
			Message: "Locations listed",
			Data:    locations,
		})
	}
}

func (r *HTTPTransport) GetLocationForThing() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		thingURN := urlParam(req, "thingURN")

		location, err := r.store.GetLocationOfThingByUrn(thingURN)
		if errors.Is(err, ErrThingLocationNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			logrus.WithError(err).Info("erred while retrieving location")
			http.Error(w, "error occured while retrieving location", http.StatusInternalServerError)
			return
		}

		sendJSON(w, http.StatusOK, apiResponse{Message: "Location fetched", Data: location})
	}
}

func (t *HTTPTransport) CreateLocation() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var location models.Location
		if err := json.NewDecoder(req.Body).Decode(&location); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if location.Name == "" {
			http.Error(w, "invalid request body, name cannot be empty", http.StatusBadRequest)
			return
		}

		err := t.store.CreateLocation(location)
		if err != nil {
			logrus.WithError(err).Info("erred creating location")
			if errors.Is(err, ErrDuplicateLocationName) {
				http.Error(w, "location name already in use", http.StatusInternalServerError)
				return
			}
			http.Error(w, "error occured while storing location", http.StatusInternalServerError)
			return
		}

		sendJSON(w, http.StatusCreated, apiResponse{Message: "Location was succesfully created", Data: location})
	}
}

func (t *HTTPTransport) DeleteLocation() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		locationId, err := strconv.Atoi(urlParam(req, "locationID"))
		if err != nil {
			http.Error(w, "locationID must be an integer", http.StatusBadRequest)
			return
		}

		locationCheck, err := t.store.GetLocationById(locationId)
		if err != nil {
			logrus.WithError(err).Info("erred while interacting with db")
			http.Error(w, "error occured while accessing storage", http.StatusInternalServerError)
			return
		}
		if locationCheck.Name == "" {
			http.Error(w, "location does not exist", http.StatusBadRequest)
			return
		}

		err = t.store.DeleteThingLocationsByLocationId(int(locationCheck.ID))
		if err != nil {
			logrus.WithError(err).Info("erred while interacting with db")
			http.Error(w, "error occured while deleting location", http.StatusInternalServerError)
			return
		}
		err = t.store.DeleteLocationById(locationId)
		if err != nil {
			logrus.WithError(err).Info("erred while interacting with db")
			http.Error(w, "error occured while deleting location", http.StatusInternalServerError)
			return
		}
		sendJSON(w, http.StatusOK, apiResponse{Message: "Location was succesfully deleted"})
	}
}

func (r *HTTPTransport) DeleteThingLocation() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		thingURN := urlParam(req, "thingURN")
		err := r.store.DeleteThingLocationByUrn(thingURN)
		if err != nil {
			logrus.WithError(err).Info("erred while interacting with db")
			http.Error(w, "error occured while deleting location", http.StatusInternalServerError)
			return
		}

		sendJSON(w, http.StatusOK, apiResponse{Message: "Thing was succesfully deleted from location"})
	}
}

func (t *HTTPTransport) SetThingLocation() http.HandlerFunc {
	type DTO struct {
		ThingURN   string `json:"thing_urn"`
		LocationID int    `json:"location_id"`
	}
	return func(w http.ResponseWriter, req *http.Request) {
		var dto DTO
		if err := json.NewDecoder(req.Body).Decode(&dto); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if dto.ThingURN == "" || dto.LocationID == 0 {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		_, err := t.store.GetLocationById(dto.LocationID)
		if errors.Is(err, ErrLocationNotFound) {
			http.Error(w, "location does not exist", http.StatusBadRequest)
			return
		}
		if err != nil {
			logrus.WithError(err).Info("erred while checking whether location id exists")
			http.Error(w, "error occured while accessing storage", http.StatusInternalServerError)
			return
		}

		_, err = t.store.GetLocationOfThingByUrn(dto.ThingURN)
		if errors.Is(err, ErrThingLocationNotFound) {
			err = t.store.CreateLocationOfThing(dto.ThingURN, dto.LocationID)
			if err != nil {
				logrus.WithError(err).Info("erred while creating new thing_location")
				http.Error(w, "error occured while setting location", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		if err != nil {
			logrus.WithError(err).Info("erred while checking if thing_location relation already exists")
			http.Error(w, "error occured while accessing storage", http.StatusInternalServerError)
			return
		}

		err = t.store.UpdateLocationOfThing(dto.ThingURN, dto.LocationID)
		if err != nil {
			logrus.WithError(err).Info("erred while updating thing_location")
			http.Error(w, "error occured while setting location", http.StatusInternalServerError)
			return
		}
		sendJSON(w, http.StatusOK, apiResponse{Message: "Thing was succesfully added to location"})
	}
}

func sendJSON(rw http.ResponseWriter, status int, v interface{}) {
	body, err := json.Marshal(v)
	if err != nil {
		http.Error(rw, "error encoding response", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("content-type", "application/json")
	rw.WriteHeader(status)
	rw.Write(body)
}

func urlParam(r *http.Request, key string) string {
	q := chi.URLParam(r, key)
	qu, _ := url.QueryUnescape(q)
	return qu
}
