package main

import "sensorbucket.nl/internal/locations/http"

func main() {
	// I must be able to create/remove locations
	// I must be able to bind one or more assets to a location
	// I must be able to remove assets from a location
	// Incoming measurements must be supplemented with a location_urn if available

	router := http.New()
	router.HandleFunc("/api/location/create", http.CreateLocation)
	router.HandleFunc("/api/location/delete", http.DeleteLocation)
	router.HandleFunc("/api/location/all", http.GetAllLocations)
	router.HandleFunc("/api/location/urn", http.GetThingLocationByUrn)
	router.HandleFunc("/api/location/patch", http.SetLocationOfUrn)
	http.ListenAndServe(":8080", router)
}
