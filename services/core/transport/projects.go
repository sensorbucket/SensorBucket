package coretransport

import (
	"net/http"

	"sensorbucket.nl/sensorbucket/internal/httpfilter"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/core/projects"
)

type HTTPProjectsFilter struct {
	projects.ProjectsFilter
	pagination.Request
}

func (t *CoreTransport) httpListProjects() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filter, err := httpfilter.Parse[HTTPProjectsFilter](r)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		page, err := t.projectsService.ListProjects(r.Context(), filter.ProjectsFilter, filter.Request)
		if err != nil {
			web.HTTPError(w, err)
			return
		}

		web.HTTPResponse(w, http.StatusOK, pagination.CreateResponse(r, t.baseURL, *page))
	}
}
