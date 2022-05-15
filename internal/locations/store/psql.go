package store

import (
	"database/sql"

	_ "github.com/lib/pq"
	"sensorbucket.nl/internal/locations/models"
)

var ConnString = ""

func GetAllLocations() (locations []models.Location, err error) {
	result, err := query(AllLocations())
	if err != nil {
		return
	}
	defer result.Close()
	for result.Next() {
		loc := models.Location{}
		err = result.Scan(
			&loc.Id,
			&loc.Name,
			&loc.Lat,
			&loc.Lng,
		)
		if err != nil {
			return
		}
		locations = append(locations, loc)
	}
	return
}

func GetLocationByName(name string) (location models.Location, err error) {
	result, err := query(LocationByName(name))
	if err != nil {
		return
	}
	defer result.Close()
	if result.Next() {
		err = result.Scan(&location.Id, &location.Name, &location.Lat, &location.Lng)
	}
	return
}

func GetLocationById(id int) (location models.Location, err error) {
	result, err := query(LocationById(id))
	if err != nil {
		return
	}
	defer result.Close()
	if result.Next() {
		err = result.Scan(&location.Id, &location.Name, &location.Lat, &location.Lng)
	}
	return
}

func GetLocationOfThingByUrn(urn string) (location models.ThingLocation, err error) {
	result, err := query(LocationOfThing(urn))
	if err != nil {
		return
	}
	defer result.Close()
	if result.Next() {
		err = result.Scan(&location.URN, &location.LocationId)
	}
	return
}

func CreateLocation(location models.Location) error {
	return exec(InsertLocation(location.Name, location.Lat, location.Lng))
}

func DeleteLocationById(locationId int) error {
	return exec(DeleteLocation(locationId))
}

func DeleteThingLocationByUrn(urn string) error {
	return exec(DeleteThingLocation(urn))
}

func UpdateLocationOfThing(thingLocation models.ThingLocation) error {
	return exec(UpdateThingLocation(thingLocation.URN, thingLocation.LocationId))
}

func CreateLocationOfThing(thingLocation models.ThingLocation) error {
	return exec(InsertThingLocation(thingLocation.URN, thingLocation.LocationId))
}

func exec(cmd string, args ...any) error {
	db, err := sql.Open("postgres", ConnString)
	if err != nil {
		return err
	}

	_, err = db.Exec(cmd, args...)
	if err != nil {
		return err
	}
	return nil
}

func query(query string, args ...any) (*sql.Rows, error) {
	db, err := sql.Open("postgres", ConnString)
	if err != nil {
		return nil, err
	}

	res, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	return res, err
}
