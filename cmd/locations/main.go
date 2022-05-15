package main

import (
	"os"

	"sensorbucket.nl/internal/locations/http"
	"sensorbucket.nl/internal/locations/store"
)

const (
	WORKER_HTTP_HOST = "WORKER_HTTP_HOST" // :8080
	WORKER_DB_CONN   = "WORKER_DB_CONN"   // postgresql://root:root@localhost:5432/todos?sslmode=disable
)

func main() {
	store.ConnString = os.Getenv(WORKER_DB_CONN)
	router := http.New()
	router.HandleFunc("/api/location/create", http.CreateLocation)
	router.HandleFunc("/api/location/delete", http.DeleteLocation)
	router.HandleFunc("/api/location/all", http.GetAllLocations)
	router.HandleFunc("/api/location/thing", http.GetThingLocationByUrn)
	router.HandleFunc("/api/location/thing/delete", http.DeleteLocationOfURN)
	router.HandleFunc("/api/location/thing/set", http.SetLocationOfUrn)
	http.ListenAndServe(os.Getenv(WORKER_HTTP_HOST), router)
}
