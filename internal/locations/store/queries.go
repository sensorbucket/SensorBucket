package store

// Retrieves in order of id, name, lat, lng
func AllLocations() string {
	return "SELECT id, name, lat, lng FROM locations;"
}

// Retrieves in order of urn, location_Id
func LocationOfThing(thingURN string) (string, string) {
	return "SELECT tl.thing_urn, tl.location_id, locs.name AS location_name, locs.lat AS location_latitude, locs.lng AS location_longitude FROM thing_locations tl LEFT JOIN locations locs ON tl.location_id=locs.id WHERE thing_urn = $1;", thingURN
}

// Retrieves in order of id, name, lat, lng
func LocationByName(name string) (string, string) {
	return "SELECT id, name, lat, lng FROM locations WHERE name=$1;", name
}

func LocationById(id int) (string, int) {
	return "SELECT id, name, lat, lng FROM locations WHERE id=$1;", id
}

func InsertLocation(name string, lat float64, lng float64) (string, string, float64, float64) {
	return "INSERT INTO locations (name, lat, lng) VALUES ($1, $2, $3);", name, lat, lng
}

func InsertThingLocation(thingURN string, location_id int) (string, string, int) {
	return "INSERT INTO thing_locations (thing_urn, location_id) VALUES ($1, $2);", thingURN, location_id
}

func DeleteThingLocation(thingURN string) (string, string) {
	return "DELETE FROM thing_locations WHERE thing_urn = $1;", thingURN
}

func DeleteThingLocations(locationId int) (string, int) {
	return "DELETE FROM thing_locations WHERE location_id = $1;", locationId
}

func UpdateThingLocation(urn string, locationId int) (string, int, string) {
	return "UPDATE thing_locations SET location_id = $1 WHERE thing_urn = $2;", locationId, urn
}

func DeleteLocation(locationId int) (string, int) {
	return "DELETE FROM locations WHERE id = $1;", locationId
}
