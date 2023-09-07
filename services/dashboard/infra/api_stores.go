package dashboardinfra

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/api"
	"sensorbucket.nl/sensorbucket/services/core/devices"
	"sensorbucket.nl/sensorbucket/services/core/processing"
	"sensorbucket.nl/sensorbucket/services/dashboard/routes"
	ingressarchiver "sensorbucket.nl/sensorbucket/services/tracing/ingress-archiver/service"
)

type SensorBucketAPI struct {
	api               *api.APIClient
	ingressEndpoint   string
	pipelinesEndpoint string
	tracesEndpoint    string
	devicesEndpoint   string
}

func NewSensorBucketAPI(ingressEndpoint, pipelinesEndpoint, tracesEndpoint, devicesEndpoint string) *SensorBucketAPI {
	cfg := api.NewConfiguration()
	cfg.Scheme = "http"
	cfg.Host = "caddy"
	return &SensorBucketAPI{
		api:               api.NewAPIClient(cfg),
		ingressEndpoint:   ingressEndpoint,
		pipelinesEndpoint: pipelinesEndpoint,
		tracesEndpoint:    tracesEndpoint,
		devicesEndpoint:   devicesEndpoint,
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
	idStrings := lo.Map(ids, func(id uuid.UUID, _ int) string { return id.String() })
	res, _, err := s.api.TracingApi.ListTraces(context.Background()).TracingId(idStrings).Execute()
	if err != nil {
		return nil, err
	}
	traces := make([]routes.TraceDTO, 0, len(res.Data))
	for _, resTrace := range res.Data {
		steps := lo.Map(resTrace.Steps, func(step api.TraceStep, _ int) routes.StepDTO {
			return routes.StepDTO{
				Status:   int(step.Status),
				Duration: time.Duration(int64(step.Duration * 1e9)),
				Error:    step.GetError(),
			}
		})
		traces = append(traces, routes.TraceDTO{
			TracingId: resTrace.TracingId,
			DeviceID:  resTrace.DeviceId,
			Status:    int(resTrace.Status),
			Steps:     steps,
		})
	}
	return traces, nil
}

func (s *SensorBucketAPI) ListDevices(ids []int64) ([]devices.Device, error) {
	q := url.Values{}
	q["id"] = lo.Map(ids, func(id int64, _ int) string { return strconv.FormatInt(id, 10) })
	url := s.devicesEndpoint + "?" + q.Encode()
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		fmt.Printf("could not fetch traceDTO: %d, %s\n", res.StatusCode, string(body))
		return nil, fmt.Errorf("could not fetch devices: %d", res.StatusCode)
	}
	var resBody web.APIResponse[[]devices.Device]
	if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
		fmt.Printf("could not decode devices response: %v", err)
		return nil, err
	}

	return resBody.Data, nil
}
