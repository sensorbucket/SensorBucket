package auth

import (
	"net/http"

	"sensorbucket.nl/sensorbucket/internal/web"
)

var (
	// Authorization errors
	ErrUnauthorized          = web.NewError(http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
	ErrNoTenantIDFound       = web.NewError(http.StatusForbidden, "Forbidden", "FORBIDDEN")
	ErrNoPermissions         = web.NewError(http.StatusForbidden, "Forbidden", "FORBIDDEN")
	ErrPermissionsNotGranted = web.NewError(http.StatusForbidden, "Forbidden", "FORBIDDEN")
	ErrNoUserID              = web.NewError(http.StatusForbidden, "Forbidden", "FORBIDDEN")

	// Request and server errors
	ErrAuthHeaderInvalidFormat = web.NewError(http.StatusBadRequest, "Authorization header must be formatted as 'Bearer {token}'", "AUTH_HEADER_INVALID_FORMAT")
)
