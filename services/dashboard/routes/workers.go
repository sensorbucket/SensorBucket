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
	router        chi.Router
	workersClient *api.APIClient
}

func CreateWorkerPageHandler(workers *api.APIClient) *WorkerPageHandler {
	handler := &WorkerPageHandler{
		router:        chi.NewRouter(),
		workersClient: workers,
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
	r.Get("/table", h.workersTable())
	r.Patch("/{id}", h.updateWorker())
}

func (h *WorkerPageHandler) listWorkers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := &views.WorkerListPage{
			BasePage: createBasePage(r),
		}
		res, _, err := h.workersClient.WorkersApi.ListWorkers(r.Context()).Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		page.Workers = res.Data
		page.WorkersNextPage = res.Links.GetNext()

		if res.Links.GetNext() != "" {
			page.WorkersNextPage = views.U("/workers/table?cursor=" + getCursor(res.Links.GetNext()))
		}

		fmt.Println("Cursor", res.Links.GetNext())
		if isHX(r) {
			page.WriteBody(w)
			return
		}
		views.WriteIndex(w, page)
	}
}

func (h *WorkerPageHandler) workersTable() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("get table")
		req := h.workersClient.WorkersApi.ListWorkers(r.Context())
		if r.URL.Query().Has("cursor") {
			req = req.Cursor(r.URL.Query().Get("cursor"))
		}
		res, _, err := req.Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		nextCursor := ""
		if res.Links.GetNext() != "" {
			nextCursor = views.U("/workers/table?cursor=" + getCursor(res.Links.GetNext()))
		}

		if isHX(r) {
			views.WriteRenderWorkerTableRows(w, res.Data, nextCursor)
		}
	}
}

func (h *WorkerPageHandler) workerDetails() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := &views.WorkerEditorPage{
			BasePage: createBasePage(r),
		}
		workerID := chi.URLParam(r, "id")
		res, _, err := h.workersClient.WorkersApi.GetWorker(r.Context(), workerID).Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		resUC, _, err := h.workersClient.WorkersApi.GetWorkerUserCode(r.Context(), workerID).Execute()
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
			web.HTTPError(w, fmt.Errorf("%w: %s", errBadRequest, err.Error()))
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

		_, _, err := h.workersClient.WorkersApi.UpdateWorker(r.Context(), workerID).UpdateWorkerRequest(dto).Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		w.Header().Set("HX-Redirect", views.U("/workers"))
		w.WriteHeader(http.StatusOK)
	}
}

func (h *WorkerPageHandler) createWorkerPage() http.HandlerFunc {
	const defaultUserCode = `
def process(msg):
    return msg
    `
	ucb64 := base64.StdEncoding.EncodeToString([]byte(defaultUserCode))
	return func(w http.ResponseWriter, r *http.Request) {
		page := &views.WorkerEditorPage{
			BasePage: createBasePage(r),
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

		_, _, err := h.workersClient.WorkersApi.CreateWorker(r.Context()).CreateUserWorkerRequest(dto).Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		w.Header().Set("HX-Redirect", views.U("/workers"))
		w.WriteHeader(http.StatusOK)
	}
}
