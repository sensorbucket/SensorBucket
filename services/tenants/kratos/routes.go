package kratos

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	ory "github.com/ory/client-go"
)

type KratosRoutes struct {
	ory    *ory.APIClient
	router chi.Router
}

func SetupRoutes() *KratosRoutes {
	k := &KratosRoutes{
		router: chi.NewRouter(),
	}
	r := k.router
	r.Get("/auth", k.httpLoginPage())

	oryConfig := ory.NewConfiguration()
	oryConfig.Servers = ory.ServerConfigurations{
		{
			URL: "http://kratos:4433/.ory",
		},
	}
	k.ory = ory.NewAPIClient(oryConfig)

	return k
}

func (k KratosRoutes) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	k.router.ServeHTTP(w, r)
}

func (k KratosRoutes) httpLoginPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}
