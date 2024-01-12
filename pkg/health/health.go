package health

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"sensorbucket.nl/sensorbucket/internal/web"
)

type Check func() bool

type checks map[string]Check

func (checks checks) Perform() checksResult {
	if len(checks) == 0 {
		return checksResult{Success: false}
	}

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

type HealthChecker struct {
	livelinessChecks checks
	readinessChecks  checks
	router           chi.Router
}

func NewHealthEndpoint() *HealthChecker {
	hc := HealthChecker{
		router:           chi.NewRouter(),
		livelinessChecks: checks{},
		readinessChecks:  checks{},
	}
	hc.setupRoutes(hc.router)
	return &hc
}

func (hc *HealthChecker) WithLiveChecks(checks checks) *HealthChecker {
	for name, c := range checks {
		hc.livelinessChecks[name] = c
	}
	return hc
}

func (hc *HealthChecker) WithReadyChecks(checks checks) *HealthChecker {
	for name, c := range checks {
		hc.readinessChecks[name] = c
	}
	return hc
}

func (hc *HealthChecker) setupRoutes(r chi.Router) {
	r.Get("/livez", hc.httpLivelinessCheck)
	r.Get("/readyz", hc.httpReadinessCheck)
}

func (hc HealthChecker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hc.router.ServeHTTP(w, r)
}

func (hc *HealthChecker) httpReadinessCheck(w http.ResponseWriter, r *http.Request) {
	checksResponse(w, hc.readinessChecks)
}

func (hc *HealthChecker) httpLivelinessCheck(w http.ResponseWriter, r *http.Request) {
	checksResponse(w, hc.livelinessChecks)
}

func (hc *HealthChecker) RunAsServer(address string) func(context.Context) error {
	srv := &http.Server{
		Addr:         address,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
		Handler:      hc,
	}
	go func() {
		log.Printf("HealthChecker endpoint available at: %s\n", srv.Addr)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) && err != nil {
			log.Printf("HealthChecker server closed unexpectedly: %s\n", err.Error())
		}
	}()
	return func(ctx context.Context) error {
		return srv.Shutdown(ctx)
	}
}

func checksResponse(w http.ResponseWriter, checks checks) {
	results := checks.Perform()
	statusCode := http.StatusOK
	if !results.Success {
		statusCode = http.StatusServiceUnavailable
	}
	web.HTTPResponse(w, statusCode, web.APIResponse[checksResult]{
		Message: fmt.Sprintf("%d/%d checks passed", len(results.ChecksSucess), len(checks)),
		Data:    results,
	})
}

type checksResult struct {
	Success      bool     `json:"success"`
	ChecksSucess []string `json:"checks_success"`
	ChecksFailed []string `json:"checks_failed"`
}
