package store

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"sensorbucket.nl/internal/locations/models"
)

type Store struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) Store {
	return Store{
		db: db,
	}
}

func (s *Store) GetAllLocations() (locations []models.Location, err error) {
	result, err := s.db.Query(AllLocations())
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

func (s *Store) GetLocationByName(name string) (location models.Location, err error) {
	result, err := s.db.Query(LocationByName(name))
	if err != nil {
		return
	}
	defer result.Close()
	if result.Next() {
		err = result.Scan(&location.Id, &location.Name, &location.Lat, &location.Lng)
	}
	return
}

func (s *Store) GetLocationById(id int) (location models.Location, err error) {
	result, err := s.db.Query(LocationById(id))
	if err != nil {
		return
	}
	defer result.Close()
	if result.Next() {
		err = result.Scan(&location.Id, &location.Name, &location.Lat, &location.Lng)
	}
	return
}

func (s *Store) GetLocationOfThingByUrn(urn string) (location models.ThingLocation, err error) {
	result, err := s.db.Query(LocationOfThing(urn))
	if err != nil {
		return
	}
	defer result.Close()
	if result.Next() {
		err = result.Scan(&location.URN, &location.LocationId)
	}
	return
}

func (s *Store) CreateLocation(location models.Location) error {
	return s.exec(InsertLocation(location.Name, location.Lat, location.Lng))
}

func (s *Store) DeleteLocationById(locationId int) error {
	return s.exec(DeleteLocation(locationId))
}

func (s *Store) DeleteThingLocationByUrn(urn string) error {
	return s.exec(DeleteThingLocation(urn))
}

func (s *Store) UpdateLocationOfThing(thingLocation models.ThingLocation) error {
	return s.exec(UpdateThingLocation(thingLocation.URN, thingLocation.LocationId))
}

func (s *Store) CreateLocationOfThing(thingLocation models.ThingLocation) error {
	return s.exec(InsertThingLocation(thingLocation.URN, thingLocation.LocationId))
}

func (s *Store) DeleteThingLocationsByLocationId(locationId int) error {
	return s.exec(DeleteThingLocations(locationId))
}

func (s *Store) exec(cmd string, args ...any) error {
	_, err := s.db.Exec(cmd, args...)
	if err != nil {
		return err
	}
	return nil
}
