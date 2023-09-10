package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/api"
	"sensorbucket.nl/sensorbucket/services/dashboard/views"
)

type WorkerPageHandler struct {
	router chi.Router
	client *api.APIClient
}

func CreateWorkerPageHandler(client *api.APIClient) *WorkerPageHandler {
	handler := &WorkerPageHandler{
		router: chi.NewRouter(),
		client: client,
	}
	handler.SetupRoutes(handler.router)
	return handler
}

func (h WorkerPageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *WorkerPageHandler) SetupRoutes(r chi.Router) {
	r.Get("/", h.listWorkers())
	r.Get("/{id}", h.workerDetails())
}

func (h *WorkerPageHandler) listWorkers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := &views.WorkerListPage{}
		res, _, err := h.client.WorkersApi.ListWorkers(r.Context()).Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		page.Workers = res.Data

		if isHX(r) {
			page.WriteBody(w)
			return
		}
		views.WriteIndex(w, page)
	}
}

func (h *WorkerPageHandler) workerDetails() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := &views.WorkerEditorPage{}
		res, _, err := h.client.WorkersApi.ListWorkers(r.Context()).Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		page.Worker = res.Data[0]

		if isHX(r) {
			page.WriteBody(w)
			return
		}
		views.WriteIndex(w, page)
	}
}
