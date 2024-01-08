package auth

import (
	"net/http"

	"sensorbucket.nl/sensorbucket/internal/web"
)

var (
	// Authorization errors
	ErrUnauthorized          = web.NewError(http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
	ErrNoTenantIdFound       = web.NewError(http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
	ErrNoPermissions         = web.NewError(http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
	ErrPermissionsNotGranted = web.NewError(http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
	ErrNoUserId              = web.NewError(http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")

	// Request and server errors
	ErrNoPermissionsToCheck    = web.NewError(http.StatusInternalServerError, "No permissions to check", "PERMISSIONS_NOT_CONFIGURED")
	ErrAuthHeaderMissing       = web.NewError(http.StatusBadRequest, "Authorization header must be set", "AUTH_HEADER_MISSING")
	ErrAuthHeaderInvalidFormat = web.NewError(http.StatusBadRequest, "Authorization header must be formatted as 'Bearer {token}'", "AUTH_HEADER_INVALID_FORMAT")
)
