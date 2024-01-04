package routes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	ory "github.com/ory/client-go"

	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/tenants/transports/webui/views"
)

type KratosRoutes struct {
	ory    *ory.APIClient
	router chi.Router
}

func SetupKratosRoutes() *KratosRoutes {
	k := &KratosRoutes{
		router: chi.NewRouter(),
	}

	oryConfig := ory.NewConfiguration()
	oryConfig.Servers = ory.ServerConfigurations{
		{
			URL: "http://kratos:4433/",
		},
	}
	k.ory = ory.NewAPIClient(oryConfig)

	k.router.Get("/", k.httpDefaultPage())
	k.router.With(requireFlow("login")).Get("/login", k.httpLoginPage())
	k.router.With(requireFlow("recovery")).Get("/recovery", k.httpRecoveryPage())
	k.router.With(requireFlow("settings")).Get("/settings", k.httpSettingsPage())
	k.router.With(requireFlow("error")).Get("/error", k.httpErrorPage())

	return k
}

func (k KratosRoutes) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	k.router.ServeHTTP(w, r)
}

func (k KratosRoutes) httpDefaultPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
			Data: r.Header.Get("Authorization"),
		})
	}
}

var ctxFlow = struct{}{}

func requireFlow(flow string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		mw := func(w http.ResponseWriter, r *http.Request) {
			if !r.URL.Query().Has("flow") {
				http.Redirect(w, r, fmt.Sprintf("http://127.0.0.1:3000/.ory/self-service/%s/browser", flow), http.StatusFound)
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), ctxFlow, r.URL.Query().Get("flow")))
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(mw)
	}
}

func flowID(r *http.Request) string {
	return r.Context().Value(ctxFlow).(string)
}

func (k KratosRoutes) httpLoginPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flow, res, err := k.ory.FrontendAPI.
			GetLoginFlow(r.Context()).Id(flowID(r)).
			Cookie(r.Header.Get("Cookie")).Execute()
		if err != nil {
			if res.StatusCode == http.StatusForbidden {
				http.Redirect(w, r, "http://127.0.0.1:3000/.ory/self-service/login/browser", http.StatusFound)
				return
			}
			web.HTTPError(w, err)
			return
		}
		views.WriteLayout(w, views.LoginPage{Flow: flow})
	}
}

func (k KratosRoutes) httpRecoveryPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flow, _, err := k.ory.FrontendAPI.GetRecoveryFlow(r.Context()).Id(flowID(r)).
			Cookie(r.Header.Get("Cookie")).Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		views.WriteLayout(w, views.RecoveryPage{Flow: flow})
	}
}

func (k KratosRoutes) httpSettingsPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flow, _, err := k.ory.FrontendAPI.GetSettingsFlow(r.Context()).Id(flowID(r)).
			Cookie(r.Header.Get("Cookie")).Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		views.WriteLayout(w, views.SettingsPage{Flow: flow})
	}
}

func (k KratosRoutes) httpErrorPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flow, _, err := k.ory.FrontendAPI.GetFlowError(r.Context()).Id(flowID(r)).Execute()
		if err != nil {
			web.HTTPError(w, err)
			return
		}
		views.WriteLayout(w, views.ErrorPage{Flow: flow})
	}
}
