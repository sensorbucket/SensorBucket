package auth

import (
	"net/http"

	"sensorbucket.nl/sensorbucket/internal/web"
)

var (
	// Authorization errors
	ErrUnauthorized    = web.NewError(http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
	ErrForbidden       = web.NewError(http.StatusUnauthorized, "Forbidden", "FORBIDDEN")
	ErrNoTenantIDFound = web.NewError(
		http.StatusForbidden,
		"Forbidden no tenant",
		"FORBIDDEN",
	)
	ErrNoPermissions = web.NewError(
		http.StatusForbidden,
		"Forbidden no permissions",
		"FORBIDDEN",
	)
	ErrNoAccessToken = web.NewError(
		http.StatusForbidden,
		"Forbidden no access token",
		"FORBIDDEN",
	)
	ErrPermissionsNotGranted = web.NewError(
		http.StatusForbidden,
		"Forbidden permissions not granted",
		"FORBIDDEN",
	)
	ErrNoUserID = web.NewError(http.StatusForbidden, "Forbidden no user", "FORBIDDEN")

	// Request and server errors
	ErrAuthHeaderInvalidFormat = web.NewError(
		http.StatusBadRequest,
		"Authorization header must be formatted as 'Bearer {token}'",
		"AUTH_HEADER_INVALID_FORMAT",
	)
)
