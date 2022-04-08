package assetmanager

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
)

func (svc *Service) setupRoutes() {
	// TODO: Implement http routes
	r := svc.router

	r.Post("/assets", svc.httpCreateAsset())
	r.Get("/assets/{assetURN}", svc.httpGetAsset())
}

func (svc *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	svc.router.ServeHTTP(w, r)
}

func (svc *Service) httpCreateAsset() http.HandlerFunc {
	type request struct {
		Content json.RawMessage `json:"content,omitempty"`
		Type    string          `json:"type,omitempty"`
	}
	type response struct {
		URN     string          `json:"urn,omitempty"`
		Content json.RawMessage `json:"content,omitempty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		opts := CreateAssetOpts{
			Type:    req.Type,
			Content: req.Content,
		}
		asset, err := svc.CreateAsset(opts)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		sendJSON(w, response{
			URN:     asset.URN().String(),
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
		assetURN, err := url.PathUnescape(chi.URLParam(r, "assetURN"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		asset, err := svc.GetAsset(assetURN)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sendJSON(w, response{
			URN:     asset.URN().String(),
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
