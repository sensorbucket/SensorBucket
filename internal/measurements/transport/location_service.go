package transport

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	locationModels "sensorbucket.nl/internal/locations/models"
	"sensorbucket.nl/internal/measurements"
)

var _ measurements.LocationService = (*LocationService)(nil)

// LocationService
type LocationService struct {
	base string
}

func NewLocationService(baseURL string) *LocationService {
	return &LocationService{
		base: baseURL,
	}
}

// DTO
type DTO struct {
	LocationId int `json:"location_id"`
}

func (l *LocationService) FindLocationID(thingURN string) (measurements.LocationData, error) {
	var location measurements.LocationData

	req, err := http.NewRequest(http.MethodGet, l.base, nil)
	if err != nil {
		return location, err
	}

	req.URL.Path = "/locations/things/" + url.PathEscape(thingURN)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return location, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return location, nil
	}
	if resp.StatusCode != http.StatusOK {
		return location, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var dto locationModels.ThingLocation
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return location, err
	}

	return measurements.LocationData{
		ID:        dto.LocationID,
		Name:      dto.LocationName,
		Longitude: dto.LocationLongitude,
		Latitude:  dto.LocationLatitude,
	}, nil
}
