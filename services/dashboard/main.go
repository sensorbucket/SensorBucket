package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"sensorbucket.nl/sensorbucket/internal/env"
)

var HTTP_ADDR = env.Could("HTTP_ADDR", ":3000")

func main() {
	if err := Run(); err != nil {
		panic(fmt.Sprintf("error: %v\n", err))
	}
}

func Run() error {
	errC := make(chan error, 1)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	// Serve static files only in development
	if os.Getenv("GO_ENV") != "production" {
		fmt.Println("Serving static files...")
		startTS := time.Now()
		staticPath := env.Must("STATIC_PATH")
		router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Last-Modified", startTS.Format("Monday, 02 January 2006 15:04:05 MST"))
				next.ServeHTTP(w, r)
			})
		})
		router.Use(middleware.GetHead)
		fileServer := http.FileServer(http.Dir(staticPath))
		router.Handle("/static/*", http.StripPrefix("/static", fileServer))
	}

	router.Get("/", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/overview", http.StatusFound) })
	router.Mount("/overview", createOverviewPageHandler())
	srv := &http.Server{
		Addr:         HTTP_ADDR,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
		Handler:      router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errC <- err
		}
	}()
	fmt.Printf("HTTP Server listening on: %s\n", srv.Addr)

	// Wait for fatal error or interrupt signal
	var err error
	select {
	case <-ctx.Done():
	case err = <-errC:
		cancel()
	}

	ctxTO, cancelTO := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelTO()

	if err := srv.Shutdown(ctxTO); err != nil {
		fmt.Printf("could not gracefully shutdown http server: %s\n", err)
	}

	return err
}