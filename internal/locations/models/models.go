package models

type ThingLocation struct {
	URN        string `json:"urn"`
	LocationId int    `json:"location_id"`
}

type Location struct {
	Id   int64   `json:"id"`
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
}
