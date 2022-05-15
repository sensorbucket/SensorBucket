package store

// Retrieves in order of id, name, lat, lng
func AllLocations() string {
	return "SELECT id, name, lat, lng FROM locations;"
}

// Retrieves in order of urn, location_Id
func LocationOfThing(urn string) (string, string) {
	return "SELECT urn, location_id FROM thing_locations WHERE urn = $1;", urn
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

func InsertThingLocation(urn string, location_id int) (string, string, int) {
	return "INSERT INTO thing_locations (urn, location_id) VALUES ($1, $2);", urn, location_id
}

func DeleteThingLocation(urn string) (string, string) {
	return "DELETE FROM thing_locations WHERE urn = $1;", urn
}

func DeleteThingLocations(locationId int) (string, int) {
	return "DELETE FROM thing_locations WHERE location_id = $1;", locationId
}

func UpdateThingLocation(urn string, locationId int) (string, int, string) {
	return "UPDATE thing_locations SET location_id = $1 WHERE urn = $2;", locationId, urn
}

func DeleteLocation(locationId int) (string, int) {
	return "DELETE FROM locations WHERE id = $1;", locationId
}
