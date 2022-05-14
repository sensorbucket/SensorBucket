package models

type ThingLocation struct {
	URN        string
	LocationId int
}

type Location struct {
	Id   int64
	Name string
	Lat  float64
	Lng  float64
}
