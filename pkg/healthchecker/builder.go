package healthchecker

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"sensorbucket.nl/sensorbucket/internal/web"
)

type (
	Check func() (string, bool)
)

type (
	Checks      map[string]Check
	CheckResult struct {
		Ok    []string `json:"ok"`
		NotOk []string `json:"not_ok"`
	}
)

func (checks Checks) Perform() CheckResult {
	if len(checks) == 0 {
		return CheckResult{}
	}

	successes := []string{}
	fails := []string{}
	for name, check := range checks {
		if msg, ok := check(); ok {
			successes = append(successes, name+": "+msg)
		} else {
			fails = append(fails, name+": not "+msg)
		}
	}

	return CheckResult{
		Ok:    successes,
		NotOk: fails,
	}
}

type Builder struct {
	Address   string
	liveness  Checks
	readiness Checks

	errors []error
}

func Create() *Builder {
	return &Builder{
		liveness:  Checks{},
		readiness: Checks{},
		errors:    []error{},
	}
}

func (b *Builder) error(err error) *Builder {
	b.errors = append(b.errors, err)
	return b
}

func (b *Builder) WithAddress(addr string) *Builder {
	b.Address = addr
	return b
}

func (b *Builder) WithEnv() *Builder {
	addr, ok := os.LookupEnv("HEALTH_ADDR")
	if !ok {
		return b.error(errors.New("HEALTH_ADDR environment variable not set"))
	}
	return b.WithAddress(addr)
}

func (b *Builder) AddLiveness(name string, check Check) *Builder {
	b.liveness[name] = check
	return b
}

func (b *Builder) AddReadiness(name string, check Check) *Builder {
	b.readiness[name] = check
	return b
}

func (b *Builder) BuildHandler() (http.Handler, error) {
	if len(b.errors) > 0 {
		err := errors.Join(b.errors...)
		log.Printf("[HealthChecks] failed with errors: \n%s", err)
		return nil, err
	}

	r := chi.NewRouter()
	r.Get("/liveness", func(w http.ResponseWriter, r *http.Request) {
		status := http.StatusOK
		result := b.liveness.Perform()
		if len(result.NotOk) > 0 {
			status = http.StatusServiceUnavailable
		}
		web.HTTPResponse(w, status, result)
	})
	r.Get("/readiness", func(w http.ResponseWriter, r *http.Request) {
		status := http.StatusOK
		result := b.readiness.Perform()
		if len(result.NotOk) > 0 {
			status = http.StatusServiceUnavailable
		}
		web.HTTPResponse(w, status, result)
	})

	return r, nil
}

func (b *Builder) Start(ctx context.Context) func(ctx context.Context) error {
	r, err := b.BuildHandler()
	if err != nil {
		log.Printf("[HealthServer] could not start due to error(s): %s\n", err)
	}
	srv := &http.Server{
		Addr:         b.Address,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
		Handler:      r,
	}
	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) && err != nil {
			log.Printf("Health Server error: %v\n", err)
		}
	}()
	log.Printf("Health Server available at: %s\n", srv.Addr)

	return func(ctx context.Context) error {
		return srv.Shutdown(ctx)
	}
}
