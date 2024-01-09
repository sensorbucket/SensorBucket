package health

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"sensorbucket.nl/sensorbucket/internal/web"
)

type Check func() bool

type HealthChecker struct {
	livelinessChecks map[string]Check
	readinessChecks  map[string]Check
	router           *chi.Mux
	cancelToken      chan interface{}
}

func NewHealthEndpoint(cancelToken chan interface{}) *HealthChecker {
	r := chi.NewRouter()
	hc := HealthChecker{
		router:           r,
		livelinessChecks: map[string]Check{},
		readinessChecks:  map[string]Check{},
	}
	hc.setupRoutes(hc.router)
	return &hc
}

func (hc *HealthChecker) setupRoutes(r chi.Router) {
	r.Get("/livez", hc.livelinessCheck)
	r.Get("/readyz", hc.readinessCheck)
}

func (hc *HealthChecker) WithLiveChecks(checks map[string]Check) *HealthChecker {
	for name, c := range checks {
		hc.livelinessChecks[name] = c
	}
	return hc
}

func (hc *HealthChecker) WithReadyChecks(checks map[string]Check) *HealthChecker {
	for name, c := range checks {
		hc.readinessChecks[name] = c
	}
	return hc
}

func (hc *HealthChecker) Start(addr string) {
	server := &http.Server{
		Addr:    addr,
		Handler: hc.router,
	}
	log.Printf("[Info] Started Health endpoint on %s", addr)
	err := server.ListenAndServe()
	if err != nil {
		log.Printf("[Error] Health endpoint stopped running: %v, shutting down...\n", err)
		if hc.cancelToken != nil {
			hc.cancelToken <- true
		}
	}
}

func (hc *HealthChecker) readinessCheck(w http.ResponseWriter, r *http.Request) {
	if len(hc.readinessChecks) == 0 {
		web.HTTPResponse(w, http.StatusOK, web.APIResponse[checksResult]{
			Message: fmt.Sprintf("WARNING: no readiness checks configured"),
		})
		return
	}
	res := performChecks(hc.readinessChecks)
	if res.Success {
		web.HTTPResponse(w, http.StatusOK, web.APIResponse[checksResult]{
			Message: fmt.Sprintf("%d/%d readiness checks passed", len(res.ChecksSucess), len(hc.readinessChecks)),
			Data:    res,
		})
		return
	}
	web.HTTPResponse(w, http.StatusInternalServerError, web.APIResponse[checksResult]{
		Message: fmt.Sprintf("%d/%d readiness checks passed", len(res.ChecksSucess), len(hc.readinessChecks)),
		Data:    res,
	})
}

func (hc *HealthChecker) livelinessCheck(w http.ResponseWriter, r *http.Request) {
	if len(hc.livelinessChecks) == 0 {
		web.HTTPResponse(w, http.StatusOK, web.APIResponse[checksResult]{
			Message: fmt.Sprintf("WARNING: no liveliness checks configured"),
		})
		return
	}
	res := performChecks(hc.livelinessChecks)
	if res.Success {
		web.HTTPResponse(w, http.StatusOK, web.APIResponse[checksResult]{
			Message: fmt.Sprintf("%d/%d liveliness checks passed", len(res.ChecksSucess), len(hc.livelinessChecks)),
			Data:    res,
		})
		return
	}
	web.HTTPResponse(w, http.StatusInternalServerError, web.APIResponse[checksResult]{
		Message: fmt.Sprintf("%d/%d liveliness checks passed", len(hc.livelinessChecks)-len(res.ChecksFailed), len(hc.livelinessChecks)),
		Data:    res,
	})
}

func performChecks(checks map[string]Check) checksResult {
	success := []string{}
	failed := []string{}
	for name, check := range checks {
		if check() {
			success = append(success, name)
			continue
		}
		failed = append(failed, name)
	}
	return checksResult{
		Success:      len(failed) == 0,
		ChecksSucess: success,
		ChecksFailed: failed,
	}
}

type checksResult struct {
	Success      bool     `json:"success"`
	ChecksSucess []string `json:"checks_success"`
	ChecksFailed []string `json:"checks_failed"`
}
