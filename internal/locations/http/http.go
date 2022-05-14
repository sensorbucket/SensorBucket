package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"sensorbucket.nl/internal/locations/models"
	"sensorbucket.nl/internal/locations/store"
)

func New() Handler {
	return mux.NewRouter()
}

type Handler interface {
	HandleFunc(path string, f func(http.ResponseWriter, *http.Request)) *mux.Route
	ServeHTTP(http.ResponseWriter, *http.Request)
}

func ListenAndServe(addr string, handler Handler) error {
	return http.ListenAndServe(addr, handler)
}

func GetAllLocations(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	locations, err := store.GetAllLocations()
	if err != nil {
		http.Error(w, "error occured while retrieving locations", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(locations)
}

func GetThingLocationByUrn(w http.ResponseWriter, req *http.Request) {
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
	location, err := store.GetLocationOfThingByUrn(thingUrn)
	if err != nil {
		http.Error(w, "error occured while retrieving location", http.StatusInternalServerError)
		return
	}

	if location.URN == "" {
		w.WriteHeader(http.StatusNoContent)
	}

	json.NewEncoder(w).Encode(location)
}

func CreateLocation(w http.ResponseWriter, req *http.Request) {
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

	locationCheck, err := store.GetLocationByName(locationBody.Name)
	if err != nil {
		http.Error(w, "error occured while accessing storage", http.StatusInternalServerError)
		return
	}
	if locationCheck.Name != "" {
		http.Error(w, "location name is not unique", http.StatusBadRequest)
		return
	}

	err = store.CreateLocation(locationBody)
	if err != nil {
		http.Error(w, "error occured while storing location", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func DeleteLocation(w http.ResponseWriter, req *http.Request) {
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

	locationCheck, err := store.GetLocationById(locationId)
	if err != nil {
		http.Error(w, "error occured while accessing storage", http.StatusInternalServerError)
		return
	}
	if locationCheck.Name == "" {
		http.Error(w, "location does not exist", http.StatusBadRequest)
		return
	}
	err = store.DeleteLocationById(locationId)
	if err != nil {
		http.Error(w, "error occured while deleting location", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func SetLocationOfUrn(w http.ResponseWriter, req *http.Request) {
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

	locationCheck, err := store.GetLocationById(thingLocationBody.LocationId)
	if err != nil {
		http.Error(w, "error occured while accessing storage", http.StatusInternalServerError)
		return
	}
	if locationCheck.Name == "" {
		http.Error(w, "location does not exist", http.StatusBadRequest)
		return
	}

	urnLocation, err := store.GetLocationOfThingByUrn(thingLocationBody.URN)
	if err != nil {
		http.Error(w, "error occured while accessing storage", http.StatusInternalServerError)
		return
	}
	if urnLocation.URN == "" {
		err = store.CreateLocationOfThing(thingLocationBody)
		if err != nil {
			http.Error(w, "error occured while setting location", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	err = store.UpdateLocationOfThing(thingLocationBody)
	if err != nil {
		http.Error(w, "error occured while setting location", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
