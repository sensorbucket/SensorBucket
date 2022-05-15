package transport

import (
	"encoding/json"
	"fmt"
	"net/http"

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

func (l *LocationService) FindLocationID(thingURN string) (*int, error) {
	req, err := http.NewRequest(http.MethodGet, l.base, nil)
	if err != nil {
		return nil, err
	}

	req.URL.Path = "/locations/thing"
	q := req.URL.Query()
	q.Set("thing_urn", thingURN)
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var dto DTO
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return nil, err
	}

	return &dto.LocationId, nil
}
