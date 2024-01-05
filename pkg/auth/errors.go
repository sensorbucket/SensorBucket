package main

import (
	"net/http"

	"sensorbucket.nl/sensorbucket/internal/web"
)

var (
	ErrUnauthorized = web.NewError(http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")

	ErrAuthHeaderMissing       = web.NewError(http.StatusBadRequest, "Authorization header must be st", "AUTH_HEADER_MISSING")
	ErrAuthHeaderInvalidFormat = web.NewError(http.StatusBadRequest, "Authorization header must be formatted as 'Bearer {token}'", "AUTH_HEADER_INVALID_FORMAT")
)
