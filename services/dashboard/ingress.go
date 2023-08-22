package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/core/processing"
	"sensorbucket.nl/sensorbucket/services/dashboard/views"
	ingressarchiver "sensorbucket.nl/sensorbucket/services/tracing/ingress-archiver/service"
)

func createIngressPageHandler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.GetHead)
	r.Get("/", ingressListPage())
	return r
}

func getIngresses() ([]ingressarchiver.ArchivedIngressDTO, error) {
	res, err := http.Get("http://tracing:3000/ingresses")
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

func getPipelines(ids []uuid.UUID) ([]processing.Pipeline, error) {
	q := url.Values{}
	//	q["id"] = lo.Map(ids, func(id uuid.UUID, _ int) string { return id.String() })
	url := "http://core:3000/pipelines?" + q.Encode()
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

func ingressListPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		archivedIngresses, err := getIngresses()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		pipelineIDs := lo.FilterMap(archivedIngresses, func(ingr ingressarchiver.ArchivedIngressDTO, _ int) (uuid.UUID, bool) {
			if ingr.IngressDTO == nil {
				return uuid.UUID{}, false
			}
			return ingr.IngressDTO.PipelineID, true
		})
		pipelineIDs = lo.Uniq(pipelineIDs)
		pipelines, err := getPipelines(pipelineIDs)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		ingresses := make([]views.Ingress, 0, len(archivedIngresses))
		for _, ingress := range archivedIngresses {
			if ingress.IngressDTO == nil {
				continue
			}
			pl, found := lo.Find(pipelines, func(pl processing.Pipeline) bool {
				return pl.ID == ingress.IngressDTO.PipelineID.String()
			})
			if !found {
				continue
			}
			ingresses = append(ingresses, views.Ingress{
				TracingID: ingress.TracingID.String(),
				CreatedAt: ingress.IngressDTO.CreatedAt,
				Steps: lo.Map(pl.Steps, func(step string, ix int) views.IngressStep {
					return views.IngressStep{
						Label:  step,
						Status: 2,
					}
				}),
			})
		}

		views.WriteIndex(w, &views.IngressPage{
			Ingresses: ingresses,
		})
	}
}
