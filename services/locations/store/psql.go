package store

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"

	"sensorbucket.nl/services/locations/models"
	"sensorbucket.nl/services/locations/transport"
)

var _ transport.Store = (*Store)(nil)

type Store struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) GetAllLocations() ([]models.Location, error) {
	locations := make([]models.Location, 0)

	if err := s.db.Select(&locations, AllLocations()); err != nil {
		return nil, err
	}

	return locations, nil
}

func (s *Store) GetLocationByName(name string) (*models.Location, error) {
	var location models.Location
	q, p1 := LocationByName(name)
	if err := s.db.Get(&location, q, p1); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, transport.ErrLocationNotFound
		}
		return nil, err
	}
	return &location, nil
}

func (s *Store) GetLocationById(id int) (*models.Location, error) {
	var location models.Location
	q, p1 := LocationById(id)
	if err := s.db.Get(&location, q, p1); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, transport.ErrLocationNotFound
		}
		return nil, err
	}
	return &location, nil
}

func (s *Store) GetLocationOfThingByUrn(thingURN string) (*models.ThingLocation, error) {
	var thingLoc models.ThingLocation
	q, p1 := LocationOfThing(thingURN)

	if err := s.db.Get(&thingLoc, q, p1); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, transport.ErrThingLocationNotFound
		}
		return nil, err
	}
	return &thingLoc, nil
}

func (s *Store) CreateLocation(location models.Location) error {
	_, err := s.db.Exec(InsertLocation(location.Name, location.Latitude, location.Longitude))
	if err != nil {
		if perr, ok := err.(*pgconn.PgError); ok {
			if perr.Code == pgerrcode.UniqueViolation {
				return transport.ErrDuplicateLocationName
			}
		}
	}
	return err
}

func (s *Store) DeleteLocationById(locationId int) error {
	_, err := s.db.Exec(DeleteLocation(locationId))
	return err
}

func (s *Store) DeleteThingLocationByUrn(urn string) error {
	_, err := s.db.Exec(DeleteThingLocation(urn))
	return err
}

func (s *Store) UpdateLocationOfThing(thingURN string, locationID int) error {
	_, err := s.db.Exec(UpdateThingLocation(thingURN, locationID))
	return err
}

func (s *Store) CreateLocationOfThing(thingURN string, locationID int) error {
	_, err := s.db.Exec(InsertThingLocation(thingURN, locationID))
	return err
}

func (s *Store) DeleteThingLocationsByLocationId(locationId int) error {
	_, err := s.db.Exec(DeleteThingLocations(locationId))
	return err
}
