package dashboardinfra

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/core/processing"
	"sensorbucket.nl/sensorbucket/services/dashboard/routes"
	ingressarchiver "sensorbucket.nl/sensorbucket/services/tracing/ingress-archiver/service"
)

type SensorBucketAPI struct {
	ingressEndpoint   string
	pipelinesEndpoint string
	tracesEndpoint    string
}

func NewSensorBucketAPI(ingressEndpoint, pipelinesEndpoint, tracesEndpoint string) *SensorBucketAPI {
	return &SensorBucketAPI{
		ingressEndpoint:   ingressEndpoint,
		pipelinesEndpoint: pipelinesEndpoint,
		tracesEndpoint:    tracesEndpoint,
	}
}

func (s *SensorBucketAPI) ListIngresses() ([]ingressarchiver.ArchivedIngressDTO, error) {
	res, err := http.Get(s.ingressEndpoint)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		fmt.Printf("could not fetch ingresses: %d, %s\n", res.StatusCode, string(body))
		return nil, fmt.Errorf("could not fetch ingresses: %d", res.StatusCode)
	}
	var resBody web.APIResponse[[]ingressarchiver.ArchivedIngressDTO]
	if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
		fmt.Printf("could not decode ingresses response: %v", err)
		return nil, err
	}

	return resBody.Data, nil
}

func (s *SensorBucketAPI) ListPipelines(ids []uuid.UUID) ([]processing.Pipeline, error) {
	q := url.Values{}
	q["id"] = lo.Map(ids, func(id uuid.UUID, _ int) string { return id.String() })
	url := s.pipelinesEndpoint + "?" + q.Encode()
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		fmt.Printf("could not fetch pipelines: %d, %s\n", res.StatusCode, string(body))
		return nil, fmt.Errorf("could not fetch pipelines: %d", res.StatusCode)
	}
	var resBody web.APIResponse[[]processing.Pipeline]
	if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
		fmt.Printf("could not decode pipelines response: %v", err)
		return nil, err
	}

	return resBody.Data, nil
}

func (s *SensorBucketAPI) ListTraces(ids []uuid.UUID) ([]routes.TraceDTO, error) {
	q := url.Values{}
	q["trace_id"] = lo.Map(ids, func(id uuid.UUID, _ int) string { return id.String() })
	url := s.tracesEndpoint + "?" + q.Encode()
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		fmt.Printf("could not fetch traceDTO: %d, %s\n", res.StatusCode, string(body))
		return nil, fmt.Errorf("could not fetch traceDTO: %d", res.StatusCode)
	}
	var resBody web.APIResponse[[]routes.TraceDTO]
	if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
		fmt.Printf("could not decode traceDTO response: %v", err)
		return nil, err
	}

	return resBody.Data, nil
}
