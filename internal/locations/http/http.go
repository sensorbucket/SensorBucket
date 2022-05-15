package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"sensorbucket.nl/internal/locations/models"
	"sensorbucket.nl/internal/locations/store"
)

type Router struct {
	handler http.Handler
	store   store.Store
}

func New(store store.Store) Router {
	muxRouter := mux.NewRouter()
	router := Router{
		handler: muxRouter,
		store:   store,
	}

	muxRouter.HandleFunc("/api/location/create", router.CreateLocation)
	muxRouter.HandleFunc("/api/location/delete", router.CreateLocation)
	muxRouter.HandleFunc("/api/location/delete", router.DeleteLocation)
	muxRouter.HandleFunc("/api/location/all", router.GetAllLocations)
	muxRouter.HandleFunc("/api/location/thing", router.GetThingLocationByUrn)
	muxRouter.HandleFunc("/api/location/thing/delete", router.DeleteLocationOfURN)
	muxRouter.HandleFunc("/api/location/thing/set", router.SetLocationOfUrn)
	return router
}

func (r *Router) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, r.handler)
}

func (r *Router) GetAllLocations(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	locations, err := r.store.GetAllLocations()
	if err != nil {
		logrus.WithError(err).Info("erred while interacting with db")
		http.Error(w, "error occured while retrieving locations", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(locations)
}

func (r *Router) GetThingLocationByUrn(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	q := req.URL.Query()
	if len(q["thing_urn"]) != 1 || q["thing_urn"][0] == "" {
		http.Error(w, "invalid query parameters", http.StatusBadRequest)
		return
	}

	thingUrn := q["thing_urn"][0]
	location, err := r.store.GetLocationOfThingByUrn(thingUrn)
	if err != nil {
		logrus.WithError(err).Info("erred while interacting with db")
		http.Error(w, "error occured while retrieving location", http.StatusInternalServerError)
		return
	}

	if location.URN == "" {
		w.WriteHeader(http.StatusNoContent)
	}

	json.NewEncoder(w).Encode(location)
}

func (r *Router) CreateLocation(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	locationBody := models.Location{}
	err = json.Unmarshal(body, &locationBody)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if locationBody.Name == "" {
		http.Error(w, "invalid request body, name cannot be empty", http.StatusBadRequest)
		return
	}

	locationCheck, err := r.store.GetLocationByName(locationBody.Name)
	if err != nil {
		logrus.WithError(err).Info("erred while interacting with db")
		http.Error(w, "error occured while accessing storage", http.StatusInternalServerError)
		return
	}
	if locationCheck.Name != "" {
		http.Error(w, "location name is not unique", http.StatusBadRequest)
		return
	}

	err = r.store.CreateLocation(locationBody)
	if err != nil {
		logrus.WithError(err).Info("erred while interacting with db")
		http.Error(w, "error occured while storing location", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (r *Router) DeleteLocation(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	q := req.URL.Query()
	if len(q["location_id"]) != 1 || q["location_id"][0] == "" {
		http.Error(w, "invalid query parameters", http.StatusBadRequest)
		return
	}

	locationId, err := strconv.Atoi(q["location_id"][0])
	if err != nil {
		http.Error(w, "invalid query parameters, location_id was not a valid number", http.StatusBadRequest)
		return
	}

	locationCheck, err := r.store.GetLocationById(locationId)
	if err != nil {
		logrus.WithError(err).Info("erred while interacting with db")
		http.Error(w, "error occured while accessing storage", http.StatusInternalServerError)
		return
	}
	if locationCheck.Name == "" {
		http.Error(w, "location does not exist", http.StatusBadRequest)
		return
	}

	err = r.store.DeleteThingLocationsByLocationId(int(locationCheck.Id))
	if err != nil {
		logrus.WithError(err).Info("erred while interacting with db")
		http.Error(w, "error occured while deleting location", http.StatusInternalServerError)
		return
	}
	err = r.store.DeleteLocationById(locationId)
	if err != nil {
		logrus.WithError(err).Info("erred while interacting with db")
		http.Error(w, "error occured while deleting location", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (r *Router) DeleteLocationOfURN(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	q := req.URL.Query()
	if len(q["thing_urn"]) != 1 || q["thing_urn"][0] == "" {
		http.Error(w, "invalid query parameters", http.StatusBadRequest)
		return
	}

	thingUrn := q["thing_urn"][0]
	err := r.store.DeleteThingLocationByUrn(thingUrn)
	if err != nil {
		logrus.WithError(err).Info("erred while interacting with db")
		http.Error(w, "error occured while deleting location", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (r *Router) SetLocationOfUrn(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	thingLocationBody := models.ThingLocation{}
	err = json.Unmarshal(body, &thingLocationBody)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if thingLocationBody.URN == "" || thingLocationBody.LocationId == 0 {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	locationCheck, err := r.store.GetLocationById(thingLocationBody.LocationId)
	if err != nil {
		logrus.WithError(err).Info("erred while interacting with db")
		http.Error(w, "error occured while accessing storage", http.StatusInternalServerError)
		return
	}
	if locationCheck.Name == "" {
		http.Error(w, "location does not exist", http.StatusBadRequest)
		return
	}

	urnLocation, err := r.store.GetLocationOfThingByUrn(thingLocationBody.URN)
	if err != nil {
		logrus.WithError(err).Info("erred while interacting with db")
		http.Error(w, "error occured while accessing storage", http.StatusInternalServerError)
		return
	}
	if urnLocation.URN == "" {
		err = r.store.CreateLocationOfThing(thingLocationBody)
		if err != nil {
			logrus.WithError(err).Info("erred while interacting with db")
			http.Error(w, "error occured while setting location", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	err = r.store.UpdateLocationOfThing(thingLocationBody)
	if err != nil {
		logrus.WithError(err).Info("erred while interacting with db")
		http.Error(w, "error occured while setting location", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
