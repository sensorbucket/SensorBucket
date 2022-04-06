package assetmanager

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (svc *Service) setupRoutes() {
	// TODO: Implement http routes
	r := svc.router

	r.Post("/assets/{assetType}", svc.httpCreateAsset())

	// TODO: Do we use the URN here? Or do we just use the assetID?
	// This choice affects the service API as well.
	r.Get("/assets/{assetType}/{assetID}", svc.httpGetAsset())
}

func (svc *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	svc.router.ServeHTTP(w, r)
}

func (svc *Service) httpCreateAsset() http.HandlerFunc {
	type response struct {
		URN     string          `json:"urn,omitempty"`
		Content json.RawMessage `json:"content,omitempty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		asset := &Asset{
			Type: chi.URLParam(r, "assetType"),
		}

		if err := json.NewDecoder(r.Body).Decode(&asset.Content); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := svc.CreateAsset(asset); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		sendJSON(w, response{
			URN:     asset.URN(),
			Content: asset.Content,
		})
	}
}

func (svc *Service) httpGetAsset() http.HandlerFunc {
	type response struct {
		URN     string          `json:"urn,omitempty"`
		Content json.RawMessage `json:"content,omitempty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		assetType := chi.URLParam(r, "assetType")
		assetID := chi.URLParam(r, "assetID")

		asset, err := svc.GetAsset(assetType, assetID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Printf("asset: %v\n", asset)

		sendJSON(w, response{
			URN:     asset.URN(),
			Content: asset.Content,
		})
	}
}

func sendJSON(w http.ResponseWriter, v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
