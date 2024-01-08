package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"sensorbucket.nl/sensorbucket/internal/web"
)

var (
	UserIdKey          = "user_id"
	CurrentTenantIdKey = "current_tenant_id"
	PermissionsKey     = "permissions"
)

// Authentication middleware for checking the validity of a JWT
// Checks if the JWT is signed using the given secret
// Serves the next HTTP handler if all is OK
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
						tenantId, ok := currentTenantFromClaims(claims)
						if !ok {
							web.HTTPError(w, ErrNoTenantIdFound)
							return
						}
						permissions, ok := permissionsFromClaims(claims)
						if !ok {
							web.HTTPError(w, ErrNoPermissions)
							return
						}
						userId, ok := userFromClaims(claims)
						if !ok {
							web.HTTPError(w, ErrNoUserId)
							return
						}

						// JWT itself is validated, pass it to the actual endpoint for further authorization
						// First fill the context with user information
						next.ServeHTTP(w, r.WithContext(contextWithValues(r.Context(), map[string]interface{}{
							CurrentTenantIdKey: tenantId,
							UserIdKey:          userId,
							PermissionsKey:     permissions,
						})))
						return
					}

				}
			}
			web.HTTPError(w, ErrUnauthorized)
		})
	}
}

func contextWithValues(ctx context.Context, values map[string]interface{}) context.Context {
	for key, val := range values {
		ctx = context.WithValue(ctx, key, val)
	}
	return ctx
}

func currentTenantFromClaims(claims jwt.MapClaims) (int64, bool) {
	return int64FromClaims(claims, CurrentTenantIdKey)
}

func userFromClaims(claims jwt.MapClaims) (int64, bool) {
	return int64FromClaims(claims, UserIdKey)
}

func int64FromClaims(claims jwt.MapClaims, key string) (int64, bool) {
	val, ok := claims[key]
	if ok {
		// The JWT library converts the value to a float64 before it does to an int64
		asFl, ok := val.(float64)
		if ok {
			return int64(asFl), true
		}
	}
	return -1, false
}

func permissionsFromClaims(claims jwt.MapClaims) ([]permission, bool) {
	permissions, ok := claims[PermissionsKey]
	if ok {

		// Permissions are given as a slice
		if asSlice, ok := permissions.([]interface{}); ok {

			// Each permission in the slice is of type interface but should be able to be converted to a string
			res := []permission{}
			for _, perm := range asSlice {
				if val, ok := perm.(string); ok {
					res = append(res, permission(val))
				} else {
					return nil, false
				}
			}
			return res, true
		}
	}
	return nil, false
}
