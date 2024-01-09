package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/golang-jwt/jwt"
	"sensorbucket.nl/sensorbucket/internal/web"
)

var (
	UserIdKey          = "user_id"
	CurrentTenantIdKey = "current_tenant_id"
	PermissionsKey     = "permissions"
)

func Protect() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, tenantIdPresent := fromRequestContext[[]int64](r.Context(), CurrentTenantIdKey)
			_, permissionsPresent := fromRequestContext[[]permission](r.Context(), PermissionsKey)
			_, userIdPresent := fromRequestContext[int64](r.Context(), UserIdKey)
			if tenantIdPresent && permissionsPresent && userIdPresent {
				// All required authentication values are present, allow the request
				next.ServeHTTP(w, r)
				return
			}
			web.HTTPError(w, ErrUnauthorized)
		})
	}
}

// Authentication middleware for checking the validity of any present JWT
// Checks if the JWT is signed using the given secret
// Serves the next HTTP handler if there is no JWT or if the JWT is OK
// Anonymous requests are allowed by this handler
func Authenticate(issuer string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				// No authorization header present, return OK since there is no info to extract
				// anonymous requests are allowed
				return
			}
			tokenStr, ok := strings.CutPrefix(auth, "Bearer ")
			if !ok {
				web.HTTPError(w, ErrAuthHeaderInvalidFormat)
				return
			}

			// Retrieve the JWT and ensure it was signed by us
			c := claims{}
			token, err := jwt.ParseWithClaims(tokenStr, &c, validateJWTFunc(issuer))
			if err == nil && token.Valid && token.Claims.Valid() == nil {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					expired := !claims.VerifyExpiresAt(time.Now().Unix(), true)
					if !expired {

						// JWT itself is validated, pass it to the actual endpoint for further authorization
						// First fill the context with user information
						next.ServeHTTP(w, r.WithContext(contextWithValues(r.Context(), map[string]interface{}{
							CurrentTenantIdKey: []int64{c.TenantId},
							UserIdKey:          c.UserId,
							PermissionsKey:     c.Permissions,
						})))
						return
					}
				}
			}
			web.HTTPError(w, ErrUnauthorized)
		})
	}
}

func validateJWTFunc(issuer string) func(token *jwt.Token) (any, error) {
	return func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			fmt.Println("sign err", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Retrieve and decode JWKS
		res, err := http.Get(fmt.Sprintf("%s/.well-known/jwks.json", issuer))
		if err != nil {
			fmt.Print("fetch err", err)
			return nil, fmt.Errorf("failed to fetch jwks: %w", err)
		}
		fmt.Println("gound stuff", res.StatusCode)
		var jwks jose.JSONWebKeySet
		if err := json.NewDecoder(res.Body).Decode(&jwks); err != nil {
			return nil, fmt.Errorf("failed to decode jwks: %w", err)
		}

		// Look for the key as indicated by the token key id
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("no kid in token")
		}
		keys := jwks.Key(kid)
		if len(keys) == 0 {
			return nil, fmt.Errorf("no keys found for token")
		}
		key := keys[0]
		if key.Algorithm != token.Method.Alg() {
			return nil, fmt.Errorf("key alg differs from token alg: %s vs %s", key.Algorithm, token.Method.Alg())
		}
		return key.Public().Key, nil
	}
}

func contextWithValues(ctx context.Context, values map[string]interface{}) context.Context {
	for key, val := range values {
		ctx = context.WithValue(ctx, key, val)
	}
	return ctx
}

type claims struct {
	TenantId    int64    `json:"current_tenant_id"`
	Permissions []string `json:"permissions"`
	UserId      int64    `json:"user_id"`
}

func (c *claims) Valid() error {
	if c.TenantId > 0 && c.UserId > 0 {
		return nil
	}
	return fmt.Errorf("claims not valid")
}
