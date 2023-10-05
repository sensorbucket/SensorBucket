package routes

import (
	"encoding/base64"
	"fmt"
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
	r.Get("/create", h.createWorkerPage())
	r.Post("/create", h.createWorker())
	r.Get("/{id}", h.workerDetails())
	r.Patch("/{id}", h.updateWorker())
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
		workerID := chi.URLParam(r, "id")
		res, _, err := h.client.WorkersApi.GetWorker(r.Context(), workerID).Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		resUC, _, err := h.client.WorkersApi.GetWorkerUserCode(r.Context(), workerID).Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		page.Worker = res.Data
		page.UserCode = resUC.GetData()

		if isHX(r) {
			page.WriteBody(w)
			return
		}
		views.WriteIndex(w, page)
	}
}

func (h *WorkerPageHandler) updateWorker() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workerID := chi.URLParam(r, "id")

		var dto api.UpdateWorkerRequest
		if err := r.ParseForm(); err != nil {
			errBadRequest := web.NewError(http.StatusBadRequest, "http request body is malformed", "ERR_BAD_REQUEST")
			web.HTTPError(w, fmt.Errorf("%w: %w", errBadRequest, err))
			return
		}
		if name := r.FormValue("name"); name != "" {
			dto.Name = &name
		}
		if desc := r.FormValue("description"); desc != "" {
			dto.Description = &desc
		}
		switch r.FormValue("state") {
		case "on":
			dto.SetState("enabled")
		default:
			dto.SetState("disabled")
		}
		if userCode := r.FormValue("userCode"); userCode != "" {
			dto.UserCode = &userCode
		}

		_, _, err := h.client.WorkersApi.UpdateWorker(r.Context(), workerID).UpdateWorkerRequest(dto).Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		w.Header().Set("HX-Redirect", "/workers")
		w.WriteHeader(http.StatusOK)
	}
}

func (h *WorkerPageHandler) createWorkerPage() http.HandlerFunc {
	const defaultUserCode = `
def process(payload, msg):
    return payload
    `
	ucb64 := base64.StdEncoding.EncodeToString([]byte(defaultUserCode))
	return func(w http.ResponseWriter, r *http.Request) {
		page := &views.WorkerEditorPage{
			UserCode: ucb64,
		}
		if isHX(r) {
			page.WriteBody(w)
			return
		}
		views.WriteIndex(w, page)
	}
}

func (h *WorkerPageHandler) createWorker() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			web.HTTPError(w, web.NewError(http.StatusBadRequest, "Bad request", ""))
			return
		}
		var dto api.CreateUserWorkerRequest
		dto.SetName(r.FormValue("name"))
		dto.SetUserCode(r.FormValue("userCode"))
		dto.SetDescription(r.FormValue("description"))

		_, _, err := h.client.WorkersApi.CreateWorker(r.Context()).CreateUserWorkerRequest(dto).Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		w.Header().Set("HX-Redirect", "/workers")
		w.WriteHeader(http.StatusOK)
	}
}
