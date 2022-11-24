package service

import (
	"net/http"
	"regexp"

	"sensorbucket.nl/sensorbucket/internal/web"
)

var (
	ErrInvalidCoordinates = web.NewError(
		http.StatusBadRequest,
		"Invalid coordinates supplied",
		"ERR_LOCATION_INVALID_COORDINATES",
	)
	ErrInvalidLocationName = web.NewError(
		http.StatusBadRequest,
		"Invalid location name. Must be a-zA-Z0-9 and not start with '-' or '_'",
		"ERR_LOCATION_INVALID_NAME",
	)

	_R_LOC_NAME = "^[a-zA-Z][a-zA-Z0-9-_]+$"
	R_LOC_NAME  = regexp.MustCompile(_R_LOC_NAME)
)

type Location struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Organisation string  `json:"organisation"`
	Longitude    float64 `json:"longitude"`
	Latitude     float64 `json:"latitude"`
}

type NewLocationOpts struct {
	Name         string  `json:"name"`
	Organisation string  `json:"organisation"`
	Longitude    float64 `json:"longitude"`
	Latitude     float64 `json:"latitude"`
}

func NewLocation(opts NewLocationOpts) (*Location, error) {
	if opts.Latitude < -90 || opts.Latitude > 90 || opts.Longitude < -180 || opts.Longitude > 180 {
		return nil, ErrInvalidCoordinates
	}

	if !R_LOC_NAME.MatchString(opts.Name) {
		return nil, ErrInvalidLocationName
	}

	return &Location{
		Name:         opts.Name,
		Organisation: opts.Organisation,
		Longitude:    opts.Longitude,
		Latitude:     opts.Latitude,
	}, nil
}
