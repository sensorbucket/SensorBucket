package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"sensorbucket.nl/sensorbucket/internal/web"
)

func MustHavePermissions(c context.Context, permissions ...string) error {
	if len(permissions) == 0 {
		return ErrNoPermissionsToCheck
	}
	_, err := tenantIdFromRequestContext(c)
	if err != nil {
		return err
	}
	permissionsFromContext, err := permissionsFromRequestContext(c)
	if err != nil {
		return err
	}
	for _, p := range permissions {
		found := false
		for _, fromContext := range permissionsFromContext {
			if p == fromContext {
				found = true
			}
		}
		if !found {
			return ErrPermissionsNotGranted
		}
	}
	return nil
}

func Protect(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				web.HTTPError(w, ErrAuthHeaderMissing)
				return
			}
			if !strings.Contains(auth, "Bearer ") {
				web.HTTPError(w, ErrAuthHeaderInvalidFormat)
				return
			}
			tokenStr := strings.TrimPrefix(auth, "Bearer ")

			// Retrieve the JWT and ensure it was signed by us
			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				return secret, nil
			})
			if err == nil && token.Valid && token.Claims.Valid() == nil {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					expired := !claims.VerifyExpiresAt(time.Now().Unix(), true)
					if !expired {
						tenantId, err := currentTenantFromClaims(claims)
						if err != nil {
							web.HTTPError(w, err)
							return
						}
						permissions, err := permissionsFromClaims(claims)
						if err != nil {
							web.HTTPError(w, err)
							return
						}

						// JWT itself is validated, pass it to the actual endpoint for further authorization
						// First fill the context with user information
						next.ServeHTTP(w, r.WithContext(
							context.
								WithValue(
									context.WithValue(r.Context(),
										"current_tenant_id", tenantId),
									"permissions",
									permissions,
								),
						))
						return
					}

				}
			}
			web.HTTPError(w, ErrUnauthorized)
		})
	}
}

func currentTenantFromClaims(claims jwt.MapClaims) (int64, error) {
	tenant, ok := claims["current_tenant_id"]
	if ok {
		// The JWT library converts the value to a float64 before it does to an int64
		asFl, ok := tenant.(float64)
		if ok {
			return int64(asFl), nil
		}
	}
	return -1, ErrNoTenantIdFound
}

func permissionsFromClaims(claims jwt.MapClaims) ([]string, error) {
	permissions, ok := claims["permissions"]
	if ok {

		// Permissions are given as a slice
		if asSlice, ok := permissions.([]interface{}); ok {

			// Each permission in the slice is of type interface but should be able to be converted to a string
			res := []string{}
			for _, perm := range asSlice {
				if val, ok := perm.(string); ok {
					res = append(res, val)
				} else {
					return nil, fmt.Errorf("permission could not be converted to string")
				}
			}
			return res, nil
		}
	}
	return nil, ErrNoPermissions
}

func permissionsFromRequestContext(c context.Context) ([]string, error) {
	if permissions, ok := c.Value("permissions").([]string); ok {
		return permissions, nil
	}
	return nil, ErrNoPermissions
}

func tenantIdFromRequestContext(c context.Context) (int64, error) {
	if tenantId, ok := c.Value("current_tenant_id").(int64); ok {
		return tenantId, nil
	}
	return -1, ErrNoTenantIdFound
}
