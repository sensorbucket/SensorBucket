package routes

import (
	"net/http"

	"github.com/ory/nosurf"

	"sensorbucket.nl/sensorbucket/internal/flash_messages"
	"sensorbucket.nl/sensorbucket/services/dashboard/views"
)

func createBasePage(r *http.Request) views.BasePage {
	base := views.BasePage{
		CSRFToken: nosurf.Token(r),
	}
	flash_messages.AddContextFlashMessages(r, &base.FlashMessagesContainer)
	return base
}
