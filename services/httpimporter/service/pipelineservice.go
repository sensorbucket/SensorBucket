package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/core/processing"
)

var _ PipelineService = (*PipelineServiceHTTP)(nil)

type PipelineServiceHTTP struct {
	host string
}

func NewPipelineServiceHTTP(serviceHost string) *PipelineServiceHTTP {
	return &PipelineServiceHTTP{host: serviceHost}
}

func (p *PipelineServiceHTTP) Get(uuid string) (*processing.Pipeline, error) {
	res, err := http.Get(fmt.Sprintf("%s/pipelines/%s", p.host, uuid))
	if err != nil {
		return nil, fmt.Errorf("could not get pipeline definition: %w", err)
	}
	defer res.Body.Close()

	// if error status, then pipeline service should have responded with APIError
	// We forward that error to the requester.
	if res.StatusCode < 200 || res.StatusCode > 299 {
		var err web.APIError
		if err := json.NewDecoder(res.Body).Decode(&err); err != nil {
			return nil, fmt.Errorf("could not read pipline service response: %w", err)
		}
		err.HTTPStatus = res.StatusCode
		log.Printf("Error status: %v\n", err)
		return nil, &err
	}

	var body web.APIResponse[processing.Pipeline]
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("could not parse pipeline service response: %w", err)
	}

	return &body.Data, nil
}
