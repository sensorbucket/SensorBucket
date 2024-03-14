package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/services/tenants/tenants"
	"sensorbucket.nl/sensorbucket/services/tenants/transports/webui/views"
)

type TenantSwitchingPageHandler struct {
	router        chi.Router
	tenantService *tenants.TenantService
}

func SetupTenantSwitchingRoutes(tenantService *tenants.TenantService) *TenantSwitchingPageHandler {
	handler := &TenantSwitchingPageHandler{
		router:        chi.NewRouter(),
		tenantService: tenantService,
	}

	// TODO: user must be authenticated to update org
	handler.router.Get("/tenant", handler.httpSwitchTenantPage())

	return handler
}

func (handler TenantSwitchingPageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.router.ServeHTTP(w, r)
}

func (handler *TenantSwitchingPageHandler) httpSwitchTenantPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantsPage, err := handler.tenantService.ListTenants(r.Context(), tenants.Filter{
			IsMember: true,
		}, pagination.Request{})
		if err != nil {
			w.Write([]byte("error occured" + err.Error()))
			return
		}
		tenantViews := lo.Map(tenantsPage.Data, func(item tenants.CreateTenantDTO, index int) views.TenantView {
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
		p := &views.TenantSwitchingPage{
			Tenants: tenantViews,
		}
		views.WriteLayout(w, p)
	}
}
