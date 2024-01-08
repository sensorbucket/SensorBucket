package auth

import (
	"net/http"

	"sensorbucket.nl/sensorbucket/internal/web"
)

var (
	// Authorization errors
	ErrUnauthorized          = web.NewError(http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
	ErrNoTenantIdFound       = web.NewError(http.StatusUnauthorized, "Not attached to any tenant", "NOT_ATTACHED_TO_TENANT")
	ErrNoPermissions         = web.NewError(http.StatusUnauthorized, "Permissions not granted", "PERMISSIONS_NOT_GRANTED")
	ErrPermissionsNotGranted = web.NewError(http.StatusUnauthorized, "Permission not granted", "PERMISSION_NOT_GRANTED")
	ErrNoUserId              = web.NewError(http.StatusUnauthorized, "User ID not valid", "USER_ID_NOT_VALID")

	// Request and server errors
	ErrNoPermissionsToCheck    = web.NewError(http.StatusInternalServerError, "No permissions to check", "PERMISSIONS_NOT_CONFIGURED")
	ErrAuthHeaderMissing       = web.NewError(http.StatusBadRequest, "Authorization header must be set", "AUTH_HEADER_MISSING")
	ErrAuthHeaderInvalidFormat = web.NewError(http.StatusBadRequest, "Authorization header must be formatted as 'Bearer {token}'", "AUTH_HEADER_INVALID_FORMAT")
)
