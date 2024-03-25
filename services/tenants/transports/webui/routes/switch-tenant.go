package routes

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ory/nosurf"
	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/internal/flash_messages"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/services/tenants/sessions"
	"sensorbucket.nl/sensorbucket/services/tenants/tenants"
	"sensorbucket.nl/sensorbucket/services/tenants/transports/webui/views"
)

type TenantSwitchingPageHandler struct {
	router          chi.Router
	tenantService   *tenants.TenantService
	userPreferences *sessions.UserPreferenceService
}

func SetupTenantSwitchingRoutes(tenantService *tenants.TenantService, userPreferences *sessions.UserPreferenceService) *TenantSwitchingPageHandler {
	handler := &TenantSwitchingPageHandler{
		router:          chi.NewRouter(),
		tenantService:   tenantService,
		userPreferences: userPreferences,
	}

	handler.router.With(flash_messages.ExtractFlashMessage).Get("/switch", handler.httpSwitchTenantPage())
	handler.router.Post("/switch", handler.httpDoSwitchTenantPage())

	return handler
}

func (handler TenantSwitchingPageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.router.ServeHTTP(w, r)
}

func (handler *TenantSwitchingPageHandler) httpSwitchTenantPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := &views.TenantSwitchingPage{
			Base: views.Base{
				CSRFToken: nosurf.Token(r),
			},
			SuccessURL: r.URL.Query().Get("success_url"),
		}
		defer func() {
			flash_messages.AddContextFlashMessages(r, &p.FlashMessagesContainer)
			views.WriteLayout(w, p)
		}()

		tenantsPage, err := handler.tenantService.ListTenants(r.Context(), tenants.Filter{
			IsMember: true,
		}, pagination.Request{})
		if err != nil {
			fmt.Printf("in httpSwitchTenantPage(), Error listing tenants: %v\n", err)
			flash_messages.AddErrorFlashMessageToPage(r, &p.FlashMessagesContainer, "An error occured listing tenants, please try again")
			return
		}
		tenantViews := lo.Map(tenantsPage.Data, func(item tenants.CreateTenantDTO, _ int) views.TenantView {
			logo := ""
			if item.Logo != nil {
				logo = *item.Logo
			}
			return views.TenantView{
				ID:       item.ID,
				Name:     item.Name,
				ImageURL: logo,
			}
		})
		p.Tenants = tenantViews
	}
}

func (handler *TenantSwitchingPageHandler) httpDoSwitchTenantPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			flash_messages.AddErrorFlashMessage(w, r, "An error occured parsing the form, please try again")
			http.Redirect(w, r, views.U("/switch"), http.StatusSeeOther)
			return
		}
		tenantIDString := r.FormValue("tenantID")
		tenantID, err := strconv.ParseInt(tenantIDString, 10, 64)
		if err != nil {
			flash_messages.AddErrorFlashMessage(w, r, "An error occured parsing the form, please try again")
			http.Redirect(w, r, views.U("/switch"), http.StatusSeeOther)
			return
		}
		if err := handler.userPreferences.SetActiveTenantID(r.Context(), tenantID); err != nil {
			flash_messages.AddErrorFlashMessage(w, r, "An error occured setting the new active organization, please try again")
			http.Redirect(w, r, views.U("/switch"), http.StatusSeeOther)
			return
		}
		flash_messages.AddSuccessFlashMessage(w, r, fmt.Sprintf("Succesfully switched to: %s", r.FormValue("tenantName")))
		successRedirect(w, r, "/tenants/auth/settings")
	}
}

func successRedirect(w http.ResponseWriter, r *http.Request, fallback string) {
	url := fallback
	if successURL := r.FormValue("successURL"); successURL != "" {
		url = successURL
	}
	http.Redirect(w, r, url, http.StatusSeeOther)
}
