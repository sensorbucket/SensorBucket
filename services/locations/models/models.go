package models

type ThingLocation struct {
	ThingURN          string  `json:"thing_urn" db:"thing_urn"`
	LocationID        int64   `json:"location_id" db:"location_id"`
	LocationName      string  `json:"location_name" db:"location_name"`
	LocationLatitude  float64 `json:"location_latitude" db:"location_latitude"`
	LocationLongitude float64 `json:"location_longitude" db:"location_longitude"`
}

type Location struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
