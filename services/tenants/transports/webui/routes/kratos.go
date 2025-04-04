package routes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	ory "github.com/ory/client-go"

	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/internal/flash_messages"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/tenants/transports/webui/views"
)

type KratosFlow string

const (
	FlowLogin    KratosFlow = "login"
	FlowSettings KratosFlow = "settings"
	FlowRecovery KratosFlow = "recovery"
	FlowError    KratosFlow = "error"
	FlowLogout   KratosFlow = "logout"
)

type KratosRoutes struct {
	ory             *ory.APIClient
	router          chi.Router
	kratosPublicURI string
	kratosServerURI string
}

func SetupKratosRoutes() *KratosRoutes {
	k := &KratosRoutes{
		router:          chi.NewRouter(),
		kratosPublicURI: env.Could("KRATOS_PUBLIC_URI", "/.ory"),
		kratosServerURI: env.Could("KRATOS_SERVER_URI", "http://kratos:4433"),
	}

	oryConfig := ory.NewConfiguration()
	oryConfig.Servers = ory.ServerConfigurations{
		{
			URL: k.kratosServerURI,
		},
	}
	k.ory = ory.NewAPIClient(oryConfig)

	k.router.With(k.extractFlow(FlowLogin)).Get("/login", k.httpLoginPage())
	k.router.With(k.extractFlow(FlowRecovery)).Get("/recovery", k.httpRecoveryPage())
	k.router.With(
		k.extractFlow(FlowSettings),
		flash_messages.ExtractFlashMessage,
	).Route("/settings", func(r chi.Router) {
		r.Get("/", k.httpSettingsPage())
	})
	k.router.With(k.extractFlow(FlowError)).Get("/error", k.httpErrorPage())
	k.router.With(k.extractFlow(FlowLogout)).Get("/logout", k.httpLogoutPage())

	return k
}

func (k KratosRoutes) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	k.router.ServeHTTP(w, r)
}

type ctxKey string

var ctxFlow ctxKey = "flow"

func (k KratosRoutes) redirectStartFlow(w http.ResponseWriter, r *http.Request, flow KratosFlow) {
	http.Redirect(w, r, fmt.Sprintf("%s/self-service/%s/browser", k.kratosPublicURI, flow), http.StatusSeeOther)
}

func (k KratosRoutes) extractFlow(flow KratosFlow) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		mw := func(w http.ResponseWriter, r *http.Request) {
			if !r.URL.Query().Has("flow") && flow != FlowLogout && flow != FlowError {
				k.redirectStartFlow(w, r, flow)
				return
			}
			flowID := r.URL.Query().Get("flow")
			errorID := r.URL.Query().Get("id")
			cookie := r.Header.Get("Cookie")

			var flowData any
			var err error
			var resp *http.Response
			switch flow {
			case FlowLogin:
				flowData, resp, err = k.ory.FrontendAPI.GetLoginFlow(r.Context()).Id(flowID).Cookie(cookie).Execute()
			case FlowRecovery:
				flowData, resp, err = k.ory.FrontendAPI.GetRecoveryFlow(r.Context()).Id(flowID).Cookie(cookie).Execute()
			case FlowSettings:
				flowData, resp, err = k.ory.FrontendAPI.GetSettingsFlow(r.Context()).Id(flowID).Cookie(cookie).Execute()
			case FlowError:
				flowData, resp, err = k.ory.FrontendAPI.GetFlowError(r.Context()).Id(errorID).Execute()
			case FlowLogout:
				flowData, resp, err = k.ory.FrontendAPI.CreateBrowserLogoutFlow(r.Context()).Cookie(cookie).Execute()
			}
			if err != nil {
				if resp != nil && (resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusGone) {
					k.redirectStartFlow(w, r, flow)
					return
				}
				web.HTTPError(w, err)
				return
			}
			if flowData == nil {
				web.HTTPError(w, fmt.Errorf("expected FlowData to not be nil after succesful request to ory"))
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), ctxFlow, flowData))
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(mw)
	}
}

func (k KratosRoutes) httpLoginPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		views.WriteLayout(w, views.LoginPage{Flow: loginFlow(r)})
	}
}

func (k KratosRoutes) httpLogoutPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		views.WriteLayout(w, views.LogoutPage{URL: logoutFlow(r).GetLogoutUrl()})
	}
}

func (k KratosRoutes) httpRecoveryPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		views.WriteLayout(w, views.RecoveryPage{Flow: recoveryFlow(r)})
	}
}

func (k KratosRoutes) httpSettingsPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flow := settingsFlow(r)
		page := views.SettingsPage{Flow: flow, Base: views.Base{FlashMessagesContainer: flash_messages.FlashMessagesContainer{}}}
		flash_messages.AddContextFlashMessages(r, &page.FlashMessagesContainer)
		views.WriteWideLayout(w, page)
	}
}

func (k KratosRoutes) httpErrorPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		views.WriteLayout(w, views.ErrorPage{Flow: errorFlow(r)})
	}
}

func flowAs[T any](r *http.Request) *T {
	return r.Context().Value(ctxFlow).(*T)
}

func loginFlow(r *http.Request) *ory.LoginFlow {
	return flowAs[ory.LoginFlow](r)
}

func logoutFlow(r *http.Request) *ory.LogoutFlow {
	return flowAs[ory.LogoutFlow](r)
}

func recoveryFlow(r *http.Request) *ory.RecoveryFlow {
	return flowAs[ory.RecoveryFlow](r)
}

func settingsFlow(r *http.Request) *ory.SettingsFlow {
	return flowAs[ory.SettingsFlow](r)
}

func errorFlow(r *http.Request) *ory.FlowError {
	return flowAs[ory.FlowError](r)
}
