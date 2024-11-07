package web

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"sensorbucket.nl/sensorbucket/internal/env"
)

func RunProfiler() (func(context.Context), error) {
	addr := env.Could("PROFILER_ADDR", "")
	if addr == "" {
		return func(ctx context.Context) {}, nil
	}
	srv := &http.Server{
		Addr:         addr,
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
		Handler:      profiler(),
	}
	go func() {
		log.Printf("[Info] Running Profiler on: %s\n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("[Warn] Profiler server crashed: %s\n", err.Error())
		}
	}()

	return func(shutdownCtx context.Context) {
		if err := srv.Shutdown(shutdownCtx); !errors.Is(err, http.ErrServerClosed) && err != nil {
			log.Printf("Profiler HTTP Server error during shutdown: %v\n", err)
		}
	}, nil
}

func profiler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.NoCache)

	r.HandleFunc("/*", pprof.Index)
	r.HandleFunc("/cmdline", pprof.Cmdline)
	r.HandleFunc("/profile", pprof.Profile)
	r.HandleFunc("/symbol", pprof.Symbol)
	r.HandleFunc("/trace", pprof.Trace)
	r.HandleFunc("/vars", expVars)

	r.Handle("/goroutine", pprof.Handler("goroutine"))
	r.Handle("/threadcreate", pprof.Handler("threadcreate"))
	r.Handle("/mutex", pprof.Handler("mutex"))
	r.Handle("/heap", pprof.Handler("heap"))
	r.Handle("/block", pprof.Handler("block"))
	r.Handle("/allocs", pprof.Handler("allocs"))

	return r
}

// Replicated from expvar.go as not public.
func expVars(w http.ResponseWriter, r *http.Request) {
	first := true
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\n")
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintf(w, ",\n")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "\n}\n")
}
