package helloworld

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/cors"
)

func Serve(addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/hello-world", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	srv := &http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Handler:      cors.AllowAll().Handler(mux),
	}

	fmt.Println("Listening on ", addr, "...")
	defer fmt.Println("Shutting down")
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
